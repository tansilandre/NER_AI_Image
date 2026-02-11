package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys
type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	OrganizationIDKey contextKey = "organization_id"
	RoleKey           contextKey = "role"
)

// AuthConfig holds JWT validation settings
type AuthConfig struct {
	SupabaseURL    string
	SupabaseAnonKey string
}

// NewAuthMiddleware creates JWT auth middleware using Supabase tokens
func NewAuthMiddleware(config AuthConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip auth for public routes
		if isPublicRoute(c.Path()) {
			return c.Next()
		}

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization header",
			})
		}

		// Extract Bearer token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid authorization format",
			})
		}

		// Parse and validate JWT
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// For Supabase JWT, we use the anon key as the secret
			// In production, you should fetch the JWT secret from Supabase
			return []byte(config.SupabaseAnonKey), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Set user info in context
		if sub, ok := claims["sub"].(string); ok {
			c.Locals(string(UserIDKey), sub)
		}

		return c.Next()
	}
}

// isPublicRoute checks if the route should skip auth
func isPublicRoute(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/v1/auth/",
		"/api/v1/callbacks/",
	}
	for _, pp := range publicPaths {
		if strings.HasPrefix(path, pp) {
			return true
		}
	}
	return false
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) string {
	userID, ok := c.Locals(string(UserIDKey)).(string)
	if !ok {
		return ""
	}
	return userID
}

// GetOrganizationID extracts org ID from context (set by org middleware)
func GetOrganizationID(c *fiber.Ctx) string {
	orgID, ok := c.Locals(string(OrganizationIDKey)).(string)
	if !ok {
		return ""
	}
	return orgID
}

// GetRole extracts user role from context
func GetRole(c *fiber.Ctx) string {
	role, ok := c.Locals(string(RoleKey)).(string)
	if !ok {
		return ""
	}
	return role
}

// RequireAdmin ensures the user is an admin
func RequireAdmin() fiber.Handler {
	return func(c *fiber.Ctx) error {
		role := GetRole(c)
		if role != "admin" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Admin access required",
			})
		}
		return c.Next()
	}
}
