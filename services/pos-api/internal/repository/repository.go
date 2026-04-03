package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/pos-api/internal/domain"
)

// ============================================================================
// PRODUCT REPOSITORY
// ============================================================================

// CreateProduct inserts a new product into the database
func (d *Database) CreateProduct(ctx context.Context, product *domain.Product) error {
	query := `
		INSERT INTO product (sku, name, description, price, cost, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		product.SKU,
		product.Name,
		product.Description,
		product.Price,
		product.Cost,
		product.IsActive,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// GetProductByID retrieves a product by ID
func (d *Database) GetProductByID(ctx context.Context, id int) (*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, price, cost, is_active, created_at, updated_at
		FROM product
		WHERE id = $1
	`

	product := &domain.Product{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Cost,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

// GetProductBySKU retrieves a product by SKU
func (d *Database) GetProductBySKU(ctx context.Context, sku string) (*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, price, cost, is_active, created_at, updated_at
		FROM product
		WHERE sku = $1
	`

	product := &domain.Product{}
	err := d.QueryRowContext(ctx, query, sku).Scan(
		&product.ID,
		&product.SKU,
		&product.Name,
		&product.Description,
		&product.Price,
		&product.Cost,
		&product.IsActive,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get product by sku: %w", err)
	}

	return product, nil
}

// GetAllProducts retrieves all active products with pagination
func (d *Database) GetAllProducts(ctx context.Context, limit int, offset int) ([]*domain.Product, error) {
	query := `
		SELECT id, sku, name, description, price, cost, is_active, created_at, updated_at
		FROM product
		WHERE is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	var products []*domain.Product
	for rows.Next() {
		product := &domain.Product{}
		err := rows.Scan(
			&product.ID,
			&product.SKU,
			&product.Name,
			&product.Description,
			&product.Price,
			&product.Cost,
			&product.IsActive,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating products: %w", err)
	}

	return products, nil
}

// UpdateProduct updates an existing product
func (d *Database) UpdateProduct(ctx context.Context, product *domain.Product) error {
	query := `
		UPDATE product
		SET sku = $1, name = $2, description = $3, price = $4, cost = $5, is_active = $6
		WHERE id = $7
		RETURNING updated_at
	`

	err := d.QueryRowContext(ctx, query,
		product.SKU,
		product.Name,
		product.Description,
		product.Price,
		product.Cost,
		product.IsActive,
		product.ID,
	).Scan(&product.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("product not found: %w", err)
		}
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// ============================================================================
// INVENTORY REPOSITORY
// ============================================================================

// GetInventoryByProductAndBranch retrieves inventory for a product at a specific branch
func (d *Database) GetInventoryByProductAndBranch(ctx context.Context, productID int, branchID int) (*domain.Inventory, error) {
	query := `
		SELECT id, product_id, branch_id, quantity, minimum_qty, created_at, updated_at
		FROM inventory
		WHERE product_id = $1 AND branch_id = $2
	`

	inv := &domain.Inventory{}
	err := d.QueryRowContext(ctx, query, productID, branchID).Scan(
		&inv.ID,
		&inv.ProductID,
		&inv.BranchID,
		&inv.Quantity,
		&inv.MinimumQty,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("inventory not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}

	return inv, nil
}

// GetInventoryByBranch retrieves all inventory for a specific branch
func (d *Database) GetInventoryByBranch(ctx context.Context, branchID int, limit int, offset int) ([]*domain.Inventory, error) {
	query := `
		SELECT id, product_id, branch_id, quantity, minimum_qty, created_at, updated_at
		FROM inventory
		WHERE branch_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := d.QueryContext(ctx, query, branchID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory by branch: %w", err)
	}
	defer rows.Close()

	var inventoryList []*domain.Inventory
	for rows.Next() {
		inv := &domain.Inventory{}
		err := rows.Scan(
			&inv.ID,
			&inv.ProductID,
			&inv.BranchID,
			&inv.Quantity,
			&inv.MinimumQty,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory: %w", err)
		}
		inventoryList = append(inventoryList, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating inventory: %w", err)
	}

	return inventoryList, nil
}

// CreateInventory creates a new inventory record
func (d *Database) CreateInventory(ctx context.Context, inv *domain.Inventory) error {
	query := `
		INSERT INTO inventory (product_id, branch_id, quantity, minimum_qty)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		inv.ProductID,
		inv.BranchID,
		inv.Quantity,
		inv.MinimumQty,
	).Scan(&inv.ID, &inv.CreatedAt, &inv.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create inventory: %w", err)
	}

	return nil
}

// UpdateInventory updates inventory quantity and minimum_qty
func (d *Database) UpdateInventory(ctx context.Context, inv *domain.Inventory) error {
	query := `
		UPDATE inventory
		SET quantity = $1, minimum_qty = $2
		WHERE id = $3
		RETURNING updated_at
	`

	err := d.QueryRowContext(ctx, query,
		inv.Quantity,
		inv.MinimumQty,
		inv.ID,
	).Scan(&inv.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("inventory not found: %w", err)
		}
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	return nil
}

// AdjustInventoryQuantity adjusts the inventory quantity by a delta (positive or negative)
func (d *Database) AdjustInventoryQuantity(ctx context.Context, inventoryID int, delta int) error {
	query := `
		UPDATE inventory
		SET quantity = quantity + $1
		WHERE id = $2 AND quantity + $1 >= 0
		RETURNING quantity
	`

	var newQuantity int
	err := d.QueryRowContext(ctx, query, delta, inventoryID).Scan(&newQuantity)

	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("inventory not found or adjustment would result in negative quantity: %w", err)
		}
		return fmt.Errorf("failed to adjust inventory: %w", err)
	}

	return nil
}

// GetLowStockInventory retrieves inventory items below minimum quantity
func (d *Database) GetLowStockInventory(ctx context.Context, branchID int) ([]*domain.Inventory, error) {
	query := `
		SELECT id, product_id, branch_id, quantity, minimum_qty, created_at, updated_at
		FROM inventory
		WHERE branch_id = $1 AND quantity <= minimum_qty
		ORDER BY quantity ASC
	`

	rows, err := d.QueryContext(ctx, query, branchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get low stock inventory: %w", err)
	}
	defer rows.Close()

	var inventoryList []*domain.Inventory
	for rows.Next() {
		inv := &domain.Inventory{}
		err := rows.Scan(
			&inv.ID,
			&inv.ProductID,
			&inv.BranchID,
			&inv.Quantity,
			&inv.MinimumQty,
			&inv.CreatedAt,
			&inv.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan inventory: %w", err)
		}
		inventoryList = append(inventoryList, inv)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating low stock inventory: %w", err)
	}

	return inventoryList, nil
}
