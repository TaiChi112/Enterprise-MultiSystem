package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/edi-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/edi-api/internal/service"
)

// Handler binds HTTP routes to EDI service logic.
type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/health", h.HealthCheck)
	app.Post("/edi/transmit", h.TransmitPurchaseOrder)
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "edi-api",
	})
}

func (h *Handler) TransmitPurchaseOrder(c *fiber.Ctx) error {
	req := &domain.InternalPOPayload{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	if err := service.ValidateInternalPO(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	resp, err := h.service.TransformAndTransmit(req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(resp)
}
