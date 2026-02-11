package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ner-studio/api/internal/model"
)

// GeminiProvider implements LLMProvider using Google Gemini API
type GeminiProvider struct {
	slug    string
	apiKey  string
	baseURL string
	model   string
	config  model.ProviderConfig
	client  *http.Client
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(slug, apiKey, baseURL, model string, config model.ProviderConfig) *GeminiProvider {
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	timeoutMs := config.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 60000
	}
	return &GeminiProvider{
		slug:    slug,
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		config:  config,
		client:  &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

// GeneratePrompts generates text using Gemini API
func (p *GeminiProvider) GeneratePrompts(ctx context.Context, messages []LLMMessage, cfg LLMConfig) (*LLMResponse, error) {
	// Convert messages to Gemini format
	var contents []map[string]interface{}
	for _, msg := range messages {
		role := msg.Role
		if role == "system" {
			// Gemini uses a different format for system instructions
			continue
		}
		if role == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": msg.Content},
			},
		})
	}

	// Extract system message
	var systemInstruction string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemInstruction = msg.Content
			break
		}
	}

	reqBody := map[string]interface{}{
		"contents": contents,
		"generationConfig": map[string]interface{}{
			"temperature": cfg.Temperature,
			"maxOutputTokens": cfg.MaxTokens,
		},
	}

	if systemInstruction != "" {
		reqBody["systemInstruction"] = map[string]interface{}{
			"parts": []map[string]string{
				{"text": systemInstruction},
			},
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)
	
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		UsageMetadata struct {
			TotalTokenCount int `json:"totalTokenCount"`
		} `json:"usageMetadata"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if len(result.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in Gemini response")
	}

	content := ""
	for _, part := range result.Candidates[0].Content.Parts {
		content += part.Text
	}

	return &LLMResponse{
		Content:      content,
		TokensUsed:   result.UsageMetadata.TotalTokenCount,
		FinishReason: result.Candidates[0].FinishReason,
	}, nil
}
