package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/internal/domain"
)

// ============================================================================
// ORDER REPOSITORY
// ============================================================================

// CreateOrder inserts a new order into the database
// Note: This should be used in a transaction context
func (d *Database) CreateOrder(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO "order" (branch_id, customer_name, total_amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		order.BranchID,
		order.CustomerName,
		order.TotalAmount,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order: %w", err)
	}

	return nil
}

// GetOrderByID retrieves an order by ID with optional order items
func (d *Database) GetOrderByID(ctx context.Context, id int, includeItems bool) (*domain.Order, error) {
	query := `
		SELECT id, branch_id, customer_name, total_amount, status, created_at, updated_at
		FROM "order"
		WHERE id = $1
	`

	order := &domain.Order{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&order.ID,
		&order.BranchID,
		&order.CustomerName,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if includeItems {
		items, err := d.GetOrderItems(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("failed to get order items: %w", err)
		}
		order.OrderItems = items
	}

	return order, nil
}

// GetOrdersByBranch retrieves orders for a specific branch
func (d *Database) GetOrdersByBranch(ctx context.Context, branchID int, limit int, offset int) ([]*domain.Order, error) {
	query := `
		SELECT id, branch_id, customer_name, total_amount, status, created_at, updated_at
		FROM "order"
		WHERE branch_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := d.QueryContext(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order := &domain.Order{}
		err := rows.Scan(
			&order.ID,
			&order.BranchID,
			&order.CustomerName,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, nil
}

// UpdateOrderStatus updates the status of an order
func (d *Database) UpdateOrderStatus(ctx context.Context, orderID int, status string) error {
	query := `
		UPDATE "order"
		SET status = $1
		WHERE id = $2
		RETURNING updated_at
	`

	var updatedAt sql.NullTime
	err := d.QueryRowContext(ctx, query, status, orderID).Scan(&updatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("order not found: %w", err)
		}
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// ============================================================================
// ORDER_ITEM REPOSITORY
// ============================================================================

// CreateOrderItem inserts a new order item into the database
func (d *Database) CreateOrderItem(ctx context.Context, item *domain.OrderItem) error {
	query := `
		INSERT INTO order_item (order_id, product_id, quantity, unit_price, discount)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		item.OrderID,
		item.ProductID,
		item.Quantity,
		item.UnitPrice,
		item.Discount,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}

	return nil
}

// GetOrderItems retrieves all items for a specific order
func (d *Database) GetOrderItems(ctx context.Context, orderID int) ([]*domain.OrderItem, error) {
	query := `
		SELECT id, order_id, product_id, quantity, unit_price, discount, created_at, updated_at
		FROM order_item
		WHERE order_id = $1
		ORDER BY id ASC
	`

	rows, err := d.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []*domain.OrderItem
	for rows.Next() {
		item := &domain.OrderItem{}
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductID,
			&item.Quantity,
			&item.UnitPrice,
			&item.Discount,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}

// ============================================================================
// BRANCH REPOSITORY
// ============================================================================

// GetBranchByID retrieves a branch by ID
func (d *Database) GetBranchByID(ctx context.Context, id int) (*domain.Branch, error) {
	query := `
		SELECT id, name, address, phone, is_active, created_at, updated_at
		FROM branch
		WHERE id = $1
	`

	branch := &domain.Branch{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&branch.ID,
		&branch.Name,
		&branch.Address,
		&branch.Phone,
		&branch.IsActive,
		&branch.CreatedAt,
		&branch.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("branch not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}

	return branch, nil
}

// GetAllBranches retrieves all active branches
func (d *Database) GetAllBranches(ctx context.Context) ([]*domain.Branch, error) {
	query := `
		SELECT id, name, address, phone, is_active, created_at, updated_at
		FROM branch
		WHERE is_active = TRUE
		ORDER BY created_at DESC
	`

	rows, err := d.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}
	defer rows.Close()

	var branches []*domain.Branch
	for rows.Next() {
		branch := &domain.Branch{}
		err := rows.Scan(
			&branch.ID,
			&branch.Name,
			&branch.Address,
			&branch.Phone,
			&branch.IsActive,
			&branch.CreatedAt,
			&branch.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan branch: %w", err)
		}
		branches = append(branches, branch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating branches: %w", err)
	}

	return branches, nil
}

// CreateBranch creates a new branch
func (d *Database) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	query := `
		INSERT INTO branch (name, address, phone, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		branch.Name,
		branch.Address,
		branch.Phone,
		branch.IsActive,
	).Scan(&branch.ID, &branch.CreatedAt, &branch.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create branch: %w", err)
	}

	return nil
}
