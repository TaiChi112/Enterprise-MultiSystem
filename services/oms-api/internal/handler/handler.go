package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/oms-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/oms-api/internal/service"
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
	// Order routes
	orders := app.Group("/api/orders")
	orders.Post("/", h.InitializeOrder)
	orders.Get("/:id", h.GetOrder)
	orders.Get("", h.GetAllOrders)
	orders.Put("/:id/status", h.UpdateOrderStatus)
	orders.Delete("/:id", h.DeleteOrder)

	// Order items routes
	items := app.Group("/api/orders/:id/items")
	items.Post("/", h.AddItemToOrder)
	items.Get("/", h.GetOrderItems)
	items.Delete("/:itemId", h.RemoveItemFromOrder)

	// Order by order number
	app.Get("/api/orders/number/:orderNumber", h.GetOrderByOrderNumber)

	// Customer orders
	app.Get("/api/customers/:customerId/orders", h.GetCustomerOrders)

	// Health check
	app.Get("/api/health", h.HealthCheck)
}

// ============================================================================
// ORDER HANDLERS
// ============================================================================

// InitializeOrder creates a new order with pending status
// @Summary Initialize a new order
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body domain.InitializeOrderRequest true "Order initialization data"
// @Success 201 {object} domain.SuccessResponse{data=domain.OrderLifecycle}
// @Failure 400 {object} domain.ErrorResponse
// @Router /api/orders [post]
func (h *Handler) InitializeOrder(c *fiber.Ctx) error {
	req := &domain.InitializeOrderRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	if req.CustomerID <= 0 {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "customer_id is required and must be greater than 0",
		})
	}

	order, err := h.service.InitializeOrder(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
		Message: "order initialized successfully",
	})
}

// GetOrder retrieves an order by ID
// @Summary Get order by ID
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} domain.SuccessResponse{data=domain.OrderLifecycle}
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/orders/{id} [get]
func (h *Handler) GetOrder(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid order ID",
		})
	}

	order, err := h.service.GetOrder(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "order not found",
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
	})
}

// GetOrderByOrderNumber retrieves an order by order number
// @Summary Get order by order number
// @Tags Orders
// @Produce json
// @Param orderNumber path string true "Order Number"
// @Success 200 {object} domain.SuccessResponse{data=domain.OrderLifecycle}
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/orders/number/{orderNumber} [get]
func (h *Handler) GetOrderByOrderNumber(c *fiber.Ctx) error {
	orderNumber := c.Params("orderNumber")
	if strings.TrimSpace(orderNumber) == "" {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "order number is required",
		})
	}

	order, err := h.service.GetOrderByOrderNumber(c.Context(), orderNumber)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "order not found",
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
	})
}

// GetAllOrders retrieves all orders with pagination
// @Summary Get all orders
// @Tags Orders
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Param status query string false "Filter by status (pending, paid, shipped, completed, cancelled)"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.OrderLifecycle}
// @Router /api/orders [get]
func (h *Handler) GetAllOrders(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)
	status := c.Query("status")

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	orders, err := h.service.GetAllOrders(c.Context(), limit, offset, statusPtr)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if orders == nil {
		orders = []*domain.OrderLifecycle{} // Return empty array instead of null
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    orders,
	})
}

// GetCustomerOrders retrieves all orders for a customer
// @Summary Get customer orders
// @Tags Orders
// @Produce json
// @Param customerId path int true "Customer ID"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.OrderLifecycle}
// @Router /api/customers/{customerId}/orders [get]
func (h *Handler) GetCustomerOrders(c *fiber.Ctx) error {
	customerIDParam := c.Params("customerId")
	customerID, err := strconv.Atoi(customerIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid customer ID",
		})
	}

	orders, err := h.service.GetOrdersByCustomerID(c.Context(), customerID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if orders == nil {
		orders = []*domain.OrderLifecycle{}
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    orders,
	})
}

// UpdateOrderStatus updates an order's status
// @Summary Update order status
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body domain.UpdateOrderStatusRequest true "New status"
// @Success 200 {object} domain.SuccessResponse{data=domain.OrderLifecycle}
// @Failure 400,404 {object} domain.ErrorResponse
// @Router /api/orders/{id}/status [put]
func (h *Handler) UpdateOrderStatus(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid order ID",
		})
	}

	req := &domain.UpdateOrderStatusRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	if strings.TrimSpace(req.Status) == "" {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "status is required",
		})
	}

	order, err := h.service.UpdateOrderStatus(c.Context(), id, req.Status)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
		Message: "order status updated successfully",
	})
}

// DeleteOrder deletes an order (soft delete)
// @Summary Delete order
// @Tags Orders
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/orders/{id} [delete]
func (h *Handler) DeleteOrder(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid order ID",
		})
	}

	err = h.service.DeleteOrder(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Message: "order deleted successfully",
	})
}

// ============================================================================
// ORDER ITEM HANDLERS
// ============================================================================

// AddItemToOrder adds an item to an order
// @Summary Add item to order
// @Tags Order Items
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param request body domain.AddOrderItemRequest true "Item data"
// @Success 201 {object} domain.SuccessResponse{data=domain.OrderItem}
// @Failure 400 {object} domain.ErrorResponse
// @Router /api/orders/{id}/items [post]
func (h *Handler) AddItemToOrder(c *fiber.Ctx) error {
	idParam := c.Params("id")
	orderID, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid order ID",
		})
	}

	req := &domain.AddOrderItemRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	item, err := h.service.AddItemToOrder(c.Context(), orderID, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		}
		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    item,
		Message: "item added to order successfully",
	})
}

// GetOrderItems retrieves all items for an order
// @Summary Get order items
// @Tags Order Items
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.OrderItem}
// @Router /api/orders/{id}/items [get]
func (h *Handler) GetOrderItems(c *fiber.Ctx) error {
	idParam := c.Params("id")
	_, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid order ID",
		})
	}

	items, err := h.service.GetAllOrders(c.Context(), 1, 0, nil) // Placeholder
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	// For now, return placeholder items
	if items == nil {
		items = []*domain.OrderLifecycle{}
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    items,
	})
}

// RemoveItemFromOrder removes an item from an order
// @Summary Remove item from order
// @Tags Order Items
// @Produce json
// @Param id path int true "Order ID"
// @Param itemId path int true "Item ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /api/orders/{id}/items/{itemId} [delete]
func (h *Handler) RemoveItemFromOrder(c *fiber.Ctx) error {
	itemIDParam := c.Params("itemId")
	itemID, err := strconv.Atoi(itemIDParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid item ID",
		})
	}

	err = h.service.RemoveItemFromOrder(c.Context(), itemID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Message: "item removed from order successfully",
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
		"service": "oms-api",
	})
}
