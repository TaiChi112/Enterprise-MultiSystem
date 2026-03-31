package service

import (
	"context"

	"github.com/user/pos-wms-mvp/services/crm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/crm-api/internal/repository"
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
// CUSTOMER SERVICE METHODS
// ============================================================================

// CreateCustomer handles customer creation business logic
func (s *Service) CreateCustomer(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.Customer, error) {
	customer := &domain.Customer{
		Name:          req.Name,
		Email:         req.Email,
		Phone:         stringPtr(req.Phone),
		LoyaltyPoints: 0, // New customers start with 0 points
		IsMember:      req.IsMember,
		IsActive:      true,
	}

	// Check if email already exists
	existing, err := s.repo.GetCustomerByEmail(ctx, customer.Email)
	if err == nil && existing != nil {
		return nil, errCustomerEmailExists
	}

	if err := s.repo.CreateCustomer(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *Service) GetCustomer(ctx context.Context, id int) (*domain.Customer, error) {
	return s.repo.GetCustomerByID(ctx, id)
}

// GetAllCustomers retrieves all active customers
func (s *Service) GetAllCustomers(ctx context.Context, limit, offset int) ([]*domain.Customer, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetAllCustomers(ctx, limit, offset)
}

// UpdateCustomer handles customer update business logic
func (s *Service) UpdateCustomer(ctx context.Context, id int, req *domain.UpdateCustomerRequest) (*domain.Customer, error) {
	return s.repo.UpdateCustomer(ctx, id, req)
}

// DeleteCustomer handles customer deletion business logic (soft delete)
func (s *Service) DeleteCustomer(ctx context.Context, id int) error {
	return s.repo.DeleteCustomer(ctx, id)
}

// AwardLoyaltyPoints awards loyalty points to a customer
func (s *Service) AwardLoyaltyPoints(ctx context.Context, customerID int, points int) (*domain.Customer, error) {
	if points <= 0 {
		return nil, errInvalidPoints
	}

	return s.repo.AwardLoyaltyPoints(ctx, customerID, points)
}

// GetCustomerCount retrieves the total count of active customers
func (s *Service) GetCustomerCount(ctx context.Context) (int, error) {
	return s.repo.GetCustomerCount(ctx)
}

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
