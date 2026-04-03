package service

import (
	"context"
	"testing"

	"github.com/user/pos-wms-mvp/services/mdm-api/internal/domain"
)

func TestValidateAndStandardizeEntity_Customer_Success(t *testing.T) {
	svc := NewService()
	req := &domain.ValidateEntityRequest{
		EntityType: "customer",
		Data: map[string]interface{}{
			"name":  "   aNothai   sri  ",
			"email": "  ANOTHAI@EXAMPLE.COM ",
			"phone": " +66 81-234-5678 ",
		},
	}

	result, err := svc.ValidateAndStandardizeEntity(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.EntityType != "customer" {
		t.Fatalf("expected entity type customer, got %s", result.EntityType)
	}

	if result.Standardized["name"] != "Anothai Sri" {
		t.Fatalf("expected normalized name Anothai Sri, got %v", result.Standardized["name"])
	}

	if result.Standardized["email"] != "anothai@example.com" {
		t.Fatalf("expected lowercase email, got %v", result.Standardized["email"])
	}

	if result.Standardized["phone"] != "+66812345678" {
		t.Fatalf("expected normalized phone +66812345678, got %v", result.Standardized["phone"])
	}
}

func TestValidateAndStandardizeEntity_Supplier_InvalidContactEmail(t *testing.T) {
	svc := NewService()
	req := &domain.ValidateEntityRequest{
		EntityType: "supplier",
		Data: map[string]interface{}{
			"company_name":  "supplier alpha",
			"contact_email": "bad-email",
		},
	}

	_, err := svc.ValidateAndStandardizeEntity(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestValidateAndStandardizeEntity_InvalidEntityType(t *testing.T) {
	svc := NewService()
	req := &domain.ValidateEntityRequest{
		EntityType: "branch",
		Data: map[string]interface{}{
			"name": "central",
		},
	}

	_, err := svc.ValidateAndStandardizeEntity(context.Background(), req)
	if err == nil {
		t.Fatal("expected invalid entity_type error, got nil")
	}
}
