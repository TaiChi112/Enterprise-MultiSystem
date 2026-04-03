package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/user/pos-wms-mvp/services/hrm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/hrm-api/internal/repository"
)

// Service holds repository dependencies and business logic.
type Service struct {
	repo *repository.Database
}

// NewService creates a new service instance.
func NewService(repo *repository.Database) *Service {
	return &Service{repo: repo}
}

// ============================================================================
// EMPLOYEE SERVICE METHODS
// ============================================================================

// CreateEmployee handles employee creation business logic.
func (s *Service) CreateEmployee(ctx context.Context, req *domain.CreateEmployeeRequest) (*domain.Employee, error) {
	// Validate salary
	if req.BaseSalary <= 0 {
		return nil, ErrInvalidSalary
	}

	// Parse hire date
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return nil, fmt.Errorf("invalid hire date format: %w", err)
	}

	// Check if email already exists
	existing, err := s.repo.GetEmployeeByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, ErrEmployeeEmailExists
	}

	employee := &domain.Employee{
		Name:       strings.TrimSpace(req.Name),
		Email:      strings.TrimSpace(req.Email),
		Phone:      stringPtr(strings.TrimSpace(req.Phone)),
		Role:       strings.ToLower(strings.TrimSpace(req.Role)),
		BaseSalary: req.BaseSalary,
		Department: stringPtr(strings.TrimSpace(req.Department)),
		HireDate:   hireDate,
		IsActive:   true,
	}

	if err := s.repo.CreateEmployee(ctx, employee); err != nil {
		return nil, err
	}

	return employee, nil
}

// GetEmployee retrieves an employee by ID.
func (s *Service) GetEmployee(ctx context.Context, id int) (*domain.Employee, error) {
	return s.repo.GetEmployeeByID(ctx, id)
}

// GetAllEmployees retrieves all active employees with pagination.
func (s *Service) GetAllEmployees(ctx context.Context, limit, offset int) ([]*domain.Employee, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetAllEmployees(ctx, limit, offset)
}

// UpdateEmployee handles employee update business logic.
func (s *Service) UpdateEmployee(ctx context.Context, id int, req *domain.UpdateEmployeeRequest) (*domain.Employee, error) {
	// Validate salary if provided
	if req.BaseSalary != nil && *req.BaseSalary <= 0 {
		return nil, ErrInvalidSalary
	}

	// Trim string fields if provided
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
	}
	if req.Email != nil {
		email := strings.TrimSpace(*req.Email)
		req.Email = &email
	}
	if req.Role != nil {
		role := strings.ToLower(strings.TrimSpace(*req.Role))
		req.Role = &role
	}
	if req.Department != nil {
		dept := strings.TrimSpace(*req.Department)
		req.Department = &dept
	}

	return s.repo.UpdateEmployee(ctx, id, req)
}

// DeleteEmployee handles employee deletion business logic (soft delete).
func (s *Service) DeleteEmployee(ctx context.Context, id int) error {
	return s.repo.DeleteEmployee(ctx, id)
}

// GetPayrollSummary retrieves payroll summary for all active employees.
func (s *Service) GetPayrollSummary(ctx context.Context) (*domain.PayrollSummary, error) {
	employees, err := s.repo.GetAllEmployeesForPayroll(ctx)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.GetEmployeeCount(ctx)
	if err != nil {
		return nil, err
	}

	var totalActiveSalary float64
	for _, emp := range employees {
		totalActiveSalary += emp.BaseSalary
	}

	return &domain.PayrollSummary{
		TotalActiveSalary:   totalActiveSalary,
		TotalEmployeeCount:  len(employees),
		ActiveEmployeeCount: count,
		Employees:           employees,
	}, nil
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
