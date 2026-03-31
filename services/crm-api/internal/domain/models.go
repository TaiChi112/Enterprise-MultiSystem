package domain

import (
	"time"
)

// ============================================================================
// CUSTOMER - Customer profile for CRM
// ============================================================================
type Customer struct {
	ID            int       `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Email         string    `db:"email" json:"email"`
	Phone         *string   `db:"phone" json:"phone"`
	LoyaltyPoints int       `db:"loyalty_points" json:"loyalty_points"`
	IsMember      bool      `db:"is_member" json:"is_member"`
	IsActive      bool      `db:"is_active" json:"is_active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// CreateCustomerRequest - Input DTO for creating a customer
type CreateCustomerRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"omitempty,max=20"`
	IsMember bool   `json:"is_member"`
}

// UpdateCustomerRequest - Input DTO for updating a customer
type UpdateCustomerRequest struct {
	Name     *string `json:"name" validate:"omitempty,min=1,max=255"`
	Email    *string `json:"email" validate:"omitempty,email"`
	Phone    *string `json:"phone" validate:"omitempty,max=20"`
	IsMember *bool   `json:"is_member"`
	IsActive *bool   `json:"is_active"`
}

// AwardLoyaltyPointsRequest - Input DTO for awarding loyalty points
type AwardLoyaltyPointsRequest struct {
	Points int `json:"points" validate:"required,gt=0"`
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
