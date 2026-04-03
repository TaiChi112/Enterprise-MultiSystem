package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/user/pos-wms-mvp/services/iam-api/internal/domain"
)

// AuthService handles login and JWT issuing.
type AuthService struct {
	secret []byte
	issuer string
	ttl    time.Duration
}

// UserCredential defines mock users for STEP 2.
type UserCredential struct {
	Password string
	Role     string
}

var mockUsers = map[string]UserCredential{
	"admin":   {Password: "admin123", Role: "admin"},
	"cashier": {Password: "cashier123", Role: "cashier"},
}

type accessClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(secret, issuer string, ttl time.Duration) *AuthService {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if issuer == "" {
		issuer = "iam-api"
	}
	return &AuthService{
		secret: []byte(secret),
		issuer: issuer,
		ttl:    ttl,
	}
}

// Login validates credentials and returns a signed JWT token.
func (a *AuthService) Login(req *domain.LoginRequest) (*domain.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	cred, ok := mockUsers[req.Username]
	if !ok || cred.Password != req.Password {
		return nil, fmt.Errorf("invalid username or password")
	}

	now := time.Now().UTC()
	expiresAt := now.Add(a.ttl)

	claims := &accessClaims{
		Username: req.Username,
		Role:     cred.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.issuer,
			Subject:   req.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(a.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign jwt: %w", err)
	}

	return &domain.LoginResponse{
		AccessToken: signedToken,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User: domain.UserInfo{
			Username: req.Username,
			Role:     cred.Role,
		},
	}, nil
}
