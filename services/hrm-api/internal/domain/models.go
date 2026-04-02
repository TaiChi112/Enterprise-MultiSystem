package domain

import "time"

// ============================================================================
// EMPLOYEE - Human Resource Profile
// ============================================================================

const (
	EmployeeRoleManager  = "manager"
	EmployeeRoleStaff    = "staff"
	EmployeeRoleIntern   = "intern"
	EmployeeRoleExecutor = "executor"
)

// Employee represents an employee profile in HRM.
type Employee struct {
	ID         int       `db:"id" json:"id"`
	Name       string    `db:"name" json:"name"`
	Email      string    `db:"email" json:"email"`
	Phone      *string   `db:"phone" json:"phone"`
	Role       string    `db:"role" json:"role"`
	BaseSalary float64   `db:"base_salary" json:"base_salary"`
	Department *string   `db:"department" json:"department"`
	HireDate   time.Time `db:"hire_date" json:"hire_date"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================================
// REQUEST / RESPONSE DTOs
// ============================================================================

// CreateEmployeeRequest - Input DTO for creating a new employee
type CreateEmployeeRequest struct {
	Name       string  `json:"name" validate:"required,min=1,max=255"`
	Email      string  `json:"email" validate:"required,email"`
	Phone      string  `json:"phone" validate:"omitempty,max=20"`
	Role       string  `json:"role" validate:"required,oneof=manager staff intern executor"`
	BaseSalary float64 `json:"base_salary" validate:"required,gt=0"`
	Department string  `json:"department" validate:"omitempty,max=100"`
	HireDate   string  `json:"hire_date" validate:"required,datetime=2006-01-02"`
}

// UpdateEmployeeRequest - Input DTO for updating an employee
type UpdateEmployeeRequest struct {
	Name       *string  `json:"name" validate:"omitempty,min=1,max=255"`
	Email      *string  `json:"email" validate:"omitempty,email"`
	Phone      *string  `json:"phone" validate:"omitempty,max=20"`
	Role       *string  `json:"role" validate:"omitempty,oneof=manager staff intern executor"`
	BaseSalary *float64 `json:"base_salary" validate:"omitempty,gt=0"`
	Department *string  `json:"department" validate:"omitempty,max=100"`
	IsActive   *bool    `json:"is_active" validate:"omitempty"`
}

// PayrollSummary - Aggregated payroll data for ERP
type PayrollSummary struct {
	TotalActiveSalary   float64     `json:"total_active_salary"`
	TotalEmployeeCount  int         `json:"total_employee_count"`
	ActiveEmployeeCount int         `json:"active_employee_count"`
	Employees           []*Employee `json:"employees,omitempty"`
}

// ============================================================================
// RESPONSE TYPES
// ============================================================================

// SuccessResponse - Standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse - Standard error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
