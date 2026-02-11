package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys
type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	OrganizationIDKey contextKey = "organization_id"
	RoleKey           contextKey = "role"
	JWTSecretKey      contextKey = "jwt_secret"
)

// AuthConfig holds JWT validation settings
type AuthConfig struct {
	JWTSecret string
}

// NewAuthMiddleware creates JWT auth middleware
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
			return []byte(config.JWTSecret), nil
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
		if orgID, ok := claims["org_id"].(string); ok {
			c.Locals(string(OrganizationIDKey), orgID)
		}
		if role, ok := claims["role"].(string); ok {
			c.Locals(string(RoleKey), role)
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

// GetOrganizationID extracts org ID from context
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

// GenerateJWT creates a new JWT token for a user
func GenerateJWT(userID, orgID, role string, secret string, expiryHours int) (string, error) {
	claims := jwt.MapClaims{
		"sub":    userID,
		"org_id": orgID,
		"role":   role,
		"exp":    time.Now().Add(time.Hour * time.Duration(expiryHours)).Unix(),
		"iat":    time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
