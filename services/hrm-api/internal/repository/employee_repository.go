package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/pos-wms-mvp/services/hrm-api/internal/domain"
)

// ============================================================================
// EMPLOYEE REPOSITORY
// ============================================================================

// CreateEmployee inserts a new employee into the database.
func (d *Database) CreateEmployee(ctx context.Context, employee *domain.Employee) error {
	query := `
		INSERT INTO employee (name, email, phone, role, base_salary, department, hire_date, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := d.QueryRowContext(ctx, query,
		employee.Name,
		employee.Email,
		employee.Phone,
		employee.Role,
		employee.BaseSalary,
		employee.Department,
		employee.HireDate,
		employee.IsActive,
	).Scan(&employee.ID, &employee.CreatedAt, &employee.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	return nil
}

// GetEmployeeByID retrieves an employee by ID.
func (d *Database) GetEmployeeByID(ctx context.Context, id int) (*domain.Employee, error) {
	query := `
		SELECT id, name, email, phone, role, base_salary, department, hire_date, is_active, created_at, updated_at
		FROM employee
		WHERE id = $1
	`

	employee := &domain.Employee{}
	err := d.QueryRowContext(ctx, query, id).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Phone,
		&employee.Role,
		&employee.BaseSalary,
		&employee.Department,
		&employee.HireDate,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	return employee, nil
}

// GetEmployeeByEmail retrieves an employee by email.
func (d *Database) GetEmployeeByEmail(ctx context.Context, email string) (*domain.Employee, error) {
	query := `
		SELECT id, name, email, phone, role, base_salary, department, hire_date, is_active, created_at, updated_at
		FROM employee
		WHERE email = $1
	`

	employee := &domain.Employee{}
	err := d.QueryRowContext(ctx, query, email).Scan(
		&employee.ID,
		&employee.Name,
		&employee.Email,
		&employee.Phone,
		&employee.Role,
		&employee.BaseSalary,
		&employee.Department,
		&employee.HireDate,
		&employee.IsActive,
		&employee.CreatedAt,
		&employee.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("employee not found")
		}
		return nil, fmt.Errorf("failed to get employee by email: %w", err)
	}

	return employee, nil
}

// GetAllEmployees retrieves all active employees with pagination.
func (d *Database) GetAllEmployees(ctx context.Context, limit, offset int) ([]*domain.Employee, error) {
	query := `
		SELECT id, name, email, phone, role, base_salary, department, hire_date, is_active, created_at, updated_at
		FROM employee
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := d.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		err := rows.Scan(
			&employee.ID,
			&employee.Name,
			&employee.Email,
			&employee.Phone,
			&employee.Role,
			&employee.BaseSalary,
			&employee.Department,
			&employee.HireDate,
			&employee.IsActive,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating employees: %w", err)
	}

	return employees, nil
}

// UpdateEmployee updates an existing employee.
func (d *Database) UpdateEmployee(ctx context.Context, id int, updates *domain.UpdateEmployeeRequest) (*domain.Employee, error) {
	// Get current employee first
	employee, err := d.GetEmployeeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if updates.Name != nil {
		employee.Name = *updates.Name
	}
	if updates.Email != nil {
		employee.Email = *updates.Email
	}
	if updates.Phone != nil {
		employee.Phone = updates.Phone
	}
	if updates.Role != nil {
		employee.Role = *updates.Role
	}
	if updates.BaseSalary != nil {
		employee.BaseSalary = *updates.BaseSalary
	}
	if updates.Department != nil {
		employee.Department = updates.Department
	}
	if updates.IsActive != nil {
		employee.IsActive = *updates.IsActive
	}

	query := `
		UPDATE employee
		SET name = $1, email = $2, phone = $3, role = $4, base_salary = $5, department = $6, is_active = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING created_at, updated_at
	`

	err = d.QueryRowContext(ctx, query,
		employee.Name,
		employee.Email,
		employee.Phone,
		employee.Role,
		employee.BaseSalary,
		employee.Department,
		employee.IsActive,
		id,
	).Scan(&employee.CreatedAt, &employee.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return employee, nil
}

// DeleteEmployee soft-deletes an employee by marking as inactive.
func (d *Database) DeleteEmployee(ctx context.Context, id int) error {
	query := `
		UPDATE employee
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	res, err := d.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("employee not found")
	}

	return nil
}

// GetAllEmployeesForPayroll retrieves all active employees for payroll aggregation.
func (d *Database) GetAllEmployeesForPayroll(ctx context.Context) ([]*domain.Employee, error) {
	query := `
		SELECT id, name, email, phone, role, base_salary, department, hire_date, is_active, created_at, updated_at
		FROM employee
		WHERE is_active = true
		ORDER BY id ASC
	`

	rows, err := d.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get payroll employees: %w", err)
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{}
		err := rows.Scan(
			&employee.ID,
			&employee.Name,
			&employee.Email,
			&employee.Phone,
			&employee.Role,
			&employee.BaseSalary,
			&employee.Department,
			&employee.HireDate,
			&employee.IsActive,
			&employee.CreatedAt,
			&employee.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, employee)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payroll employees: %w", err)
	}

	return employees, nil
}

// GetEmployeeCount retrieves the total count of active employees.
func (d *Database) GetEmployeeCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM employee WHERE is_active = true`

	var count int
	err := d.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get employee count: %w", err)
	}

	return count, nil
}
