package provider

import (
	"fmt"
	"sort"

	"github.com/ner-studio/api/internal/model"
)

// Factory creates provider instances from configuration
type Factory struct {
	visionProvider  VisionProvider
	llmProviders    []LLMProviderWithPriority
	imageProviders  map[string]ImageGenerationProvider
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		imageProviders: make(map[string]ImageGenerationProvider),
	}
}

// RegisterVisionProvider registers a vision provider
func (f *Factory) RegisterVisionProvider(provider VisionProvider) {
	f.visionProvider = provider
}

// RegisterLLMProvider registers an LLM provider
func (f *Factory) RegisterLLMProvider(p model.Provider, client LLMProvider) {
	f.llmProviders = append(f.llmProviders, LLMProviderWithPriority{
		Provider: p,
		Client:   client,
	})
	// Sort by priority
	sort.Slice(f.llmProviders, func(i, j int) bool {
		return f.llmProviders[i].Provider.Priority < f.llmProviders[j].Provider.Priority
	})
}

// RegisterImageProvider registers an image generation provider
func (f *Factory) RegisterImageProvider(providerID string, provider ImageGenerationProvider) {
	f.imageProviders[providerID] = provider
}

// GetVisionProvider returns the registered vision provider
func (f *Factory) GetVisionProvider() (VisionProvider, error) {
	if f.visionProvider == nil {
		return nil, fmt.Errorf("no vision provider registered")
	}
	return f.visionProvider, nil
}

// GetLLMProviders returns all registered LLM providers sorted by priority
func (f *Factory) GetLLMProviders() []LLMProviderWithPriority {
	return f.llmProviders
}

// GetImageGenerationProvider returns an image generation provider by ID
func (f *Factory) GetImageGenerationProvider(providerID string) (ImageGenerationProvider, error) {
	provider, ok := f.imageProviders[providerID]
	if !ok {
		return nil, fmt.Errorf("image provider not found: %s", providerID)
	}
	return provider, nil
}

// CreateProviderFromConfig creates a provider instance from database config
func CreateProviderFromConfig(p model.Provider) (interface{}, error) {
	switch p.Category {
	case "vision":
		if p.Slug == "openai-gpt4o" {
			return NewOpenAIVisionProvider(p.APIKey, p.BaseURL), nil
		}
		return nil, fmt.Errorf("unknown vision provider: %s", p.Slug)

	case "llm":
		if p.Slug == "google-gemini" {
			return NewGeminiProvider(p.Slug, p.APIKey, p.BaseURL, p.Model, p.Config), nil
		}
		// kie.ai LLM providers
		if p.Slug == "kieai-gemini3" || p.Slug == "kieai-gemini25" {
			return NewKieAIProvider(p.Slug, p.APIKey, p.BaseURL, p.Model, p.Config), nil
		}
		return nil, fmt.Errorf("unknown LLM provider: %s", p.Slug)

	case "image_generation":
		// kie.ai image generation
		if p.Slug == "kieai-seedream" || p.Slug == "kieai-nano" {
			return NewKieAIProvider(p.Slug, p.APIKey, p.BaseURL, p.Model, p.Config), nil
		}
		return nil, fmt.Errorf("unknown image generation provider: %s", p.Slug)

	default:
		return nil, fmt.Errorf("unknown provider category: %s", p.Category)
	}
}
