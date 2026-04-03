package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/idp-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/idp-api/internal/service"
)

// Handler binds HTTP routes to IDP business logic.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/idp/extract", h.Extract)
	app.Get("/api/health", h.HealthCheck)
}

// Extract accepts a file reference and returns simulated extracted data.
func (h *Handler) Extract(c *fiber.Ctx) error {
	req := &domain.ExtractRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	result, err := h.service.SimulateExtraction(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    result,
		Message: "document extracted successfully (simulated)",
	})
}

// HealthCheck returns service liveness.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "idp-api",
	})
}
