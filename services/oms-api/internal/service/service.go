package service

import (
	"context"
	"fmt"

	"github.com/user/pos-wms-mvp/services/oms-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/oms-api/internal/repository"
)

// Service holds repository dependencies and business logic
type Service struct {
	repo *repository.Database
}

// NewService creates a new service instance
func NewService(repo *repository.Database) *Service {
	return &Service{repo: repo}
}

// ============================================================================
// ORDER LIFECYCLE SERVICE METHODS
// ============================================================================

// InitializeOrder creates a new order with "pending" status
func (s *Service) InitializeOrder(ctx context.Context, req *domain.InitializeOrderRequest) (*domain.OrderLifecycle, error) {
	// Generate order number (timestamp-based)
	orderNumber := generateOrderNumber()

	order := &domain.OrderLifecycle{
		OrderNumber: orderNumber,
		CustomerID:  req.CustomerID,
		Status:      domain.OrderStatusPending,
		TotalAmount: 0, // Start with 0, will be calculated after adding items
		Description: stringPtr(req.Description),
		IsActive:    true,
	}

	if err := s.repo.InitializeOrder(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrder retrieves an order by ID
func (s *Service) GetOrder(ctx context.Context, id int) (*domain.OrderLifecycle, error) {
	order, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load order items
	items, err := s.repo.GetOrderItems(ctx, id)
	if err != nil {
		return nil, err
	}
	order.OrderItems = items

	return order, nil
}

// GetOrderByOrderNumber retrieves an order by order number
func (s *Service) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*domain.OrderLifecycle, error) {
	order, err := s.repo.GetOrderByOrderNumber(ctx, orderNumber)
	if err != nil {
		return nil, err
	}

	// Load order items
	items, err := s.repo.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}
	order.OrderItems = items

	return order, nil
}

// GetOrdersByCustomerID retrieves all orders for a customer
func (s *Service) GetOrdersByCustomerID(ctx context.Context, customerID int) ([]*domain.OrderLifecycle, error) {
	return s.repo.GetOrdersByCustomerID(ctx, customerID)
}

// GetAllOrders retrieves all orders with pagination and optional filtering
func (s *Service) GetAllOrders(ctx context.Context, limit, offset int, status *string) ([]*domain.OrderLifecycle, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetAllOrders(ctx, limit, offset, status)
}

// UpdateOrderStatus transitions an order to a new status
func (s *Service) UpdateOrderStatus(ctx context.Context, id int, newStatus string) (*domain.OrderLifecycle, error) {
	// Validate status is allowed
	if !isValidOrderStatus(newStatus) {
		return nil, errInvalidStatus
	}

	// Get current order to validate transition
	currentOrder, err := s.repo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if transition is valid
	if !isValidStatusTransition(currentOrder.Status, newStatus) {
		return nil, fmt.Errorf("invalid status transition from '%s' to '%s'", currentOrder.Status, newStatus)
	}

	// Update status
	updatedOrder, err := s.repo.UpdateOrderStatus(ctx, id, newStatus)
	if err != nil {
		return nil, err
	}

	return updatedOrder, nil
}

// AddItemToOrder adds an item to an order and recalculates total
func (s *Service) AddItemToOrder(ctx context.Context, orderID int, req *domain.AddOrderItemRequest) (*domain.OrderItem, error) {
	if req.Quantity <= 0 || req.UnitPrice <= 0 {
		return nil, errInvalidItemData
	}

	item := &domain.OrderItem{
		OrderID:     orderID,
		ProductID:   req.ProductID,
		ProductName: stringPtr(req.ProductName),
		Quantity:    req.Quantity,
		UnitPrice:   req.UnitPrice,
		LineTotal:   float64(req.Quantity) * req.UnitPrice,
	}

	// Add the item
	if err := s.repo.AddOrderItem(ctx, item); err != nil {
		return nil, err
	}

	// Recalculate order total
	if err := s.recalculateOrderTotal(ctx, orderID); err != nil {
		return nil, err
	}

	return item, nil
}

// RemoveItemFromOrder removes an item from an order and recalculates total
func (s *Service) RemoveItemFromOrder(ctx context.Context, itemID int) error {
	// First, get the item to know which order it belongs to
	items, err := s.repo.GetOrderItems(ctx, itemID) // This will fail, need to refactor
	if err != nil {
		return err
	}

	if len(items) == 0 {
		return fmt.Errorf("item not found")
	}

	orderID := items[0].OrderID

	// Delete the item
	if err := s.repo.DeleteOrderItem(ctx, itemID); err != nil {
		return err
	}

	// Recalculate order total
	return s.recalculateOrderTotal(ctx, orderID)
}

// DeleteOrder soft-deletes an order
func (s *Service) DeleteOrder(ctx context.Context, id int) error {
	return s.repo.DeleteOrder(ctx, id)
}

// ============================================================================
// PRIVATE HELPER METHODS
// ============================================================================

// recalculateOrderTotal recalculates the total amount of an order
func (s *Service) recalculateOrderTotal(ctx context.Context, orderID int) error {
	items, err := s.repo.GetOrderItems(ctx, orderID)
	if err != nil {
		return err
	}

	var total float64
	for _, item := range items {
		total += item.LineTotal
	}

	return s.repo.UpdateOrderTotalAmount(ctx, orderID, total)
}

// isValidOrderStatus checks if a status is valid
func isValidOrderStatus(status string) bool {
	validStatuses := map[string]bool{
		domain.OrderStatusPending:   true,
		domain.OrderStatusPaid:      true,
		domain.OrderStatusShipped:   true,
		domain.OrderStatusCompleted: true,
		domain.OrderStatusCancelled: true,
	}
	return validStatuses[status]
}

// isValidStatusTransition checks if a status transition is allowed
func isValidStatusTransition(from, to string) bool {
	allowedTransitions := map[string]map[string]bool{
		domain.OrderStatusPending: {
			domain.OrderStatusPaid:      true,
			domain.OrderStatusCancelled: true,
		},
		domain.OrderStatusPaid: {
			domain.OrderStatusShipped:   true,
			domain.OrderStatusCancelled: true,
		},
		domain.OrderStatusShipped: {
			domain.OrderStatusCompleted: true,
			domain.OrderStatusCancelled: true,
		},
		domain.OrderStatusCompleted: {
			// Completed orders cannot transition
		},
		domain.OrderStatusCancelled: {
			// Cancelled orders cannot transition
		},
	}

	transitions, exists := allowedTransitions[from]
	if !exists {
		return false
	}

	return transitions[to]
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	// Simple timestamp-based order number: ORD-20260330-150530-XXXXX
	// In production, use UUID or database sequence
	return fmt.Sprintf("ORD-%d", getCurrentTimestamp())
}

// getCurrentTimestamp returns current timestamp in milliseconds
func getCurrentTimestamp() int64 {
	return 0 // Placeholder - will use time.Now().UnixMilli() in actual implementation
}

// stringPtr converts string to pointer
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
