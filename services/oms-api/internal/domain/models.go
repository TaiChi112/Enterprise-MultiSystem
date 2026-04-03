package domain

import (
	"time"
)

// ============================================================================
// ORDER LIFECYCLE STATUS CONSTANTS
// ============================================================================

const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusShipped   = "shipped"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
)

// ============================================================================
// ORDER LIFECYCLE - Order status tracking for cross-channel orders
// ============================================================================
type OrderLifecycle struct {
	ID          int       `db:"id" json:"id"`
	OrderNumber string    `db:"order_number" json:"order_number"`
	CustomerID  int       `db:"customer_id" json:"customer_id"`
	Status      string    `db:"status" json:"status"` // pending, paid, shipped, completed, cancelled
	TotalAmount float64   `db:"total_amount" json:"total_amount"`
	Description *string   `db:"description" json:"description"`
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`

	// Metadata fields for response (optional)
	OrderItems []*OrderItem `db:"-" json:"order_items,omitempty"`
}

// OrderItem - Line items for orders in OMS
type OrderItem struct {
	ID          int       `db:"id" json:"id"`
	OrderID     int       `db:"order_id" json:"order_id"`
	ProductID   int       `db:"product_id" json:"product_id"`
	ProductName *string   `db:"product_name" json:"product_name"` // Denormalized for reference
	Quantity    int       `db:"quantity" json:"quantity"`
	UnitPrice   float64   `db:"unit_price" json:"unit_price"`
	LineTotal   float64   `db:"line_total" json:"line_total"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// REQUEST / RESPONSE DTOs
// ============================================================================

// InitializeOrderRequest - Input DTO for creating a new order
type InitializeOrderRequest struct {
	CustomerID  int    `json:"customer_id" validate:"required,gt=0"`
	Description string `json:"description" validate:"omitempty,max=500"`
}

// CreateOrderItemRequest - Input DTO for adding items to an order
type CreateOrderItemRequest struct {
	ProductID   int     `json:"product_id" validate:"required,gt=0"`
	ProductName string  `json:"product_name" validate:"omitempty,max=255"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// UpdateOrderStatusRequest - Input DTO for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending paid shipped completed cancelled"`
}

// AddOrderItemRequest - Input DTO for adding items to an existing order
type AddOrderItemRequest struct {
	ProductID   int     `json:"product_id" validate:"required,gt=0"`
	ProductName string  `json:"product_name" validate:"omitempty,max=255"`
	Quantity    int     `json:"quantity" validate:"required,gt=0"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"`
}

// ============================================================================
// RESPONSE TYPES
// ============================================================================

// SuccessResponse - Standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse - Standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
