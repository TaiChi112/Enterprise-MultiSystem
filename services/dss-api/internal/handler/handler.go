package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/dss-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/dss-api/internal/service"
)

// Handler binds HTTP routes to DSS service logic.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/dss/insights/sales-trend", h.GetSalesTrend)
	app.Get("/api/health", h.HealthCheck)
}

// GetSalesTrend returns sales trend insights derived from ERP data.
func (h *Handler) GetSalesTrend(c *fiber.Ctx) error {
	req := &domain.SalesTrendRequest{
		Period: c.Query("period", ""),
		Months: c.QueryInt("months", 3),
	}

	insight, err := h.service.GetSalesTrendInsights(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    insight,
		Message: "sales trend insight generated successfully",
	})
}

// HealthCheck returns service liveness.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "dss-api",
	})
}
