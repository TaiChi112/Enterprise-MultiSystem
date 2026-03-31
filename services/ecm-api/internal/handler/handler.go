package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/ecm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/ecm-api/internal/service"
)

// Handler binds HTTP routes to ECM business logic.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/ecm/upload", h.UploadFile)
	app.Get("/api/health", h.HealthCheck)
}

// UploadFile accepts multipart/form-data and stores image/PDF files locally.
func (h *Handler) UploadFile(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to read multipart file field 'file': %v", err),
		})
	}

	result, err := h.service.SaveUpload(fileHeader)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    result,
		Message: "file uploaded successfully",
	})
}

// HealthCheck returns service liveness.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "ecm-api",
	})
}
