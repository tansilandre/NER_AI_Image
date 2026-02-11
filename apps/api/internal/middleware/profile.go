package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ner-studio/api/internal/repository"
)

// ProfileMiddleware loads user profile and sets org context
func ProfileMiddleware(repo *repository.Repository) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID := GetUserID(c)
		if userID == "" {
			return c.Next() // Skip if no user
		}

		// Load profile
		profile, err := repo.GetProfileByUserID(c.Context(), userID)
		if err != nil {
			// User exists but no profile - let them complete onboarding
			return c.Next()
		}

		// Set org and role in context
		c.Locals(string(OrganizationIDKey), profile.OrganizationID.String())
		c.Locals(string(RoleKey), profile.Role)

		return c.Next()
	}
}
