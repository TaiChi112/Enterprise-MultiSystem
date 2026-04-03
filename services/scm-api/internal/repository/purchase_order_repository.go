package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/scm-api/internal/domain"
)

func (d *Database) CreatePurchaseOrder(ctx context.Context, po *domain.PurchaseOrder) error {
	query := `
		INSERT INTO purchase_order (supplier_id, product_id, quantity, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		po.SupplierID,
		po.ProductID,
		po.Quantity,
		po.Status,
	).Scan(&po.ID, &po.CreatedAt, &po.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create purchase order: %w", err)
	}

	return nil
}

func (d *Database) UpdatePurchaseOrderStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE purchase_order
		SET status = $1
		WHERE id = $2
	`

	res, err := d.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update purchase order status: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("purchase order not found")
	}

	return nil
}

func (d *Database) GetPurchaseOrderByID(ctx context.Context, id int) (*domain.PurchaseOrder, error) {
	query := `
		SELECT id, supplier_id, product_id, quantity, status, created_at, updated_at
		FROM purchase_order
		WHERE id = $1
	`

	po := &domain.PurchaseOrder{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&po.ID,
		&po.SupplierID,
		&po.ProductID,
		&po.Quantity,
		&po.Status,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found")
		}
		return nil, fmt.Errorf("failed to get purchase order: %w", err)
	}

	return po, nil
}

func (d *Database) GetPurchaseOrders(ctx context.Context, limit, offset int) ([]*domain.PurchaseOrder, error) {
	query := `
		SELECT
			po.id,
			po.supplier_id,
			po.product_id,
			po.quantity,
			po.status,
			COALESCE(p.cost, 0) AS unit_cost,
			COALESCE(p.cost, 0) * po.quantity AS total_cost,
			po.created_at,
			po.updated_at
		FROM purchase_order po
		LEFT JOIN product p ON p.id = po.product_id
		ORDER BY po.created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase orders: %w", err)
	}
	defer rows.Close()

	var orders []*domain.PurchaseOrder
	for rows.Next() {
		po := &domain.PurchaseOrder{}
		err = rows.Scan(
			&po.ID,
			&po.SupplierID,
			&po.ProductID,
			&po.Quantity,
			&po.Status,
			&po.UnitCost,
			&po.TotalCost,
			&po.CreatedAt,
			&po.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan purchase order: %w", err)
		}
		orders = append(orders, po)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating purchase orders: %w", err)
	}

	return orders, nil
}
