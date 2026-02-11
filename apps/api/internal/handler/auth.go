package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ner-studio/api/internal/middleware"
	"github.com/ner-studio/api/internal/service"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *service.AuthService
	jwtSecret   string
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		jwtSecret:   jwtSecret,
	}
}

// LoginRequest request body
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Login handles user login
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Authenticate user
	user, err := h.authService.AuthenticateUser(c.Context(), req.Email, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid credentials",
		})
	}

	// Get user's organization and role
	profile, err := h.authService.GetUserProfile(c.Context(), user.ID.String())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get user profile",
		})
	}

	// Generate JWT
	token, err := middleware.GenerateJWT(
		user.ID.String(),
		profile.OrganizationID.String(),
		profile.Role,
		h.jwtSecret,
		24, // 24 hours expiry
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    user.ID,
			"email": user.Email,
			"name":  profile.FullName,
			"role":  profile.Role,
		},
	})
}

// RegisterRequest request body
type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=8"`
	FullName   string `json:"full_name" validate:"required"`
	OrgName    string `json:"org_name" validate:"required"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create organization and user
	org, profile, err := h.authService.CreateOrganizationAndProfile(
		c.Context(),
		req.Email,
		req.FullName,
		req.OrgName,
		req.Password,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate JWT
	token, err := middleware.GenerateJWT(
		profile.UserID.String(),
		org.ID.String(),
		profile.Role,
		h.jwtSecret,
		24,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"token": token,
		"user": fiber.Map{
			"id":    profile.UserID,
			"email": req.Email,
			"name":  profile.FullName,
			"role":  profile.Role,
		},
		"organization": fiber.Map{
			"id":   org.ID,
			"name": org.Name,
			"slug": org.Slug,
		},
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get current user from context
	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrganizationID(c)
	role := middleware.GetRole(c)

	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Not authenticated",
		})
	}

	// Generate new token
	token, err := middleware.GenerateJWT(userID, orgID, role, h.jwtSecret, 24)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}
