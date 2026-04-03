package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/oms-api/internal/domain"
)

// ============================================================================
// ORDER LIFECYCLE REPOSITORY
// ============================================================================

// InitializeOrder creates a new order with "pending" status
func (d *Database) InitializeOrder(ctx context.Context, orderLC *domain.OrderLifecycle) error {
	query := `
		INSERT INTO order_lifecycle (order_number, customer_id, status, total_amount, description, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		orderLC.OrderNumber,
		orderLC.CustomerID,
		orderLC.Status,
		orderLC.TotalAmount,
		orderLC.Description,
		orderLC.IsActive,
	).Scan(&orderLC.ID, &orderLC.CreatedAt, &orderLC.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to initialize order: %w", err)
	}

	return nil
}

// GetOrderByID retrieves an order by ID
func (d *Database) GetOrderByID(ctx context.Context, id int) (*domain.OrderLifecycle, error) {
	query := `
		SELECT id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
		FROM order_lifecycle
		WHERE id = $1
	`

	order := &domain.OrderLifecycle{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.Description,
		&order.IsActive,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetOrderByOrderNumber retrieves an order by order number
func (d *Database) GetOrderByOrderNumber(ctx context.Context, orderNumber string) (*domain.OrderLifecycle, error) {
	query := `
		SELECT id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
		FROM order_lifecycle
		WHERE order_number = $1
	`

	order := &domain.OrderLifecycle{}
	err := d.QueryRowContext(ctx, query, orderNumber).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.Description,
		&order.IsActive,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return order, nil
}

// GetOrdersByCustomerID retrieves all orders for a customer
func (d *Database) GetOrdersByCustomerID(ctx context.Context, customerID int) ([]*domain.OrderLifecycle, error) {
	query := `
		SELECT id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
		FROM order_lifecycle
		WHERE customer_id = $1
		ORDER BY created_at DESC
	`

	rows, err := d.QueryContext(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.OrderLifecycle
	for rows.Next() {
		order := &domain.OrderLifecycle{}
		err = rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.CustomerID,
			&order.Status,
			&order.TotalAmount,
			&order.Description,
			&order.IsActive,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// GetAllOrders retrieves all orders with pagination and optional status filter
func (d *Database) GetAllOrders(ctx context.Context, limit, offset int, status *string) ([]*domain.OrderLifecycle, error) {
	var query string
	var args []interface{}

	if status != nil && *status != "" {
		query = `
			SELECT id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
			FROM order_lifecycle
			WHERE status = $1 AND is_active = true
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{*status, limit, offset}
	} else {
		query = `
			SELECT id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
			FROM order_lifecycle
			WHERE is_active = true
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := d.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.OrderLifecycle
	for rows.Next() {
		order := &domain.OrderLifecycle{}
		err = rows.Scan(
			&order.ID,
			&order.OrderNumber,
			&order.CustomerID,
			&order.Status,
			&order.TotalAmount,
			&order.Description,
			&order.IsActive,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (d *Database) UpdateOrderStatus(ctx context.Context, id int, status string) (*domain.OrderLifecycle, error) {
	// Get current order first
	order, err := d.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE order_lifecycle
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, order_number, customer_id, status, total_amount, description, is_active, created_at, updated_at
	`

	err = d.QueryRowContext(ctx, query, status, id).Scan(
		&order.ID,
		&order.OrderNumber,
		&order.CustomerID,
		&order.Status,
		&order.TotalAmount,
		&order.Description,
		&order.IsActive,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to update order status: %w", err)
	}

	return order, nil
}

// DeleteOrder soft-deletes an order by marking as inactive
func (d *Database) DeleteOrder(ctx context.Context, id int) error {
	query := `
		UPDATE order_lifecycle
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	res, err := d.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete order: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// UpdateOrderTotalAmount updates the total amount of an order
func (d *Database) UpdateOrderTotalAmount(ctx context.Context, id int, amount float64) error {
	query := `
		UPDATE order_lifecycle
		SET total_amount = $1, updated_at = NOW()
		WHERE id = $2
	`

	res, err := d.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to update order amount: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// ============================================================================
// ORDER ITEM REPOSITORY
// ============================================================================

// AddOrderItem adds an item to an order
func (d *Database) AddOrderItem(ctx context.Context, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_item_oms (order_id, product_id, product_name, quantity, unit_price, line_total)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		item.OrderID,
		item.ProductID,
		item.ProductName,
		item.Quantity,
		item.UnitPrice,
		item.LineTotal,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to add order item: %w", err)
	}

	return nil
}

// GetOrderItems retrieves all items for an order
func (d *Database) GetOrderItems(ctx context.Context, orderID int) ([]*domain.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, line_total, created_at, updated_at
		FROM order_item_oms
		WHERE order_id = $1
		ORDER BY created_at ASC
	`

	rows, err := d.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		item := &domain.OrderItem{}
		err = rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.LineTotal,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

// DeleteOrderItem removes an item from an order
func (d *Database) DeleteOrderItem(ctx context.Context, itemID int) error {
	query := `
		DELETE FROM order_item_oms
		WHERE id = $1
	`

	res, err := d.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to delete order item: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order item not found")
	}

	return nil
}
