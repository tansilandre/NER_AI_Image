package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ner-studio/api/internal/middleware"
	"github.com/ner-studio/api/internal/service"
)

// GenerationHandler handles generation endpoints
type GenerationHandler struct {
	generationService *service.GenerationService
}

// NewGenerationHandler creates a new generation handler
func NewGenerationHandler(genService *service.GenerationService) *GenerationHandler {
	return &GenerationHandler{
		generationService: genService,
	}
}

// CreateGenerationRequest request body
type CreateGenerationRequest struct {
	BasePrompt      string   `json:"base_prompt" validate:"required"`
	ReferenceImages []string `json:"reference_images"`
	ProductImages   []string `json:"product_images"`
	ProviderID      string   `json:"provider_id" validate:"required"`
	NumVariations   int      `json:"num_variations"`
}

// CreateGeneration starts a new image generation
func (h *GenerationHandler) CreateGeneration(c *fiber.Ctx) error {
	var req CreateGenerationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validation
	if req.BasePrompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "base_prompt is required",
		})
	}
	if req.ProviderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "provider_id is required",
		})
	}

	userID := middleware.GetUserID(c)
	orgID := middleware.GetOrganizationID(c)

	// TODO: Get orgID from user's profile (need middleware to set this)
	// For now, we'll need to get it from the database

	gen, err := h.generationService.CreateGeneration(c.Context(), service.CreateGenerationRequest{
		UserID:          userID,
		OrganizationID:  orgID, // This might be empty, need to handle
		BasePrompt:      req.BasePrompt,
		ReferenceImages: req.ReferenceImages,
		ProductImages:   req.ProductImages,
		ProviderID:      req.ProviderID,
		NumVariations:   req.NumVariations,
	})
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"id":      gen.ID,
		"status":  gen.Status,
		"message": "Generation started",
	})
}

// ListGenerations lists user's generations
func (h *GenerationHandler) ListGenerations(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	_ = userID

	// Get pagination
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	// TODO: Get orgID from user's profile
	// generations, err := h.generationService.ListGenerations(c.Context(), orgID, limit, offset)

	return c.JSON(fiber.Map{
		"generations": []interface{}{},
		"limit":       limit,
		"offset":      offset,
	})
}

// GetGeneration gets a single generation
func (h *GenerationHandler) GetGeneration(c *fiber.Ctx) error {
	id := c.Params("id")
	genID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid generation ID",
		})
	}

	gen, images, err := h.generationService.GetGeneration(c.Context(), genID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Generation not found",
		})
	}

	return c.JSON(fiber.Map{
		"generation": gen,
		"images":     images,
	})
}

// HandleCallback handles provider callbacks
func (h *GenerationHandler) HandleCallback(c *fiber.Ctx) error {
	providerSlug := c.Params("provider")
	
	body := c.Body()
	
	if err := h.generationService.HandleCallback(c.Context(), providerSlug, body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusOK)
}
