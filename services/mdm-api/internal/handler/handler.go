package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/mdm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/mdm-api/internal/service"
)

// Handler binds HTTP routes to MDM business logic.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Post("/mdm/validate/entity", h.ValidateEntity)
	app.Get("/api/health", h.HealthCheck)
}

// ValidateEntity validates and standardizes customer/supplier entity payload.
func (h *Handler) ValidateEntity(c *fiber.Ctx) error {
	req := &domain.ValidateEntityRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	result, err := h.service.ValidateAndStandardizeEntity(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    result,
		Message: "entity validated and standardized successfully",
	})
}

// HealthCheck returns service liveness.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "mdm-api",
	})
}
