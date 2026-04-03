package domain

import (
	"time"
)

// ============================================================================
// PRODUCT - Master data for products
// ============================================================================
type Product struct {
	ID          int       `db:"id" json:"id"`
	SKU         string    `db:"sku" json:"sku"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	Price       float64   `db:"price" json:"price"`
	Cost        *float64  `db:"cost" json:"cost"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// CreateProductRequest - Input DTO for creating a product
type CreateProductRequest struct {
	SKU         string   `json:"sku" validate:"required,min=3,max=50"`
	Name        string   `json:"name" validate:"required,min=1,max=255"`
	Description string   `json:"description" validate:"omitempty,max=1000"`
	Price       float64  `json:"price" validate:"required,gt=0"`
	Cost        *float64 `json:"cost" validate:"omitempty,gte=0"`
}

// ============================================================================
// BRANCH - Physical branch/location data
// ============================================================================
type Branch struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Address   *string   `db:"address" json:"address"`
	Phone     *string   `db:"phone" json:"phone"`
	IsActive  bool      `db:"is_active" json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateBranchRequest - Input DTO for creating a branch
type CreateBranchRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=255"`
	Address string `json:"address" validate:"omitempty,max=500"`
	Phone   string `json:"phone" validate:"omitempty,max=20"`
}

// ============================================================================
// INVENTORY - Stock levels per product per branch
// ============================================================================
type Inventory struct {
	ID         int       `db:"id" json:"id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	BranchID   int       `db:"branch_id" json:"branch_id"`
	Quantity   int       `db:"quantity" json:"quantity"`
	MinimumQty int       `db:"minimum_qty" json:"minimum_qty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`

	// Denormalized fields for response (optional)
	Product *Product `db:"-" json:"product,omitempty"`
	Branch  *Branch  `db:"-" json:"branch,omitempty"`
}

// CreateInventoryRequest - Input DTO for creating inventory records
type CreateInventoryRequest struct {
	ProductID  int `json:"product_id" validate:"required,gt=0"`
	BranchID   int `json:"branch_id" validate:"required,gt=0"`
	Quantity   int `json:"quantity" validate:"required,gte=0"`
	MinimumQty int `json:"minimum_qty" validate:"gte=0"`
}

// UpdateInventoryRequest - Input DTO for updating inventory
type UpdateInventoryRequest struct {
	Quantity   *int `json:"quantity" validate:"omitempty,gte=0"`
	MinimumQty *int `json:"minimum_qty" validate:"omitempty,gte=0"`
}

// AdjustInventoryRequest - Adjust stock (for manual adjustments, stock checks, etc.)
type AdjustInventoryRequest struct {
	QuantityDelta int `json:"quantity_delta" validate:"required"`
}

// ============================================================================
// ORDER - Sales transaction header
// ============================================================================
type Order struct {
	ID           int       `db:"id" json:"id"`
	BranchID     int       `db:"branch_id" json:"branch_id"`
	CustomerName *string   `db:"customer_name" json:"customer_name"`
	TotalAmount  float64   `db:"total_amount" json:"total_amount"`
	Status       string    `db:"status" json:"status"` // pending, completed, cancelled, refunded
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	// Denormalized fields for response (optional)
	OrderItems []*OrderItem `db:"-" json:"order_items,omitempty"`
}

// OrderStatus constants
const (
	OrderStatusPending   = "pending"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusRefunded  = "refunded"
)

// CreateOrderRequest - Input DTO for creating an order
type CreateOrderRequest struct {
	BranchID     int                  `json:"branch_id" validate:"required,gt=0"`
	CustomerName string               `json:"customer_name" validate:"omitempty,max=255"`
	OrderItems   []CreateOrderItemReq `json:"order_items" validate:"required,min=1,dive,required"`
}

// CreateOrderItemReq - Line item in order request
type CreateOrderItemReq struct {
	ProductID int     `json:"product_id" validate:"required,gt=0"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	Discount  float64 `json:"discount" validate:"omitempty,gte=0"`
}

// ProcessSaleRequest - Alternative input DTO for processing a sale
type ProcessSaleRequest struct {
	BranchID     int               `json:"branch_id" validate:"required,gt=0"`
	CustomerName string            `json:"customer_name" validate:"omitempty,max=255"`
	Items        []SaleItemRequest `json:"items" validate:"required,min=1,dive,required"`
}

// SaleItemRequest - Item in sale
type SaleItemRequest struct {
	ProductID int     `json:"product_id" validate:"required,gt=0"`
	Quantity  int     `json:"quantity" validate:"required,gt=0"`
	Discount  float64 `json:"discount" validate:"omitempty,gte=0"`
}

// ============================================================================
// ORDER_ITEM - Sales transaction details (line items)
// ============================================================================
type OrderItem struct {
	ID        int       `db:"id" json:"id"`
	OrderID   int       `db:"order_id" json:"order_id"`
	ProductID int       `db:"product_id" json:"product_id"`
	Quantity  int       `db:"quantity" json:"quantity"`
	UnitPrice float64   `db:"unit_price" json:"unit_price"`
	Discount  float64   `db:"discount" json:"discount"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	// Denormalized fields for response (optional)
	Product *Product `db:"-" json:"product,omitempty"`
}

// OrderItemResponse - Response DTO with calculated subtotal
type OrderItemResponse struct {
	ID        int       `json:"id"`
	ProductID int       `json:"product_id"`
	Product   *Product  `json:"product,omitempty"`
	Quantity  int       `json:"quantity"`
	UnitPrice float64   `json:"unit_price"`
	Discount  float64   `json:"discount"`
	Subtotal  float64   `json:"subtotal"` // (unit_price * quantity) - discount
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ============================================================================
// RESPONSE DTOS
// ============================================================================

// SuccessResponse - Generic success response wrapper
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse - Generic error response wrapper
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}

// PaginatedResponse - Generic paginated response
type PaginatedResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}

// OrderResponse - Order with calculated totals (response DTO)
type OrderResponse struct {
	ID           int                 `json:"id"`
	BranchID     int                 `json:"branch_id"`
	Branch       *Branch             `json:"branch,omitempty"`
	CustomerName *string             `json:"customer_name"`
	Status       string              `json:"status"`
	OrderItems   []OrderItemResponse `json:"order_items"`
	TotalAmount  float64             `json:"total_amount"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
}
