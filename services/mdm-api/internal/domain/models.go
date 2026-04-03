package domain

import "time"

// ValidateEntityRequest defines the input payload for entity normalization.
type ValidateEntityRequest struct {
	EntityType string                 `json:"entity_type"`
	Data       map[string]interface{} `json:"data"`
}

// ValidateEntityResult represents standardized and validated output.
type ValidateEntityResult struct {
	EntityType   string                 `json:"entity_type"`
	Standardized map[string]interface{} `json:"standardized"`
	ValidatedAt  time.Time              `json:"validated_at"`
}

// SuccessResponse is a standard success response envelope.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is a standard error response envelope.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
