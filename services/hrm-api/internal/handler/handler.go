package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/user/pos-wms-mvp/services/hrm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/hrm-api/internal/service"
)

// Handler holds the service dependency.
type Handler struct {
	service *service.Service
}

// NewHandler creates a new handler instance.
func NewHandler(svc *service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes registers all API routes.
func (h *Handler) RegisterRoutes(app *fiber.App) {
	employees := app.Group("/hrm/employees")
	employees.Post("/", h.CreateEmployee)
	employees.Get("/:id", h.GetEmployee)
	employees.Get("", h.GetAllEmployees)
	employees.Put("/:id", h.UpdateEmployee)
	employees.Delete("/:id", h.DeleteEmployee)

	app.Get("/hrm/payroll", h.GetPayrollSummary)
	app.Get("/api/health", h.HealthCheck)
}

// ============================================================================
// EMPLOYEE HANDLERS
// ============================================================================

// CreateEmployee creates a new employee.
// @Summary Create a new employee
// @Tags Employees
// @Accept json
// @Produce json
// @Param request body domain.CreateEmployeeRequest true "Employee data"
// @Success 201 {object} domain.SuccessResponse{data=domain.Employee}
// @Failure 400 {object} domain.ErrorResponse
// @Router /hrm/employees [post]
func (h *Handler) CreateEmployee(c *fiber.Ctx) error {
	req := &domain.CreateEmployeeRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	employee, err := h.service.CreateEmployee(c.Context(), req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()

		if errors.Is(err, service.ErrEmployeeEmailExists) {
			statusCode = http.StatusBadRequest
		} else if errors.Is(err, service.ErrInvalidSalary) {
			statusCode = http.StatusBadRequest
		}

		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errorMessage,
		})
	}

	return c.Status(http.StatusCreated).JSON(domain.SuccessResponse{
		Success: true,
		Data:    employee,
		Message: "employee created successfully",
	})
}

// GetEmployee retrieves an employee by ID.
// @Summary Get employee by ID
// @Tags Employees
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} domain.SuccessResponse{data=domain.Employee}
// @Failure 404 {object} domain.ErrorResponse
// @Router /hrm/employees/{id} [get]
func (h *Handler) GetEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid employee ID",
		})
	}

	employee, err := h.service.GetEmployee(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "employee not found",
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    employee,
	})
}

// GetAllEmployees retrieves all active employees with pagination.
// @Summary Get all employees
// @Tags Employees
// @Produce json
// @Param limit query int false "Limit (default 20)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} domain.SuccessResponse{data=[]domain.Employee}
// @Router /hrm/employees [get]
func (h *Handler) GetAllEmployees(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	employees, err := h.service.GetAllEmployees(c.Context(), limit, offset)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	if employees == nil {
		employees = []*domain.Employee{}
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    employees,
	})
}

// UpdateEmployee updates an employee.
// @Summary Update employee
// @Tags Employees
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Param request body domain.UpdateEmployeeRequest true "Updated employee data"
// @Success 200 {object} domain.SuccessResponse{data=domain.Employee}
// @Failure 400,404 {object} domain.ErrorResponse
// @Router /hrm/employees/{id} [put]
func (h *Handler) UpdateEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid employee ID",
		})
	}

	req := &domain.UpdateEmployeeRequest{}
	if err := c.BodyParser(req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   fmt.Sprintf("invalid request body: %v", err),
		})
	}

	employee, err := h.service.UpdateEmployee(c.Context(), id, req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorMessage := err.Error()

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		} else if errors.Is(err, service.ErrInvalidSalary) {
			statusCode = http.StatusBadRequest
		}

		return c.Status(statusCode).JSON(domain.ErrorResponse{
			Success: false,
			Error:   errorMessage,
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    employee,
		Message: "employee updated successfully",
	})
}

// DeleteEmployee deletes an employee (soft delete).
// @Summary Delete employee
// @Tags Employees
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} domain.SuccessResponse
// @Failure 404 {object} domain.ErrorResponse
// @Router /hrm/employees/{id} [delete]
func (h *Handler) DeleteEmployee(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(domain.ErrorResponse{
			Success: false,
			Error:   "invalid employee ID",
		})
	}

	err = h.service.DeleteEmployee(c.Context(), id)
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Message: "employee deleted successfully",
	})
}

// ============================================================================
// PAYROLL HANDLERS
// ============================================================================

// GetPayrollSummary retrieves payroll summary for all active employees.
// @Summary Get payroll summary
// @Tags Payroll
// @Produce json
// @Success 200 {object} domain.SuccessResponse{data=domain.PayrollSummary}
// @Router /hrm/payroll [get]
func (h *Handler) GetPayrollSummary(c *fiber.Ctx) error {
	summary, err := h.service.GetPayrollSummary(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(domain.ErrorResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(domain.SuccessResponse{
		Success: true,
		Data:    summary,
	})
}

// ============================================================================
// HEALTH CHECK
// ============================================================================

// HealthCheck returns the health status of the service.
func (h *Handler) HealthCheck(c *fiber.Ctx) error {
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"service": "hrm-api",
	})
}
