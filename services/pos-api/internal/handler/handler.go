package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/service"
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
	// Product routes
	products := app.Group("/api/products")
	products.Post("/", h.CreateProduct)
	products.Get("/:id", h.GetProduct)
	products.Get("", h.GetAllProducts)

	// Branch routes
	branches := app.Group("/api/branches")
	branches.Post("/", h.CreateBranch)
	branches.Get("/:id", h.GetBranch)
	branches.Get("", h.GetAllBranches)

	// Inventory routes
	inventory := app.Group("/api/inventory")
	inventory.Post("/", h.CreateInventory)
	inventory.Get("/branch/:branchId", h.GetBranchInventory)
	inventory.Get("/low-stock/:branchId", h.GetLowStockItems)
	inventory.Get("/product/:productId/branch/:branchId", h.GetInventory)

	// Order routes
	orders := app.Group("/api/orders")
	orders.Post("/", h.CreateOrder)
	orders.Get("/:id", h.GetOrder)
	orders.Get("/branch/:branchId", h.GetOrders)

	// Sale processing route
	app.Post("/api/sales", h.ProcessSale)

	// Health check
	app.Get("/api/health", h.HealthCheck)
}

// ============================================================================
// PRODUCT HANDLERS
// ============================================================================

// CreateProduct creates a new product
// @Summary Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param request body domain.CreateProductRequest true "Product data"
// @Success 201 {object} domain.SuccessResponse{data=domain.Product}
// @Failure 400 {object} domain.ErrorResponse
// @Router /api/products [post]
func (h *Handler) CreateProduct(c *fiber.Ctx) error {
	req := &domain.CreateProductRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	product, err := h.service.CreateProduct(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    product,
		Message: "Product created successfully",
	})
}

// GetProduct retrieves a product by ID
func (h *Handler) GetProduct(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid product ID",
		})
	}

	product, err := h.service.GetProduct(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    product,
	})
}

// GetAllProducts retrieves all products
func (h *Handler) GetAllProducts(c *fiber.Ctx) error {
	limit := c.Query("limit", "10")
	offset := c.Query("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	if limitInt <= 0 || limitInt > 100 {
		limitInt = 10
	}
	if offsetInt < 0 {
		offsetInt = 0
	}

	products, err := h.service.GetAllProducts(c.Context(), limitInt, offsetInt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if products == nil {
		products = make([]*domain.Product, 0)
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    products,
	})
}

// ============================================================================
// BRANCH HANDLERS
// ============================================================================

// CreateBranch creates a new branch
func (h *Handler) CreateBranch(c *fiber.Ctx) error {
	req := &domain.CreateBranchRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	branch, err := h.service.CreateBranch(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    branch,
		Message: "Branch created successfully",
	})
}

// GetBranch retrieves a branch by ID
func (h *Handler) GetBranch(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid branch ID",
		})
	}

	branch, err := h.service.GetBranch(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    branch,
	})
}

// GetAllBranches retrieves all branches
func (h *Handler) GetAllBranches(c *fiber.Ctx) error {
	branches, err := h.service.GetAllBranches(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if branches == nil {
		branches = make([]*domain.Branch, 0)
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    branches,
	})
}

// ============================================================================
// INVENTORY HANDLERS
// ============================================================================

// CreateInventory creates a new inventory record
func (h *Handler) CreateInventory(c *fiber.Ctx) error {
	req := &domain.CreateInventoryRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	inv, err := h.service.CreateInventory(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    inv,
		Message: "Inventory created successfully",
	})
}

// GetInventory retrieves inventory for a product at a specific branch
func (h *Handler) GetInventory(c *fiber.Ctx) error {
	productID, err := strconv.Atoi(c.Params("productId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid product ID",
		})
	}

	branchID, err := strconv.Atoi(c.Params("branchId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid branch ID",
		})
	}

	inv, err := h.service.GetInventory(c.Context(), productID, branchID)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    inv,
	})
}

// GetBranchInventory retrieves all inventory for a branch
func (h *Handler) GetBranchInventory(c *fiber.Ctx) error {
	branchID, err := strconv.Atoi(c.Params("branchId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid branch ID",
		})
	}

	limit := c.Query("limit", "100")
	offset := c.Query("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	inventory, err := h.service.GetBranchInventory(c.Context(), branchID, limitInt, offsetInt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if inventory == nil {
		inventory = make([]*domain.Inventory, 0)
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    inventory,
	})
}

// GetLowStockItems retrieves low stock items for a branch
func (h *Handler) GetLowStockItems(c *fiber.Ctx) error {
	branchID, err := strconv.Atoi(c.Params("branchId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid branch ID",
		})
	}

	inventory, err := h.service.GetLowStockItems(c.Context(), branchID)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if inventory == nil {
		inventory = make([]*domain.Inventory, 0)
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    inventory,
	})
}

// ============================================================================
// ORDER HANDLERS
// ============================================================================

// CreateOrder creates a new order
func (h *Handler) CreateOrder(c *fiber.Ctx) error {
	req := &domain.CreateOrderRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	return c.Status(http.StatusNotImplemented).JSON(domain.ErrorResponse{
		Success: false,
		Error:   "Use /api/sales endpoint instead",
	})
}

// ProcessSale processes a complete sale with inventory deduction
func (h *Handler) ProcessSale(c *fiber.Ctx) error {
	req := &domain.ProcessSaleRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	order, err := h.service.ProcessSale(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
		Message: "Sale processed successfully",
	})
}

// GetOrder retrieves an order by ID
func (h *Handler) GetOrder(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
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
			Error:   err.Error(),
		})
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    order,
	})
}

// GetOrders retrieves orders for a branch
func (h *Handler) GetOrders(c *fiber.Ctx) error {
	branchID, err := strconv.Atoi(c.Params("branchId"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid branch ID",
		})
	}

	limit := c.Query("limit", "50")
	offset := c.Query("offset", "0")

	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	orders, err := h.service.GetOrders(c.Context(), branchID, limitInt, offsetInt)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if orders == nil {
		orders = make([]*domain.Order, 0)
	}

	return c.JSON(domain.SuccessResponse{
		Success: true,
		Data:    orders,
	})
}

// ============================================================================
// HEALTH CHECK
// ============================================================================

// HealthCheck is a simple health check endpoint
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"message": "POS & WMS MVP API is running",
	})
}
