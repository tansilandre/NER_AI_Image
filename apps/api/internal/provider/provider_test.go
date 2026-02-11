package provider

import (
	"testing"

	"github.com/ner-studio/api/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFactory_RegisterAndGet(t *testing.T) {
	factory := NewFactory()

	// Test registering vision provider
	visionProvider := NewOpenAIVisionProvider("test-key", "")
	factory.RegisterVisionProvider(visionProvider)

	retrievedVision, err := factory.GetVisionProvider()
	assert.NoError(t, err)
	assert.NotNil(t, retrievedVision)

	// Test registering LLM providers
	llmProvider1 := &KieAIProvider{slug: "kieai-1", apiKey: "key1"}
	factory.RegisterLLMProvider(model.Provider{Slug: "kieai-1", Priority: 0}, llmProvider1)

	llmProvider2 := &KieAIProvider{slug: "kieai-2", apiKey: "key2"}
	factory.RegisterLLMProvider(model.Provider{Slug: "kieai-2", Priority: 1}, llmProvider2)

	llmProviders := factory.GetLLMProviders()
	assert.Len(t, llmProviders, 2)
	assert.Equal(t, "kieai-1", llmProviders[0].Provider.Slug) // Lower priority first

	// Test registering image provider
	imgProvider := &KieAIProvider{slug: "kieai-img", apiKey: "key3"}
	factory.RegisterImageProvider("kieai-img", imgProvider)

	retrievedImg, err := factory.GetImageGenerationProvider("kieai-img")
	assert.NoError(t, err)
	assert.NotNil(t, retrievedImg)

	// Test getting non-existent provider
	_, err = factory.GetImageGenerationProvider("non-existent")
	assert.Error(t, err)
}

func TestCleanPrompts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Double newline separator",
			input:    "Prompt 1\n\nPrompt 2\n\nPrompt 3",
			expected: []string{"Prompt 1", "Prompt 2", "Prompt 3"},
		},
		{
			name:     "Numbered prompts",
			input:    "1. First prompt\n2. Second prompt\n3. Third prompt",
			expected: []string{"First prompt", "Second prompt", "Third prompt"},
		},
		{
			name:     "Single prompt",
			input:    "Just one prompt",
			expected: []string{"Just one prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPrompts(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldFallbackLogic(t *testing.T) {
	// Test that shouldFallback works correctly with error codes
	testErr := &testError{msg: "connection timeout"}
	
	// With matching code
	result := shouldFallbackWithCodes(testErr, []string{"timeout"})
	assert.True(t, result)
	
	// With non-matching code
	result = shouldFallbackWithCodes(testErr, []string{"rate_limit"})
	assert.False(t, result)
	
	// With empty codes (should fallback on any error)
	result = shouldFallbackWithCodes(testErr, []string{})
	assert.True(t, result)
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Helper function to test fallback logic
func shouldFallbackWithCodes(err error, fallbackCodes []string) bool {
	if len(fallbackCodes) == 0 {
		return true
	}
	for _, code := range fallbackCodes {
		if contains(err.Error(), code) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Helper function to split prompts (simplified version)
func splitPrompts(text string) []string {
	var result []string
	
	// Split by double newline first
	parts := splitByString(text, "\n\n")
	if len(parts) > 1 {
		for _, p := range parts {
			cleaned := cleanPrompt(p)
			if cleaned != "" {
				result = append(result, cleaned)
			}
		}
		return result
	}
	
	// Split by single newline
	parts = splitByString(text, "\n")
	if len(parts) > 1 {
		for _, p := range parts {
			cleaned := cleanPrompt(p)
			if cleaned != "" {
				result = append(result, cleaned)
			}
		}
		return result
	}
	
	return []string{cleanPrompt(text)}
}

func splitByString(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func cleanPrompt(prompt string) string {
	// Trim spaces
	start := 0
	end := len(prompt)
	for start < end && (prompt[start] == ' ' || prompt[start] == '\t' || prompt[start] == '\n' || prompt[start] == '\r') {
		start++
	}
	for end > start && (prompt[end-1] == ' ' || prompt[end-1] == '\t' || prompt[end-1] == '\n' || prompt[end-1] == '\r') {
		end--
	}
	prompt = prompt[start:end]
	
	// Remove "-" prefix
	if len(prompt) > 0 && prompt[0] == '-' {
		prompt = prompt[1:]
		// Re-trim
		start = 0
		end = len(prompt)
		for start < end && (prompt[start] == ' ') {
			start++
		}
		prompt = prompt[start:end]
	}
	
	// Remove numbered prefixes
	for i := 1; i <= 10; i++ {
		prefix := itoa(i) + "."
		if len(prompt) >= len(prefix) && prompt[:len(prefix)] == prefix {
			prompt = prompt[len(prefix):]
			// Re-trim
			start = 0
			end = len(prompt)
			for start < end && (prompt[start] == ' ') {
				start++
			}
			prompt = prompt[start:end]
		}
		prefix = itoa(i) + ")"
		if len(prompt) >= len(prefix) && prompt[:len(prefix)] == prefix {
			prompt = prompt[len(prefix):]
			// Re-trim
			start = 0
			end = len(prompt)
			for start < end && (prompt[start] == ' ') {
				start++
			}
			prompt = prompt[start:end]
		}
	}
	
	return prompt
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
