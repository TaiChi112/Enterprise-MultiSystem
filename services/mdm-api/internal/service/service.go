package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/user/pos-wms-mvp/services/mdm-api/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Service contains MDM normalization and validation logic.
type Service struct{}

// NewService creates a new MDM service instance.
func NewService() *Service {
	return &Service{}
}

// ValidateAndStandardizeEntity validates and standardizes customer/supplier data.
func (s *Service) ValidateAndStandardizeEntity(_ context.Context, req *domain.ValidateEntityRequest) (*domain.ValidateEntityResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	entityType := strings.ToLower(strings.TrimSpace(req.EntityType))
	if entityType != "customer" && entityType != "supplier" {
		return nil, fmt.Errorf("entity_type must be either customer or supplier")
	}
	if req.Data == nil {
		return nil, fmt.Errorf("data is required")
	}

	standardized := cloneMap(req.Data)

	standardizeStringField(standardized, "name", toTitleCase)
	standardizeStringField(standardized, "company_name", toTitleCase)
	standardizeStringField(standardized, "contact_name", toTitleCase)
	standardizeStringField(standardized, "address", strings.TrimSpace)

	if err := standardizeEmailField(standardized, "email"); err != nil {
		return nil, err
	}
	if err := standardizeEmailField(standardized, "contact_email"); err != nil {
		return nil, err
	}
	if err := standardizePhoneField(standardized, "phone"); err != nil {
		return nil, err
	}
	if err := standardizePhoneField(standardized, "contact_phone"); err != nil {
		return nil, err
	}

	if err := validateRequired(entityType, standardized); err != nil {
		return nil, err
	}

	return &domain.ValidateEntityResult{
		EntityType:   entityType,
		Standardized: standardized,
		ValidatedAt:  time.Now().UTC(),
	}, nil
}

func validateRequired(entityType string, data map[string]interface{}) error {
	if !hasNonEmptyString(data, "name") && !hasNonEmptyString(data, "company_name") {
		return fmt.Errorf("name or company_name is required")
	}

	if entityType == "customer" {
		if !hasNonEmptyString(data, "email") && !hasNonEmptyString(data, "phone") {
			return fmt.Errorf("customer requires at least one contact field: email or phone")
		}
	}

	if entityType == "supplier" {
		if !hasNonEmptyString(data, "contact_email") && !hasNonEmptyString(data, "contact_phone") {
			return fmt.Errorf("supplier requires at least one contact field: contact_email or contact_phone")
		}
	}

	return nil
}

func hasNonEmptyString(data map[string]interface{}, key string) bool {
	value, ok := data[key]
	if !ok {
		return false
	}
	str, ok := value.(string)
	if !ok {
		return false
	}
	return strings.TrimSpace(str) != ""
}

func standardizeStringField(data map[string]interface{}, key string, transform func(string) string) {
	value, ok := data[key]
	if !ok {
		return
	}
	str, ok := value.(string)
	if !ok {
		return
	}
	data[key] = transform(str)
}

func standardizeEmailField(data map[string]interface{}, key string) error {
	value, ok := data[key]
	if !ok {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", key)
	}
	str = strings.ToLower(strings.TrimSpace(str))
	if str == "" {
		delete(data, key)
		return nil
	}
	if !emailRegex.MatchString(str) {
		return fmt.Errorf("%s format is invalid", key)
	}
	data[key] = str
	return nil
}

func standardizePhoneField(data map[string]interface{}, key string) error {
	value, ok := data[key]
	if !ok {
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("%s must be a string", key)
	}

	trimmed := strings.TrimSpace(str)
	if trimmed == "" {
		delete(data, key)
		return nil
	}

	digits := make([]rune, 0, len(trimmed))
	for _, r := range trimmed {
		if unicode.IsDigit(r) {
			digits = append(digits, r)
		}
	}
	if len(digits) < 8 || len(digits) > 15 {
		return fmt.Errorf("%s format is invalid", key)
	}

	normalizedDigits := string(digits)
	if strings.HasPrefix(normalizedDigits, "00") {
		normalizedDigits = strings.TrimPrefix(normalizedDigits, "00")
	}
	if strings.HasPrefix(trimmed, "+") || strings.HasPrefix(strings.TrimSpace(str), "00") {
		data[key] = "+" + normalizedDigits
		return nil
	}

	data[key] = normalizedDigits
	return nil
}

func cloneMap(src map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(src))
	for k, v := range src {
		out[k] = v
	}
	return out
}

func toTitleCase(in string) string {
	words := strings.Fields(strings.TrimSpace(in))
	if len(words) == 0 {
		return ""
	}
	for i, w := range words {
		lower := strings.ToLower(w)
		runes := []rune(lower)
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
