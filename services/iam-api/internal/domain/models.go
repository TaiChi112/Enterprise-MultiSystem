package domain

import "time"

// LoginRequest is the payload for user authentication.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserInfo describes authenticated user details returned to clients.
type UserInfo struct {
	Username string `json:"username"`
	Role     string `json:"role"`
}

// LoginResponse holds JWT details after successful authentication.
type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        UserInfo  `json:"user"`
}

// ErrorResponse is the API error shape.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// SuccessResponse is the API success shape.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}
