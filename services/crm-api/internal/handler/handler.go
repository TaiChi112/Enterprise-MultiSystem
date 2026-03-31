package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/crm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/crm-api/internal/service"
)

// Handler holds the service dependency
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(app *fiber.App) {
	// Customer routes
	customers := app.Group("/api/customers")
	customers.Post("/", h.CreateCustomer)
	customers.Get("/:id", h.GetCustomer)
	customers.Get("", h.GetAllCustomers)
	customers.Put("/:id", h.UpdateCustomer)
	customers.Delete("/:id", h.DeleteCustomer)
	customers.Post("/:id/loyalty", h.AwardLoyaltyPoints)

	// Health check
	app.Get("/api/health", h.HealthCheck)
}

// ============================================================================
// CUSTOMER HANDLERS
// ============================================================================

// CreateCustomer creates a new customer
// @Summary Create a new customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param request body domain.CreateCustomerRequest true "Customer data"
// @Success 201 {object} domain.SuccessResponse{data=domain.Customer}
// @Failure 400 {object} domain.ErrorResponse
// @Router /api/customers [post]
func (h *Handler) CreateCustomer(c *fiber.Ctx) error {
	req := &domain.CreateCustomerRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Email) == "" {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "name and email are required",
		})
	}

	customer, err := h.service.CreateCustomer(c.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    customer,
		Message: "customer created successfully",
	})
}

// GetCustomer retrieves a customer by ID
// @Summary Get customer by ID
// @Tags Customers
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} domain.SuccessResponse{data=domain.Customer}
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/customers/{id} [get]
func (h *Handler) GetCustomer(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid customer ID",
		})
	}

	customer, err := h.service.GetCustomer(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "customer not found",
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    customer,
	})
}

// GetAllCustomers retrieves all active customers with pagination
// @Summary Get all customers
// @Tags Customers
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.Customer}
// @Router /api/customers [get]
func (h *Handler) GetAllCustomers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	customers, err := h.service.GetAllCustomers(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if customers == nil {
		customers = []*domain.Customer{} // Return empty array instead of null
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    customers,
	})
}

// UpdateCustomer updates an existing customer
// @Summary Update customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param request body domain.UpdateCustomerRequest true "Customer data to update"
// @Success 200 {object} domain.SuccessResponse{data=domain.Customer}
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/customers/{id} [put]
func (h *Handler) UpdateCustomer(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid customer ID",
		})
	}

	req := &domain.UpdateCustomerRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	customer, err := h.service.UpdateCustomer(c.Context(), id, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    customer,
		Message: "customer updated successfully",
	})
}

// DeleteCustomer deletes a customer (soft delete)
// @Summary Delete customer
// @Tags Customers
// @Produce json
// @Param id path int true "Customer ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/customers/{id} [delete]
func (h *Handler) DeleteCustomer(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid customer ID",
		})
	}

	err = h.service.DeleteCustomer(c.Context(), id)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Message: "customer deleted successfully",
	})
}

// AwardLoyaltyPoints awards loyalty points to a customer
// @Summary Award loyalty points
// @Tags Customers
// @Accept json
// @Produce json
// @Param id path int true "Customer ID"
// @Param request body domain.AwardLoyaltyPointsRequest true "Points to award"
// @Success 200 {object} domain.SuccessResponse{data=domain.Customer}
// @Failure 400,404 {object} domain.ErrorResponse
// @Router /api/customers/{id}/loyalty [post]
func (h *Handler) AwardLoyaltyPoints(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid customer ID",
		})
	}

	req := &domain.AwardLoyaltyPointsRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	customer, err := h.service.AwardLoyaltyPoints(c.Context(), id, req.Points)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "must be greater than") {
			statusCode = http.StatusBadRequest
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    customer,
		Message: "loyalty points awarded successfully",
	})
}

// HealthCheck returns the health status of the service
// @Summary Health check
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/health [get]
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "crm-api",
	})
}
