package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/erp-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/erp-api/internal/service"
)

// Handler holds the service dependency.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	app.Get("/erp/financial-summary", h.GetFinancialSummary)
	app.Post("/erp/financial-summary", h.PostFinancialSummary)
	app.Get("/api/health", h.HealthCheck)
}

// ============================================================================
// FINANCIAL SUMMARY HANDLERS
// ============================================================================

// GetFinancialSummary retrieves financial summary. (Skeleton for STEP 4)
// @Summary Get financial summary (P&L Report)
// @Tags Financial
// @Produce json
// @Param period query string false "Period (YYYY-MM format, default: current month)"
// @Success 200 {object} domain.SuccessResponse{data=domain.FinancialSummary}
// @Failure 500 {object} domain.ErrorResponse
// @Router /erp/financial-summary [get]
func (h *Handler) GetFinancialSummary(c *fiber.Ctx) error {
	period := c.Query("period", "")

	req := &domain.FinancialSummaryRequest{
		Period: period,
	}

	summary, err := h.service.GetFinancialSummary(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPeriod) {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to retrieve financial summary: %v", err),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    summary,
		Message: "financial summary aggregated successfully",
	})
}

// PostFinancialSummary is an alternative endpoint for retrieving financial summary with request body.
func (h *Handler) PostFinancialSummary(c *fiber.Ctx) error {
	req := &domain.FinancialSummaryRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	summary, err := h.service.GetFinancialSummary(c.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidPeriod) {
			return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
				Success: false,
				Error:   err.Error(),
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("failed to retrieve financial summary: %v", err),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    summary,
		Message: "financial summary aggregated successfully",
	})
}

// ============================================================================
// HEALTH CHECK
// ============================================================================

// HealthCheck returns the health status of the service.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "erp-api",
		"message": "ERP Aggregator Service (non-transactional)",
	})
}
