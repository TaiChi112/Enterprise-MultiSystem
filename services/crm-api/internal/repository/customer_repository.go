package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/crm-api/internal/domain"
)

// ============================================================================
// CUSTOMER REPOSITORY
// ============================================================================

// CreateCustomer inserts a new customer into the database
func (d *Database) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	query := `
		INSERT INTO customer (name, email, phone, loyalty_points, is_member, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		customer.Name,
		customer.Email,
		customer.Phone,
		customer.LoyaltyPoints,
		customer.IsMember,
		customer.IsActive,
	).Scan(&customer.ID, &customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetCustomerByID retrieves a customer by ID
func (d *Database) GetCustomerByID(ctx context.Context, id int) (*domain.Customer, error) {
	query := `
		SELECT id, name, email, phone, loyalty_points, is_member, is_active, created_at, updated_at
		FROM customer
		WHERE id = $1
	`

	customer := &domain.Customer{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&customer.ID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.LoyaltyPoints,
		&customer.IsMember,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// GetAllCustomers retrieves all active customers with pagination
func (d *Database) GetAllCustomers(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	query := `
		SELECT id, name, email, phone, loyalty_points, is_member, is_active, created_at, updated_at
		FROM customer
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}
	defer rows.Close()

	var customers []*domain.Customer
	for rows.Next() {
		customer := &domain.Customer{}
		err = rows.Scan(
			&customer.ID,
			&customer.Name,
			&customer.Email,
			&customer.Phone,
			&customer.LoyaltyPoints,
			&customer.IsMember,
			&customer.IsActive,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, nil
}

// UpdateCustomer updates an existing customer
func (d *Database) UpdateCustomer(ctx context.Context, id int, updates *domain.UpdateCustomerRequest) (*domain.Customer, error) {
	// Get current customer
	customer, err := d.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Name != nil {
		customer.Name = *updates.Name
	}
	if updates.Email != nil {
		customer.Email = *updates.Email
	}
	if updates.Phone != nil {
		customer.Phone = updates.Phone
	}
	if updates.IsMember != nil {
		customer.IsMember = *updates.IsMember
	}
	if updates.IsActive != nil {
		customer.IsActive = *updates.IsActive
	}

	query := `
		UPDATE customer
		SET name = $1, email = $2, phone = $3, is_member = $4, is_active = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING created_at, updated_at
	`

	err = d.QueryRowContext(ctx, query,
		customer.Name,
		customer.Email,
		customer.Phone,
		customer.IsMember,
		customer.IsActive,
		id,
	).Scan(&customer.CreatedAt, &customer.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	return customer, nil
}

// DeleteCustomer soft-deletes a customer by marking as inactive
func (d *Database) DeleteCustomer(ctx context.Context, id int) error {
	query := `
		UPDATE customer
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	res, err := d.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

// GetCustomerByEmail retrieves a customer by email
func (d *Database) GetCustomerByEmail(ctx context.Context, email string) (*domain.Customer, error) {
	query := `
		SELECT id, name, email, phone, loyalty_points, is_member, is_active, created_at, updated_at
		FROM customer
		WHERE email = $1
	`

	customer := &domain.Customer{}
	err := d.QueryRowContext(ctx, query, email).Scan(
		&customer.ID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.LoyaltyPoints,
		&customer.IsMember,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// AwardLoyaltyPoints adds loyalty points to a customer
func (d *Database) AwardLoyaltyPoints(ctx context.Context, customerID int, points int) (*domain.Customer, error) {
	query := `
		UPDATE customer
		SET loyalty_points = loyalty_points + $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, name, email, phone, loyalty_points, is_member, is_active, created_at, updated_at
	`

	customer := &domain.Customer{}
	err := d.QueryRowContext(ctx, query, points, customerID).Scan(
		&customer.ID,
		&customer.Name,
		&customer.Email,
		&customer.Phone,
		&customer.LoyaltyPoints,
		&customer.IsMember,
		&customer.IsActive,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to award loyalty points: %w", err)
	}

	return customer, nil
}

// GetCustomerCount retrieves the total count of active customers
func (d *Database) GetCustomerCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM customer WHERE is_active = true`

	var count int
	err := d.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get customer count: %w", err)
	}

	return count, nil
}
