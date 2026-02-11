package handler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/ner-studio/api/internal/middleware"
	"github.com/ner-studio/api/internal/service"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	uploadService *service.UploadService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(uploadService *service.UploadService) *UploadHandler {
	return &UploadHandler{
		uploadService: uploadService,
	}
}

// UploadImage handles image uploads
func (h *UploadHandler) UploadImage(c *fiber.Ctx) error {
	_ = middleware.GetUserID(c)
	orgID := middleware.GetOrganizationID(c)

	// TODO: Get orgID from profile
	if orgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	// Get folder type
	folder := c.FormValue("folder", "uploads")
	if folder != "references" && folder != "products" && folder != "uploads" {
		folder = "uploads"
	}

	// Get file from request
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No image file provided",
		})
	}

	// Check file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File too large (max 10MB)",
		})
	}

	// Open file
	fileReader, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to read file",
		})
	}
	defer fileReader.Close()

	// Upload
	result, err := h.uploadService.UploadImage(c.Context(), orgID, folder, file.Filename, fileReader, file.Size)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"url":      result.URL,
		"key":      result.Key,
		"filename": result.Filename,
	})
}

// GetUploadURL returns a presigned URL for direct upload
func (h *UploadHandler) GetUploadURL(c *fiber.Ctx) error {
	// TODO: Implement presigned URL generation
	return c.JSON(fiber.Map{
		"upload_url": "",
		"fields":     fiber.Map{},
	})
}
