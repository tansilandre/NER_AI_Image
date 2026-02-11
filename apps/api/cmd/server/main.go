package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/ner-studio/api/internal/config"
	"github.com/ner-studio/api/internal/external"
	"github.com/ner-studio/api/internal/handler"
	"github.com/ner-studio/api/internal/middleware"
	"github.com/ner-studio/api/internal/model"
	"github.com/ner-studio/api/internal/provider"
	"github.com/ner-studio/api/internal/repository"
	"github.com/ner-studio/api/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize repository
	repo, err := repository.NewRepository(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer repo.Close()

	// Test database connection
	if err := repo.Ping(context.Background()); err != nil {
		log.Printf("Warning: Database ping failed: %v", err)
	} else {
		log.Println("✅ Database connected successfully!")
	}

	// Initialize R2 client
	r2Client, err := external.NewR2Client(external.R2Config{
		AccountID:       cfg.R2AccountID,
		AccessKeyID:     cfg.R2AccessKeyID,
		SecretAccessKey: cfg.R2SecretAccessKey,
		BucketName:      cfg.R2BucketName,
		PublicURL:       cfg.R2PublicURL,
	})
	if err != nil {
		log.Fatalf("Failed to initialize R2 client: %v", err)
	}

	// Initialize provider factory and load providers
	factory := provider.NewFactory()
	loadProviders(factory, cfg)

	// Initialize services
	authService := service.NewAuthService(repo)
	generationService := service.NewGenerationService(repo, factory, cfg.CallbackBaseURL)
	uploadService := service.NewUploadService(r2Client)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, cfg.JWTSecret)
	generationHandler := handler.NewGenerationHandler(generationService)
	uploadHandler := handler.NewUploadHandler(uploadService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "NER Studio API",
		ErrorHandler:          errorHandler,
		ReadBufferSize:        8192,  // Increase header size limit
		WriteBufferSize:       8192,
		DisableStartupMessage: false,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(middleware.NewLogger())
	app.Use(middleware.NewCORS("*"))

	// Setup API documentation (Scalar)
	handler.SetupDocs(app)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"version":   "1.0.0",
			"database":  "connected",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	api.Post("/auth/login", authHandler.Login)
	api.Post("/auth/register", authHandler.Register)
	api.Post("/auth/refresh", authHandler.RefreshToken)

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.NewAuthMiddleware(middleware.AuthConfig{
		JWTSecret: cfg.JWTSecret,
	}))
	protected.Use(middleware.ProfileMiddleware(repo))

	// Generation routes
	protected.Post("/generations", generationHandler.CreateGeneration)
	protected.Get("/generations", generationHandler.ListGenerations)
	protected.Get("/generations/:id", generationHandler.GetGeneration)

	// Gallery routes
	protected.Get("/gallery", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"images": []interface{}{}})
	})

	// Upload routes
	protected.Post("/uploads", uploadHandler.UploadImage)

	// Callback routes (public - no auth)
	api.Post("/callbacks/:provider", generationHandler.HandleCallback)

	// Admin routes
	admin := protected.Group("/admin", middleware.RequireAdmin())
	admin.Get("/organization", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"organization": nil})
	})
	admin.Get("/members", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"members": []interface{}{}})
	})
	admin.Post("/members/invite", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"invitation": nil})
	})
	admin.Get("/credits/history", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"transactions": []interface{}{}})
	})
	admin.Post("/credits", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	// Provider admin routes
	admin.Get("/providers", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"providers": []interface{}{}})
	})
	admin.Post("/providers", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"provider": nil})
	})
	admin.Patch("/providers/:id", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"provider": nil})
	})
	admin.Delete("/providers/:id", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})
	admin.Post("/providers/:slug/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	// Public provider list (for users)
	protected.Get("/providers", func(c *fiber.Ctx) error {
		category := c.Query("category")
		_ = category
		return c.JSON(fiber.Map{"providers": []interface{}{}})
	})

	// Start server
	go func() {
		port := cfg.Port
		if port == "" {
			port = "8080"
		}
		log.Printf("Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}
	log.Println("Server stopped")
}

// loadProviders initializes providers from environment or DB
func loadProviders(factory *provider.Factory, cfg *config.Config) {
	// Register OpenAI vision provider
	if cfg.OpenAIAPIKey != "" && cfg.OpenAIAPIKey != "..." {
		factory.RegisterVisionProvider(provider.NewOpenAIVisionProvider(cfg.OpenAIAPIKey, ""))
		log.Println("Registered OpenAI vision provider")
	}

	// Register KieAI providers
	if cfg.KieAIAPIKey != "" {
		// LLM providers via KieAI
		kieGemini3 := provider.NewKieAIProvider("kieai-gemini3", cfg.KieAIAPIKey, "", "gemini-3.0", model.ProviderConfig{
			TimeoutMs:  60000,
			MaxRetries: 3,
		})
		factory.RegisterLLMProvider(model.Provider{
			Slug:     "kieai-gemini3",
			Name:     "Kie.ai Gemini 3.0",
			Priority: 0,
		}, kieGemini3)

		kieGemini25 := provider.NewKieAIProvider("kieai-gemini25", cfg.KieAIAPIKey, "", "gemini-2.5", model.ProviderConfig{
			TimeoutMs:  60000,
			MaxRetries: 3,
		})
		factory.RegisterLLMProvider(model.Provider{
			Slug:     "kieai-gemini25",
			Name:     "Kie.ai Gemini 2.5",
			Priority: 1,
		}, kieGemini25)

		// Image generation providers
		factory.RegisterImageProvider("kieai-seedream", provider.NewKieAIProvider("kieai-seedream", cfg.KieAIAPIKey, "", "seedream-v1", model.ProviderConfig{
			TimeoutMs: 120000,
		}))
		factory.RegisterImageProvider("kieai-nano", provider.NewKieAIProvider("kieai-nano", cfg.KieAIAPIKey, "", "nano-banana-pro", model.ProviderConfig{
			TimeoutMs: 120000,
		}))

		log.Println("Registered KieAI providers")
	}

	// Register Google Gemini (direct) as fallback
	if cfg.GeminiAPIKey != "" && cfg.GeminiAPIKey != "..." {
		gemini := provider.NewGeminiProvider("google-gemini", cfg.GeminiAPIKey, "", "gemini-2.0-flash", model.ProviderConfig{
			TimeoutMs: 60000,
		})
		factory.RegisterLLMProvider(model.Provider{
			Slug:     "google-gemini",
			Name:     "Google Gemini (Direct)",
			Priority: 2,
		}, gemini)
		log.Println("Registered Google Gemini provider")
	}
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	message := "Internal server error"

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}
