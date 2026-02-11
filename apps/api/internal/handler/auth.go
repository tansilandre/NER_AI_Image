package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ner-studio/api/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// AuthCallbackRequest request body
type AuthCallbackRequest struct {
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// AuthCallback handles post-OAuth callback
func (h *AuthHandler) AuthCallback(c *fiber.Ctx) error {
	var req AuthCallbackRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate token and get/create user
	// TODO: Implement

	return c.JSON(fiber.Map{
		"user": fiber.Map{
			"id": "user-id",
			"email": "user@example.com",
		},
		"profile": nil, // null if needs onboarding
	})
}

// OnboardingRequest request body
type OnboardingRequest struct {
	Action string `json:"action" validate:"required,oneof=create join"`
	// For create
	OrgName string `json:"org_name,omitempty"`
	// For join
	InviteToken string `json:"invite_token,omitempty"`
}

// Onboarding handles create/join organization
func (h *AuthHandler) Onboarding(c *fiber.Ctx) error {
	var req OnboardingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// TODO: Implement

	return c.JSON(fiber.Map{
		"organization": fiber.Map{
			"id": "org-id",
			"name": req.OrgName,
		},
		"profile": fiber.Map{
			"role": "admin",
		},
	})
}
