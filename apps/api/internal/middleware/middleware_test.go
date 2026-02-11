package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware_PublicRoutes(t *testing.T) {
	app := fiber.New()
	
	// Setup auth middleware with test config
	authConfig := AuthConfig{
		SupabaseURL:     "http://localhost:54321",
		SupabaseAnonKey: "test-secret-key-for-jwt-validation-in-tests",
	}
	app.Use(NewAuthMiddleware(authConfig))
	
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	app.Get("/api/v1/callbacks/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "callback"})
	})
	app.Get("/api/v1/auth/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "auth"})
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{
			name:       "Health endpoint - public",
			path:       "/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Callback endpoint - public",
			path:       "/api/v1/callbacks/test",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Auth endpoint - public",
			path:       "/api/v1/auth/test",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			resp, err := app.Test(req)
			
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestAuthMiddleware_ProtectedRoutes(t *testing.T) {
	app := fiber.New()
	
	secret := []byte("test-secret-key-for-jwt-validation-in-tests")
	authConfig := AuthConfig{
		SupabaseURL:     "http://localhost:54321",
		SupabaseAnonKey: string(secret),
	}
	app.Use(NewAuthMiddleware(authConfig))
	
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"user_id": GetUserID(c)})
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
			name:       "Invalid format - no Bearer",
			authHeader: "Basic dGVzdA==",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid.token.here",
			wantStatus: http.StatusUnauthorized,
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

func TestCORSMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(NewCORS("http://localhost:5173,http://localhost:3000"))
	
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "ok"})
	})

	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	req.Header.Set("Access-Control-Request-Method", "POST")
	
	resp, err := app.Test(req)
	assert.NoError(t, err)
	
	// Should return 204 for preflight
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	
	// Check CORS headers
	allowOrigin := resp.Header.Get("Access-Control-Allow-Origin")
	assert.NotEmpty(t, allowOrigin)
}

func TestRequireAdmin(t *testing.T) {
	app := fiber.New()
	
	// Setup middleware to set role
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(string(RoleKey), c.Get("X-Test-Role"))
		return c.Next()
	})
	
	app.Use(RequireAdmin())
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "admin access granted"})
	})

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{
			name:       "Admin role",
			role:       "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "Member role",
			role:       "member",
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "No role",
			role:       "",
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			req.Header.Set("X-Test-Role", tt.role)
			
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name     string
		locals   interface{}
		expected string
	}{
		{
			name:     "Valid user ID",
			locals:   "user-123",
			expected: "user-123",
		},
		{
			name:     "No user ID",
			locals:   nil,
			expected: "",
		},
		{
			name:     "Wrong type",
			locals:   123,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c *fiber.Ctx) error {
				if tt.locals != nil {
					c.Locals(string(UserIDKey), tt.locals)
				}
				userID := GetUserID(c)
				return c.JSON(fiber.Map{"user_id": userID})
			})
			
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp, err := app.Test(req)
			
			assert.NoError(t, err)
			
			body, _ := io.ReadAll(resp.Body)
			assert.Contains(t, string(body), tt.expected)
		})
	}
}
