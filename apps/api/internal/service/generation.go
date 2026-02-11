package service

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
	"github.com/ner-studio/api/internal/provider"
	"github.com/ner-studio/api/internal/repository"
)

// GenerationService handles image generation workflow
type GenerationService struct {
	repo            *repository.Repository
	factory         *provider.Factory
	callbackBaseURL string
}

// NewGenerationService creates a new generation service
func NewGenerationService(repo *repository.Repository, factory *provider.Factory, callbackBaseURL string) *GenerationService {
	return &GenerationService{
		repo:            repo,
		factory:         factory,
		callbackBaseURL: callbackBaseURL,
	}
}

// CreateGenerationRequest holds parameters for creating a generation
type CreateGenerationRequest struct {
	UserID          string
	OrganizationID  string
	BasePrompt      string
	ReferenceImages []string
	ProductImages   []string
	ProviderID      string
	NumVariations   int
}

// CreateGeneration starts the image generation workflow
func (s *GenerationService) CreateGeneration(ctx context.Context, req CreateGenerationRequest) (*model.Generation, error) {
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, fmt.Errorf("invalid organization ID: %w", err)
	}
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}
	providerID, err := uuid.Parse(req.ProviderID)
	if err != nil {
		return nil, fmt.Errorf("invalid provider ID: %w", err)
	}

	// Get provider to calculate cost
	prov, err := s.repo.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Set default variations
	numVariations := req.NumVariations
	if numVariations < 1 {
		numVariations = 4
	}
	if numVariations > 10 {
		numVariations = 10
	}

	// Calculate estimated cost
	estimatedCost := prov.CostPerUse * int64(numVariations)

	// Check credits
	org, err := s.repo.GetOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org.Credits < estimatedCost {
		return nil, fmt.Errorf("insufficient credits: have %d, need %d", org.Credits, estimatedCost)
	}

	// Create generation record
	gen := &model.Generation{
		ID:              uuid.New(),
		OrganizationID:  orgID,
		UserID:          userID,
		Status:          "pending",
		BasePrompt:      req.BasePrompt,
		ReferenceImages: req.ReferenceImages,
		ProductImages:   req.ProductImages,
		ProviderID:      providerID,
		EstimatedCost:   estimatedCost,
	}

	// Save to database
	if err := s.repo.CreateGeneration(ctx, gen); err != nil {
		return nil, fmt.Errorf("failed to create generation: %w", err)
	}

	// Start async workflow
	go func() {
		// Create background context
		bgCtx := context.Background()
		if err := s.runGenerationWorkflow(bgCtx, gen.ID); err != nil {
			log.Printf("Generation workflow failed for %s: %v", gen.ID, err)
		}
	}()

	return gen, nil
}

// runGenerationWorkflow executes the full generation pipeline
func (s *GenerationService) runGenerationWorkflow(ctx context.Context, genID uuid.UUID) error {
	// Update status to processing
	if err := s.repo.UpdateGenerationStatus(ctx, genID, "processing", ""); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Get generation details
	gen, err := s.repo.GetGeneration(ctx, genID)
	if err != nil {
		return fmt.Errorf("failed to get generation: %w", err)
	}

	// Step 1: Analyze reference images (if any)
	var visionResults []*model.VisionAnalysisResult
	if len(gen.ReferenceImages) > 0 {
		visionResults, err = s.analyzeReferenceImages(ctx, gen.ReferenceImages)
		if err != nil {
			s.repo.UpdateGenerationStatus(ctx, genID, "failed", err.Error())
			return fmt.Errorf("vision analysis failed: %w", err)
		}
	}

	// Step 2: Build LLM messages
	messages := s.buildLLMMessages(gen.BasePrompt, visionResults)

	// Step 3: Generate prompts with fallback
	llmResp, err := s.generatePromptsWithFallback(ctx, messages)
	if err != nil {
		s.repo.UpdateGenerationStatus(ctx, genID, "failed", err.Error())
		return fmt.Errorf("LLM generation failed: %w", err)
	}

	// Step 4: Split prompts into variations
	prompts := s.splitPrompts(llmResp.Content)
	if len(prompts) == 0 {
		s.repo.UpdateGenerationStatus(ctx, genID, "failed", "no prompts generated")
		return fmt.Errorf("no prompts generated")
	}

	// Step 5: Create generation image records
	images := make([]*model.GenerationImage, 0, len(prompts))
	for _, prompt := range prompts {
		img := &model.GenerationImage{
			ID:           uuid.New(),
			GenerationID: genID,
			Prompt:       prompt,
			Status:       "pending",
		}
		images = append(images, img)
	}

	// Save images to database
	for _, img := range images {
		if err := s.repo.CreateGenerationImage(ctx, img); err != nil {
			return fmt.Errorf("failed to create image record: %w", err)
		}
	}

	// Step 6: Submit image generation jobs
	providerSlug := "kieai-seedream" // Default, should come from provider config
	callbackURL := fmt.Sprintf("%s/api/v1/callbacks/%s", s.callbackBaseURL, providerSlug)

	imgProvider, err := s.factory.GetImageGenerationProvider(providerSlug)
	if err != nil {
		s.repo.UpdateGenerationStatus(ctx, genID, "failed", err.Error())
		return fmt.Errorf("failed to get image provider: %w", err)
	}

	for _, img := range images {
		result, err := imgProvider.GenerateImage(ctx, img.Prompt, provider.ImageGenConfig{
			Model:       "seedream-v1",
			Width:       1024,
			Height:      1024,
			CallbackURL: callbackURL,
		})
		if err != nil {
			log.Printf("Failed to submit image job for %s: %v", img.ID, err)
			s.repo.UpdateGenerationImageFailed(ctx, img.ID, err.Error())
			continue
		}

		// Update with task ID
		img.TaskID = result.TaskID
		img.Status = "processing"
	}

	return nil
}

// analyzeReferenceImages analyzes uploaded reference images
func (s *GenerationService) analyzeReferenceImages(ctx context.Context, imageURLs []string) ([]*model.VisionAnalysisResult, error) {
	visionProvider, err := s.factory.GetVisionProvider()
	if err != nil {
		return nil, err
	}

	var results []*model.VisionAnalysisResult
	for _, url := range imageURLs {
		result, err := visionProvider.AnalyzeImage(ctx, url)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze image %s: %w", url, err)
		}
		results = append(results, result)
	}

	return results, nil
}

// buildLLMMessages builds the prompt for LLM
func (s *GenerationService) buildLLMMessages(basePrompt string, visionResults []*model.VisionAnalysisResult) []provider.LLMMessage {
	var systemPrompt strings.Builder
	systemPrompt.WriteString("You are a creative prompt engineer for AI image generation. ")
	systemPrompt.WriteString("Given a base prompt and optional reference image analysis, ")
	systemPrompt.WriteString("generate 4-6 detailed, creative variations of prompts. ")
	systemPrompt.WriteString("Each prompt should be unique and optimized for image generation.\n\n")
	systemPrompt.WriteString("Format: Separate each prompt with a blank line (double newline).")

	// Add vision analysis context
	if len(visionResults) > 0 {
		systemPrompt.WriteString("\n\nReference Image Analysis:\n")
		for i, result := range visionResults {
			systemPrompt.WriteString(fmt.Sprintf("Image %d: %s\nStyle Notes: %s\n",
				i+1, result.Description, result.StyleNotes))
		}
	}

	userPrompt := fmt.Sprintf("Base Prompt: %s\n\nGenerate 4-6 creative variations:", basePrompt)

	return []provider.LLMMessage{
		{Role: "system", Content: systemPrompt.String()},
		{Role: "user", Content: userPrompt},
	}
}

// generatePromptsWithFallback tries LLM providers in order
func (s *GenerationService) generatePromptsWithFallback(ctx context.Context, messages []provider.LLMMessage) (*provider.LLMResponse, error) {
	llmProviders := s.factory.GetLLMProviders()

	if len(llmProviders) == 0 {
		return nil, fmt.Errorf("no LLM providers available")
	}

	for _, p := range llmProviders {
		resp, err := p.Client.GeneratePrompts(ctx, messages, provider.LLMConfig{
			Model:       p.Provider.Model,
			Temperature: 0.8,
			MaxTokens:   2000,
		})

		if err != nil {
			// Check if this error should trigger fallback
			if shouldFallback(err, p.Provider.Config.ErrorCodeForFallback) {
				log.Printf("LLM provider %s failed, trying next: %v", p.Provider.Slug, err)
				continue
			}
			return nil, err
		}

		return resp, nil
	}

	return nil, fmt.Errorf("all LLM providers failed")
}

// shouldFallback checks if error should trigger provider fallback
func shouldFallback(err error, fallbackCodes []string) bool {
	if len(fallbackCodes) == 0 {
		return true // Default: fallback on any error
	}

	errStr := strings.ToLower(err.Error())
	for _, code := range fallbackCodes {
		if strings.Contains(errStr, strings.ToLower(code)) {
			return true
		}
	}
	return false
}

// splitPrompts splits generated text into individual prompts
func (s *GenerationService) splitPrompts(text string) []string {
	// Try double newline first
	prompts := strings.Split(text, "\n\n")
	if len(prompts) > 1 {
		return cleanPrompts(prompts)
	}

	// Try single newline
	prompts = strings.Split(text, "\n")
	if len(prompts) > 1 {
		return cleanPrompts(prompts)
	}

	// Try semicolon
	prompts = strings.Split(text, ";")
	if len(prompts) > 1 {
		return cleanPrompts(prompts)
	}

	// Return as single prompt
	return []string{cleanPrompt(text)}
}

func cleanPrompts(prompts []string) []string {
	var result []string
	for _, p := range prompts {
		cleaned := cleanPrompt(p)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}
	return result
}

func cleanPrompt(prompt string) string {
	prompt = strings.TrimSpace(prompt)
	// Remove numbered prefixes like "1. " or "1) "
	prompt = strings.TrimPrefix(prompt, "-")
	prompt = strings.TrimSpace(prompt)

	// Remove common prefixes
	for i := 1; i <= 10; i++ {
		prefix := fmt.Sprintf("%d.", i)
		prompt = strings.TrimPrefix(prompt, prefix)
		prefix = fmt.Sprintf("%d)", i)
		prompt = strings.TrimPrefix(prompt, prefix)
		prompt = strings.TrimSpace(prompt)
	}

	return prompt
}

// HandleCallback processes provider callbacks
func (s *GenerationService) HandleCallback(ctx context.Context, providerSlug string, payload []byte) error {
	// Get provider to parse callback
	imgProvider, err := s.factory.GetImageGenerationProvider(providerSlug)
	if err != nil {
		return fmt.Errorf("provider not found: %w", err)
	}

	// Parse callback data
	callbackData, err := imgProvider.ParseCallback(payload)
	if err != nil {
		return fmt.Errorf("failed to parse callback: %w", err)
	}

	// Find image by task ID
	img, err := s.repo.GetGenerationImageByTaskID(ctx, callbackData.TaskID)
	if err != nil {
		return fmt.Errorf("image not found for task %s: %w", callbackData.TaskID, err)
	}

	// Handle success/failure
	if callbackData.Status == "completed" {
		// Download image and upload to R2
		// TODO: Implement actual download and re-upload
		r2Key := fmt.Sprintf("generations/%s/%s.jpg", img.GenerationID, img.ID)
		r2URL := fmt.Sprintf("https://bucket.tansil.pro/%s", r2Key)

		if err := s.repo.UpdateGenerationImageComplete(ctx, img.ID, r2URL, r2Key); err != nil {
			return fmt.Errorf("failed to update image: %w", err)
		}

		// Check if all images are done
		if err := s.checkGenerationComplete(ctx, img.GenerationID); err != nil {
			log.Printf("Failed to check generation status: %v", err)
		}
	} else {
		// Failed
		if err := s.repo.UpdateGenerationImageFailed(ctx, img.ID, callbackData.ErrorMessage); err != nil {
			return fmt.Errorf("failed to update image: %w", err)
		}
	}

	return nil
}

// checkGenerationComplete checks if all images are done and updates generation status
func (s *GenerationService) checkGenerationComplete(ctx context.Context, generationID uuid.UUID) error {
	total, completed, failed, err := s.repo.GetGenerationStats(ctx, generationID)
	if err != nil {
		return err
	}

	if total == 0 {
		return nil
	}

	if completed+failed == total {
		// All done
		status := "completed"
		if failed == total {
			status = "failed"
		}

		if err := s.repo.UpdateGenerationStatus(ctx, generationID, status, ""); err != nil {
			return err
		}

		// Deduct actual credits
		gen, err := s.repo.GetGeneration(ctx, generationID)
		if err != nil {
			return err
		}

		actualCost := gen.EstimatedCost / int64(total) * int64(completed)
		s.repo.UpdateGenerationActualCost(ctx, generationID, actualCost)

		// Deduct credits
		description := fmt.Sprintf("Image generation %s (%d/%d completed)", generationID, completed, total)
		if err := s.repo.DeductCredits(ctx, gen.OrganizationID, actualCost, description, gen.UserID, &generationID); err != nil {
			log.Printf("Failed to deduct credits: %v", err)
		}
	}

	return nil
}

// GetGeneration retrieves a generation with its images
func (s *GenerationService) GetGeneration(ctx context.Context, id uuid.UUID) (*model.Generation, []*model.GenerationImage, error) {
	gen, err := s.repo.GetGeneration(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	images, err := s.repo.ListGenerationImages(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	return gen, images, nil
}

// ListGenerations lists generations for an organization
func (s *GenerationService) ListGenerations(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*model.Generation, error) {
	return s.repo.ListGenerations(ctx, orgID, limit, offset)
}
