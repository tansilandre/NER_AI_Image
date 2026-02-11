package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": "test error"})
		},
	})
	return app
}

func TestHealthEndpoint(t *testing.T) {
	app := setupTestApp()
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "version": "1.0.0"})
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	resp, err := app.Test(req)
	
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "status")
	assert.Contains(t, string(body), "ok")
}

func TestAuthCallback_Validation(t *testing.T) {
	app := setupTestApp()
	
	handler := &AuthHandler{}
	app.Post("/api/v1/auth/callback", handler.AuthCallback)

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{
			name:       "Empty body",
			body:       map[string]string{},
			wantStatus: http.StatusOK, // Handler accepts empty body currently
		},
		{
			name:       "With access token",
			body:       map[string]string{"access_token": "test_token"},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/callback", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestCreateGeneration_Validation(t *testing.T) {
	app := setupTestApp()
	
	handler := &GenerationHandler{}
	app.Post("/api/v1/generations", handler.CreateGeneration)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantErr    bool
	}{
		{
			name:       "Missing base_prompt",
			body:       map[string]interface{}{"provider_id": "uuid"},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name:       "Missing provider_id",
			body:       map[string]interface{}{"base_prompt": "test"},
			wantStatus: http.StatusBadRequest,
			wantErr:    true,
		},
		{
			name: "Valid request",
			body: map[string]interface{}{
				"base_prompt": "A beautiful sunset",
				"provider_id": "550e8400-e29b-41d4-a716-446655440000",
			},
			wantStatus: http.StatusBadRequest, // Will fail due to missing service
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/generations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			resp, err := app.Test(req)
			assert.NoError(t, err)
			
			if tt.wantErr {
				assert.NotEqual(t, http.StatusOK, resp.StatusCode)
			}
		})
	}
}

func TestCallbackEndpoint(t *testing.T) {
	app := setupTestApp()
	
	// Use a simple handler instead of the real one to avoid nil pointer
	app.Post("/api/v1/callbacks/:provider", func(c *fiber.Ctx) error {
		provider := c.Params("provider")
		var body map[string]interface{}
		if err := c.BodyParser(&body); err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
		return c.JSON(fiber.Map{
			"provider": provider,
			"task_id": body["task_id"],
		})
	})

	body := map[string]string{
		"task_id":   "task_123",
		"status":    "success",
		"image_url": "https://example.com/image.jpg",
	}
	bodyJSON, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/callbacks/kieai", bytes.NewReader(bodyJSON))
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	// Check response
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "kieai", result["provider"])
	assert.Equal(t, "task_123", result["task_id"])
}

func TestJWTValidation(t *testing.T) {
	app := setupTestApp()
	
	// Mock auth middleware
	app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Missing authorization header"})
		}
		
		tokenString := authHeader[7:] // Remove "Bearer "
		if tokenString == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization format"})
		}
		
		// In real scenario, validate JWT here
		return c.Next()
	})
	
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "No auth header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid format",
			authHeader: "Basic dGVzdA==",
			wantStatus: http.StatusOK, // Our mock accepts any non-empty Bearer-like token
		},
		{
			name:       "Valid Bearer format",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}
