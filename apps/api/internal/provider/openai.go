package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ner-studio/api/internal/model"
)

// OpenAIVisionProvider implements VisionProvider using OpenAI GPT-4o
type OpenAIVisionProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIVisionProvider creates a new OpenAI vision provider
func NewOpenAIVisionProvider(apiKey, baseURL string) *OpenAIVisionProvider {
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	return &OpenAIVisionProvider{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30},
	}
}

// AnalyzeImage analyzes a reference image using GPT-4o Vision
func (p *OpenAIVisionProvider) AnalyzeImage(ctx context.Context, imageURL string) (*model.VisionAnalysisResult, error) {
	systemPrompt := `You are a professional creative director analyzing reference images for an AI image generation platform.

Analyze the provided image and describe:
1. Overall visual style (artistic style, mood, atmosphere)
2. Color palette and lighting
3. Composition and framing
4. Key visual elements that should be preserved

Format your response as JSON with these fields:
{
  "description": "detailed description of what's in the image",
  "style_notes": "key style characteristics to apply to generated images"
}`

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]interface{}{
			{
				"role":    "system",
				"content": systemPrompt,
			},
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": "Analyze this reference image for style direction:",
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": imageURL,
						},
					},
				},
			},
		},
		"max_tokens": 1000,
		"response_format": map[string]string{"type": "json_object"},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse JSON content
	content := result.Choices[0].Message.Content
	var analysis model.VisionAnalysisResult
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		// Fallback: use content as description
		analysis.Description = content
		analysis.StyleNotes = "Style derived from reference image"
	}

	return &analysis, nil
}
