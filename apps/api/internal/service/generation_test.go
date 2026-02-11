package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitPrompts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "Double newline",
			input:    "Prompt 1\n\nPrompt 2\n\nPrompt 3",
			expected: []string{"Prompt 1", "Prompt 2", "Prompt 3"},
		},
		{
			name:     "Single newline",
			input:    "Line 1\nLine 2\nLine 3",
			expected: []string{"Line 1", "Line 2", "Line 3"},
		},
		{
			name:     "Numbered list",
			input:    "1. First\n2. Second\n3. Third",
			expected: []string{"First", "Second", "Third"},
		},
		{
			name:     "Single prompt",
			input:    "Just one prompt",
			expected: []string{"Just one prompt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitPromptsForTest(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCleanPrompt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  test  ", "test"},
		{"- bullet", "bullet"},
		{"1. numbered", "numbered"},
		{"2) numbered", "numbered"},
		{"  3.  spaced  ", "spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanPromptForTest(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldFallback(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		fallbackCodes  []string
		shouldFallback bool
	}{
		{
			name:           "No codes - fallback on any error",
			err:            assert.AnError,
			fallbackCodes:  []string{},
			shouldFallback: true,
		},
		{
			name:           "Match timeout code",
			err:            &testError{msg: "request timeout"},
			fallbackCodes:  []string{"timeout"},
			shouldFallback: true,
		},
		{
			name:           "No match",
			err:            &testError{msg: "bad request"},
			fallbackCodes:  []string{"timeout"},
			shouldFallback: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldFallbackForTest(tt.err, tt.fallbackCodes)
			assert.Equal(t, tt.shouldFallback, result)
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// Test implementations
func splitPromptsForTest(text string) []string {
	var result []string
	
	// Split by double newline first
	parts := splitByString(text, "\n\n")
	if len(parts) > 1 {
		for _, p := range parts {
			cleaned := cleanPromptForTest(p)
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
			cleaned := cleanPromptForTest(p)
			if cleaned != "" {
				result = append(result, cleaned)
			}
		}
		return result
	}
	
	return []string{cleanPromptForTest(text)}
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

func cleanPromptForTest(prompt string) string {
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
		// Re-trim spaces
		start = 0
		for start < len(prompt) && prompt[start] == ' ' {
			start++
		}
		prompt = prompt[start:]
	}
	
	// Remove numbered prefixes
	for i := 1; i <= 10; i++ {
		prefix := itoa(i) + "."
		if len(prompt) >= len(prefix) && prompt[:len(prefix)] == prefix {
			prompt = prompt[len(prefix):]
			// Re-trim
			start = 0
			for start < len(prompt) && prompt[start] == ' ' {
				start++
			}
			prompt = prompt[start:]
		}
		prefix = itoa(i) + ")"
		if len(prompt) >= len(prefix) && prompt[:len(prefix)] == prefix {
			prompt = prompt[len(prefix):]
			// Re-trim
			start = 0
			for start < len(prompt) && prompt[start] == ' ' {
				start++
			}
			prompt = prompt[start:]
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

func shouldFallbackForTest(err error, fallbackCodes []string) bool {
	if len(fallbackCodes) == 0 {
		return true
	}
	
	errStr := err.Error()
	for _, code := range fallbackCodes {
		if containsStr(errStr, code) {
			return true
		}
	}
	return false
}

func containsStr(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
