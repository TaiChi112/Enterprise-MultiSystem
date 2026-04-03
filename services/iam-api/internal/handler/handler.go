package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/iam-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/iam-api/internal/service"
)

// Handler wires HTTP handlers to IAM services.
type Handler struct {
	auth *service.AuthService
}

func NewHandler(auth *service.AuthService) *Handler {
	return &Handler{auth: auth}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/api/health", h.HealthCheck)
	app.Post("/login", h.Login)
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "iam-api",
	})
}

func (h *Handler) Login(c *fiber.Ctx) error {
	req := &domain.LoginRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid request body",
		})
	}

	resp, err := h.auth.Login(req)
	if err != nil {
		return c.Status(http.StatusUnauthorized).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    resp,
	})
}
