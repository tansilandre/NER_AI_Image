package test

import (
	"context"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/model"
)

// TestConfig holds test configuration
type TestConfig struct {
	DatabaseURL string
	APIKey      string
}

// MockRepository provides mock repository methods
type MockRepository struct {
	Organizations map[uuid.UUID]*model.Organization
	Profiles      map[uuid.UUID]*model.Profile
	Generations   map[uuid.UUID]*model.Generation
	Images        map[uuid.UUID]*model.GenerationImage
	Providers     map[uuid.UUID]*model.Provider
}

// NewMockRepository creates a new mock repository
func NewMockRepository() *MockRepository {
	return &MockRepository{
		Organizations: make(map[uuid.UUID]*model.Organization),
		Profiles:      make(map[uuid.UUID]*model.Profile),
		Generations:   make(map[uuid.UUID]*model.Generation),
		Images:        make(map[uuid.UUID]*model.GenerationImage),
		Providers:     make(map[uuid.UUID]*model.Provider),
	}
}

// GenerateTestToken creates a test JWT token
func GenerateTestToken(userID string, secret string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

// TestUser creates a test user
func TestUser() (userID uuid.UUID, orgID uuid.UUID, profile *model.Profile) {
	userID = uuid.New()
	orgID = uuid.New()
	
	profile = &model.Profile{
		ID:             uuid.New(),
		UserID:         userID,
		OrganizationID: orgID,
		FullName:       "Test User",
		Role:           "admin",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	return
}

// SetupTestApp creates a test Fiber app
func SetupTestApp() *fiber.App {
	return fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
}

// WaitForCondition waits for a condition to be met
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("Timeout waiting for condition")
}

// ContextWithTimeout creates a context with timeout for tests
func ContextWithTimeout(t *testing.T, duration time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	t.Cleanup(cancel)
	return ctx
}
