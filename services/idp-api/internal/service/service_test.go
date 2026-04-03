package service

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/user/pos-wms-mvp/services/idp-api/internal/domain"
)

func TestSimulateExtraction_SuccessWithFileID(t *testing.T) {
	svc := NewService()
	start := time.Now()

	result, err := svc.SimulateExtraction(context.Background(), &domain.ExtractRequest{FileID: "file_123"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.FileID != "file_123" {
		t.Fatalf("expected file_id file_123, got %s", result.FileID)
	}
	if result.Extracted.DocumentType != "invoice" {
		t.Fatalf("expected document_type invoice, got %s", result.Extracted.DocumentType)
	}
	if result.Extracted.Amount != 1500.00 {
		t.Fatalf("expected amount 1500, got %v", result.Extracted.Amount)
	}

	if time.Since(start) < simulatedDelay {
		t.Fatalf("expected simulated delay at least %s", simulatedDelay)
	}
}

func TestSimulateExtraction_RequiresReference(t *testing.T) {
	svc := NewService()
	_, err := svc.SimulateExtraction(context.Background(), &domain.ExtractRequest{})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "either file_id or file_path is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
