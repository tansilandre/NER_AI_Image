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

// KieAIProvider implements both LLM and ImageGeneration providers
type KieAIProvider struct {
	slug    string
	apiKey  string
	baseURL string
	model   string
	config  model.ProviderConfig
	client  *http.Client
}

// NewKieAIProvider creates a new KieAI provider
func NewKieAIProvider(slug, apiKey, baseURL, model string, config model.ProviderConfig) *KieAIProvider {
	if baseURL == "" {
		baseURL = "https://api.kie.ai"
	}
	timeoutMs := config.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 60000
	}
	return &KieAIProvider{
		slug:    slug,
		apiKey:  apiKey,
		baseURL: baseURL,
		model:   model,
		config:  config,
		client:  &http.Client{Timeout: time.Duration(timeoutMs) * time.Millisecond},
	}
}

// --- LLM Provider Implementation ---

// GeneratePrompts generates text using KieAI's LLM proxy
func (p *KieAIProvider) GeneratePrompts(ctx context.Context, messages []LLMMessage, cfg LLMConfig) (*LLMResponse, error) {
	reqBody := map[string]interface{}{
		"model": cfg.Model,
		"messages": messages,
		"temperature": cfg.Temperature,
		"max_tokens": cfg.MaxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/chat/completions", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kie.ai LLM error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kie.ai LLM API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in LLM response")
	}

	return &LLMResponse{
		Content:      result.Choices[0].Message.Content,
		TokensUsed:   result.Usage.TotalTokens,
		FinishReason: result.Choices[0].FinishReason,
	}, nil
}

// --- Image Generation Provider Implementation ---

// GenerateImage submits an image generation job
func (p *KieAIProvider) GenerateImage(ctx context.Context, prompt string, cfg ImageGenConfig) (*ImageGenResult, error) {
	reqBody := map[string]interface{}{
		"model":        cfg.Model,
		"prompt":       prompt,
		"width":        cfg.Width,
		"height":       cfg.Height,
		"callback_url": cfg.CallbackURL,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/v1/images/generations", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("kie.ai image gen error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("kie.ai image gen API error %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		TaskID string `json:"task_id"`
		Status string `json:"status"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse image gen response: %w", err)
	}

	return &ImageGenResult{
		TaskID: result.TaskID,
		Status: result.Status,
	}, nil
}

// ParseCallback parses the callback payload from KieAI
func (p *KieAIProvider) ParseCallback(payload []byte) (*CallbackData, error) {
	var result struct {
		TaskID       string `json:"task_id"`
		Status       string `json:"status"` // success, failed
		ImageURL     string `json:"image_url,omitempty"`
		ErrorCode    string `json:"error_code,omitempty"`
		ErrorMessage string `json:"error_message,omitempty"`
	}

	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, fmt.Errorf("failed to parse callback: %w", err)
	}

	status := result.Status
	if status == "success" {
		status = "completed"
	}

	return &CallbackData{
		TaskID:       result.TaskID,
		Status:       status,
		ImageURL:     result.ImageURL,
		ErrorCode:    result.ErrorCode,
		ErrorMessage: result.ErrorMessage,
	}, nil
}
