package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/service"
)

const (
	errInvalidRequestBodyFmt = "invalid request body: %v"
	errInvalidSupplierID     = "invalid supplier ID"
)

// Handler holds the service dependency.
type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) RegisterRoutes(app *fiber.App) {
	suppliers := app.Group("/scm/suppliers")
	suppliers.Post("/", h.CreateSupplier)
	suppliers.Get("/:id", h.GetSupplier)
	suppliers.Get("", h.GetSuppliers)
	suppliers.Put("/:id", h.UpdateSupplier)
	suppliers.Delete("/:id", h.DeleteSupplier)

	app.Post("/scm/replenish", h.Replenish)
	app.Get("/scm/purchase-orders", h.GetPurchaseOrders)

	app.Get("/api/health", h.HealthCheck)
}

func (h *Handler) CreateSupplier(c *fiber.Ctx) error {
	req := &domain.CreateSupplierRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf(errInvalidRequestBodyFmt, err),
		})
	}

	if strings.TrimSpace(req.Name) == "" {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "name is required",
		})
	}

	supplier, err := h.service.CreateSupplier(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    supplier,
		Message: "supplier created successfully",
	})
}

func (h *Handler) GetSupplier(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errInvalidSupplierID,
		})
	}

	supplier, err := h.service.GetSupplier(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{Success: true, Data: supplier})
}

func (h *Handler) GetSuppliers(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	suppliers, err := h.service.GetSuppliers(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if suppliers == nil {
		suppliers = []*domain.Supplier{}
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{Success: true, Data: suppliers})
}

func (h *Handler) UpdateSupplier(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errInvalidSupplierID,
		})
	}

	req := &domain.UpdateSupplierRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf(errInvalidRequestBodyFmt, err),
		})
	}

	supplier, err := h.service.UpdateSupplier(c.Context(), id, req)
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
		Data:    supplier,
		Message: "supplier updated successfully",
	})
}

func (h *Handler) DeleteSupplier(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errInvalidSupplierID,
		})
	}

	err = h.service.DeleteSupplier(c.Context(), id)
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
		Message: "supplier deleted successfully",
	})
}

func (h *Handler) Replenish(c *fiber.Ctx) error {
	req := &domain.ReplenishRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf(errInvalidRequestBodyFmt, err),
		})
	}

	resp, err := h.service.Replenish(c.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()

		if errors.Is(err, service.ErrInvalidReplenishProduct) {
			statusCode = http.StatusBadRequest
			errorMessage = fmt.Sprintf("product_id %d not found in product catalog", req.ProductID)
		}
		if strings.Contains(err.Error(), "must be greater than") {
			statusCode = http.StatusBadRequest
		}
		if strings.Contains(err.Error(), "transmit purchase order to edi") {
			statusCode = http.StatusBadGateway
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errorMessage,
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    resp,
		Message: "replenishment purchase order created and transmitted",
	})
}

func (h *Handler) GetPurchaseOrders(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 100)
	offset := c.QueryInt("offset", 0)

	orders, err := h.service.GetPurchaseOrders(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if orders == nil {
		orders = []*domain.PurchaseOrder{}
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    orders,
	})
}

func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "scm-api",
	})
}
