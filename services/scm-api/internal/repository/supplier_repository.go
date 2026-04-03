package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/scm-api/internal/domain"
)

func (d *Database) CreateSupplier(ctx context.Context, supplier *domain.Supplier) error {
	query := `
		INSERT INTO supplier (name, contact)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		supplier.Name,
		supplier.Contact,
	).Scan(&supplier.ID, &supplier.CreatedAt, &supplier.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create supplier: %w", err)
	}

	return nil
}

func (d *Database) GetSupplierByID(ctx context.Context, id int) (*domain.Supplier, error) {
	query := `
		SELECT id, name, contact, created_at, updated_at
		FROM supplier
		WHERE id = $1
	`

	supplier := &domain.Supplier{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&supplier.ID,
		&supplier.Name,
		&supplier.Contact,
		&supplier.CreatedAt,
		&supplier.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}

	return supplier, nil
}

func (d *Database) GetSuppliers(ctx context.Context, limit, offset int) ([]*domain.Supplier, error) {
	query := `
		SELECT id, name, contact, created_at, updated_at
		FROM supplier
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get suppliers: %w", err)
	}
	defer rows.Close()

	var suppliers []*domain.Supplier
	for rows.Next() {
		supplier := &domain.Supplier{}
		err = rows.Scan(
			&supplier.ID,
			&supplier.Name,
			&supplier.Contact,
			&supplier.CreatedAt,
			&supplier.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan supplier: %w", err)
		}
		suppliers = append(suppliers, supplier)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating suppliers: %w", err)
	}

	return suppliers, nil
}

func (d *Database) UpdateSupplier(ctx context.Context, id int, updates *domain.UpdateSupplierRequest) (*domain.Supplier, error) {
	supplier, err := d.GetSupplierByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if updates.Name != nil {
		supplier.Name = *updates.Name
	}
	if updates.Contact != nil {
		supplier.Contact = updates.Contact
	}

	query := `
		UPDATE supplier
		SET name = $1, contact = $2
		WHERE id = $3
		RETURNING updated_at
	`

	err = d.QueryRowContext(ctx, query,
		supplier.Name,
		supplier.Contact,
		id,
	).Scan(&supplier.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("supplier not found")
		}
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

func (d *Database) DeleteSupplier(ctx context.Context, id int) error {
	query := `
		DELETE FROM supplier
		WHERE id = $1
	`

	res, err := d.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete supplier: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("supplier not found")
	}

	return nil
}
