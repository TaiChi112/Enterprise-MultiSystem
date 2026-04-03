package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/user/pos-wms-mvp/services/pos-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/pos-api/internal/repository"
)

// Service encapsulates business logic
type Service struct {
	db *repository.Database
}

// NewService creates a new service instance
func NewService(db *repository.Database) *Service {
	return &Service{db: db}
}

// ============================================================================
// PRODUCT SERVICE
// ============================================================================

// CreateProduct creates a new product
func (s *Service) CreateProduct(ctx context.Context, req *domain.CreateProductRequest) (*domain.Product, error) {
	product := &domain.Product{
		SKU:         req.SKU,
		Name:        req.Name,
		Description: &req.Description,
		Price:       req.Price,
		Cost:        req.Cost,
		IsActive:    true,
	}

	if err := s.db.CreateProduct(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

// GetProduct retrieves a product by ID
func (s *Service) GetProduct(ctx context.Context, id int) (*domain.Product, error) {
	return s.db.GetProductByID(ctx, id)
}

// GetAllProducts retrieves all products with pagination
func (s *Service) GetAllProducts(ctx context.Context, limit int, offset int) ([]*domain.Product, error) {
	return s.db.GetAllProducts(ctx, limit, offset)
}

// ============================================================================
// BRANCH SERVICE
// ============================================================================

// CreateBranch creates a new branch
func (s *Service) CreateBranch(ctx context.Context, req *domain.CreateBranchRequest) (*domain.Branch, error) {
	branch := &domain.Branch{
		Name:     req.Name,
		Address:  &req.Address,
		Phone:    &req.Phone,
		IsActive: true,
	}

	if err := s.db.CreateBranch(ctx, branch); err != nil {
		return nil, err
	}

	return branch, nil
}

// GetBranch retrieves a branch by ID
func (s *Service) GetBranch(ctx context.Context, id int) (*domain.Branch, error) {
	return s.db.GetBranchByID(ctx, id)
}

// GetAllBranches retrieves all branches
func (s *Service) GetAllBranches(ctx context.Context) ([]*domain.Branch, error) {
	return s.db.GetAllBranches(ctx)
}

// ============================================================================
// INVENTORY SERVICE
// ============================================================================

// GetInventory retrieves inventory for a product at a specific branch
func (s *Service) GetInventory(ctx context.Context, productID int, branchID int) (*domain.Inventory, error) {
	return s.db.GetInventoryByProductAndBranch(ctx, productID, branchID)
}

// CreateInventory creates a new inventory record for a product at a branch
func (s *Service) CreateInventory(ctx context.Context, req *domain.CreateInventoryRequest) (*domain.Inventory, error) {
	inv := &domain.Inventory{
		ProductID:  req.ProductID,
		BranchID:   req.BranchID,
		Quantity:   req.Quantity,
		MinimumQty: req.MinimumQty,
	}

	if err := s.db.CreateInventory(ctx, inv); err != nil {
		if strings.Contains(err.Error(), "inventory_unique_product_branch") {
			return nil, fmt.Errorf("inventory already exists for product %d at branch %d", req.ProductID, req.BranchID)
		}
		return nil, err
	}

	return inv, nil
}

// GetBranchInventory retrieves all inventory for a branch
func (s *Service) GetBranchInventory(ctx context.Context, branchID int, limit int, offset int) ([]*domain.Inventory, error) {
	return s.db.GetInventoryByBranch(ctx, branchID, limit, offset)
}

// GetLowStockItems retrieves low stock items for a branch
func (s *Service) GetLowStockItems(ctx context.Context, branchID int) ([]*domain.Inventory, error) {
	return s.db.GetLowStockInventory(ctx, branchID)
}

// ============================================================================
// SALE PROCESSING SERVICE - WITH TRANSACTION (ACID)
// ============================================================================

// ProcessSale processes a complete sale transaction with ACID properties:
// 1. Creates an Order
// 2. Creates OrderItems
// 3. Deducts Inventory
// If any step fails, the entire transaction is rolled back
func (s *Service) ProcessSale(ctx context.Context, req *domain.ProcessSaleRequest) (*domain.Order, error) {
	// Start transaction
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of error
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				fmt.Printf("rollback error: %v\n", rollbackErr)
			}
		}
	}()

	// Step 1: Validate branch exists
	_, err = s.db.GetBranchByID(ctx, req.BranchID)
	if err != nil {
		return nil, fmt.Errorf("branch validation failed: %w", err)
	}

	// Step 2: Calculate total amount and validate all items exist with sufficient stock
	totalAmount := float64(0)
	var productData []*struct {
		product   *domain.Product
		inventory *domain.Inventory
		quantity  int
		discount  float64
	}

	for _, item := range req.Items {
		// Get product
		product, err := s.db.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("product %d not found: %w", item.ProductID, err)
		}

		// Get inventory for this branch
		inv, err := s.db.GetInventoryByProductAndBranch(ctx, item.ProductID, req.BranchID)
		if err != nil {
			return nil, fmt.Errorf("inventory for product %d not found at branch: %w", item.ProductID, err)
		}

		// Check sufficient stock
		if inv.Quantity < item.Quantity {
			return nil, fmt.Errorf("insufficient stock for product %d (available: %d, requested: %d)",
				item.ProductID, inv.Quantity, item.Quantity)
		}

		// Calculate line total: (price * quantity) - discount
		lineTotal := (product.Price * float64(item.Quantity)) - item.Discount
		totalAmount += lineTotal

		productData = append(productData, &struct {
			product   *domain.Product
			inventory *domain.Inventory
			quantity  int
			discount  float64
		}{
			product:   product,
			inventory: inv,
			quantity:  item.Quantity,
			discount:  item.Discount,
		})
	}

	// Step 3: Create Order
	order := &domain.Order{
		BranchID:     req.BranchID,
		CustomerName: &req.CustomerName,
		TotalAmount:  totalAmount,
		Status:       domain.OrderStatusCompleted, // For MVP, default to completed
	}

	// Use transaction-aware create (would need txCreateOrder method in a real impl)
	// For now, we'll use the regular db methods within the transaction
	query := `
		INSERT INTO "order" (branch_id, customer_name, total_amount, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err = tx.QueryRowContext(ctx, query,
		order.BranchID,
		order.CustomerName,
		order.TotalAmount,
		order.Status,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Step 4: Create OrderItems and deduct inventory
	for _, pd := range productData {
		// Create OrderItem
		orderItem := &domain.OrderItem{
			OrderID:   order.ID,
			ProductID: pd.product.ID,
			Quantity:  pd.quantity,
			UnitPrice: pd.product.Price,
			Discount:  pd.discount,
		}

		itemQuery := `
			INSERT INTO order_item (order_id, product_id, quantity, unit_price, discount)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at
		`
		err = tx.QueryRowContext(ctx, itemQuery,
			orderItem.OrderID,
			orderItem.ProductID,
			orderItem.Quantity,
			orderItem.UnitPrice,
			orderItem.Discount,
		).Scan(&orderItem.ID, &orderItem.CreatedAt, &orderItem.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Deduct inventory
		invQuery := `
			UPDATE inventory
			SET quantity = quantity - $1
			WHERE id = $2 AND quantity >= $1
			RETURNING quantity
		`
		var newQuantity int
		err = tx.QueryRowContext(ctx, invQuery, pd.quantity, pd.inventory.ID).Scan(&newQuantity)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("inventory deduction failed for product %d (concurrent modification)", pd.product.ID)
			}
			return nil, fmt.Errorf("failed to deduct inventory: %w", err)
		}

		order.OrderItems = append(order.OrderItems, orderItem)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return order, nil
}

// ============================================================================
// ORDER SERVICE
// ============================================================================

// GetOrder retrieves an order by ID with order items
func (s *Service) GetOrder(ctx context.Context, id int) (*domain.Order, error) {
	return s.db.GetOrderByID(ctx, id, true)
}

// GetOrders retrieves orders for a specific branch
func (s *Service) GetOrders(ctx context.Context, branchID int, limit int, offset int) ([]*domain.Order, error) {
	return s.db.GetOrdersByBranch(ctx, branchID, limit, offset)
}
