package provider

import (
	"context"

	"github.com/ner-studio/api/internal/model"
)

// VisionProvider analyzes reference images
type VisionProvider interface {
	AnalyzeImage(ctx context.Context, imageURL string) (*model.VisionAnalysisResult, error)
}

// LLMProvider generates text/prompts
type LLMProvider interface {
	GeneratePrompts(ctx context.Context, messages []LLMMessage, config LLMConfig) (*LLMResponse, error)
}

// ImageGenerationProvider generates images
type ImageGenerationProvider interface {
	GenerateImage(ctx context.Context, prompt string, config ImageGenConfig) (*ImageGenResult, error)
	ParseCallback(payload []byte) (*CallbackData, error)
}

// LLMMessage represents a message in LLM conversation
type LLMMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// LLMConfig for generation
type LLMConfig struct {
	Model       string
	Temperature float64
	MaxTokens   int
}

// LLMResponse from LLM
type LLMResponse struct {
	Content      string
	TokensUsed   int
	FinishReason string
}

// ImageGenConfig for image generation
type ImageGenConfig struct {
	Model        string
	Width        int
	Height       int
	CallbackURL  string
}

// ImageGenResult from image generation request
type ImageGenResult struct {
	TaskID string
	Status string
}

// CallbackData parsed from provider webhook
type CallbackData struct {
	TaskID      string
	Status      string // completed, failed
	ImageURL    string // temporary URL to download
	ErrorCode   string
	ErrorMessage string
}

// ProviderRegistry manages all providers
type ProviderRegistry interface {
	GetVisionProvider() (VisionProvider, error)
	GetLLMProviders() []LLMProviderWithPriority
	GetImageGenerationProvider(providerID string) (ImageGenerationProvider, error)
}

// LLMProviderWithPriority wraps an LLM provider with its priority
type LLMProviderWithPriority struct {
	Provider model.Provider
	Client   LLMProvider
}
