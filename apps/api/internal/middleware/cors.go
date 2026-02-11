package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// NewCORS creates CORS middleware
func NewCORS(allowOrigins string) fiber.Handler {
	// Default to localhost for development
	if allowOrigins == "" || allowOrigins == "*" {
		allowOrigins = "http://localhost:5173,http://localhost:3000"
	}
	return cors.New(cors.Config{
		AllowOrigins:     allowOrigins,
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true,
		MaxAge:           86400,
	})
}
