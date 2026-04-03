package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/user/pos-wms-mvp/pkg/config"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/repository"
)

const ediTransmitPath = "/edi/transmit"

func TestReplenishIntegrationTransmitsToEDIAndMarksPOTransmitted(t *testing.T) {
	db := openIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	productID := insertTestProduct(t, ctx, db)
	var createdPOID int
	var createdSupplierID int
	t.Cleanup(func() {
		cleanupReplenishFixtures(t, ctx, db, createdPOID, createdSupplierID, productID)
	})

	var captured domain.EDITransmitRequest
	mockEDI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != ediTransmitPath {
			t.Fatalf("unexpected EDI request: %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("failed to decode EDI payload: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"mock transmitted"}`))
	}))
	defer mockEDI.Close()

	t.Setenv("EDI_API_URL", mockEDI.URL+ediTransmitPath)
	t.Setenv("DEFAULT_SUPPLIER_ID", "0")
	t.Setenv("DEFAULT_SUPPLIER_NAME", fmt.Sprintf("Integration Supplier %d", time.Now().UnixNano()))
	t.Setenv("DEFAULT_SUPPLIER_CONTACT", "integration.supplier@example.com")

	svc := NewService(db)
	resp, err := svc.Replenish(ctx, &domain.ReplenishRequest{ProductID: productID, Quantity: 7})
	if err != nil {
		t.Fatalf("replenish returned error: %v", err)
	}

	if resp == nil || resp.PurchaseOrder == nil {
		t.Fatal("replenish response missing purchase order")
	}

	createdPOID = resp.PurchaseOrder.ID
	createdSupplierID = resp.PurchaseOrder.SupplierID

	if resp.PurchaseOrder.Status != domain.PurchaseOrderStatusTransmitted {
		t.Fatalf("expected final status %q, got %q", domain.PurchaseOrderStatusTransmitted, resp.PurchaseOrder.Status)
	}
	if !resp.EDI.Success {
		t.Fatalf("expected EDI success=true, got false")
	}

	if captured.PurchaseOrderID != createdPOID {
		t.Fatalf("expected EDI payload purchase_order_id=%d, got %d", createdPOID, captured.PurchaseOrderID)
	}
	if captured.ProductID != productID {
		t.Fatalf("expected EDI payload product_id=%d, got %d", productID, captured.ProductID)
	}
	if captured.Status != domain.PurchaseOrderStatusApproved {
		t.Fatalf("expected EDI payload status %q before transition, got %q", domain.PurchaseOrderStatusApproved, captured.Status)
	}

	poRow, err := db.GetPurchaseOrderByID(ctx, createdPOID)
	if err != nil {
		t.Fatalf("failed to load purchase_order row: %v", err)
	}
	if poRow.Status != domain.PurchaseOrderStatusTransmitted {
		t.Fatalf("expected DB purchase_order status %q, got %q", domain.PurchaseOrderStatusTransmitted, poRow.Status)
	}
}

func TestReplenishIntegrationProductFKViolationIsMapped(t *testing.T) {
	db := openIntegrationDB(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	mockEDICalls := 0
	mockEDI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mockEDICalls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"mock transmitted"}`))
	}))
	defer mockEDI.Close()

	t.Setenv("EDI_API_URL", mockEDI.URL+ediTransmitPath)
	t.Setenv("DEFAULT_SUPPLIER_ID", "0")
	t.Setenv("DEFAULT_SUPPLIER_NAME", fmt.Sprintf("Integration Supplier %d", time.Now().UnixNano()))
	t.Setenv("DEFAULT_SUPPLIER_CONTACT", "integration.supplier@example.com")

	missingProductID := lookupMissingProductID(t, ctx, db)
	svc := NewService(db)

	_, err := svc.Replenish(ctx, &domain.ReplenishRequest{ProductID: missingProductID, Quantity: 3})
	if err == nil {
		t.Fatal("expected replenish to fail for missing product_id")
	}
	if !errors.Is(err, ErrInvalidReplenishProduct) {
		t.Fatalf("expected ErrInvalidReplenishProduct, got %v", err)
	}
	if mockEDICalls != 0 {
		t.Fatalf("expected EDI not to be called when FK validation fails, got %d calls", mockEDICalls)
	}
}

func openIntegrationDB(t *testing.T) *repository.Database {
	t.Helper()

	dsn := os.Getenv("SCM_TEST_DB_DSN")
	if dsn == "" {
		dsn = config.GetDatabaseURL()
	}

	db, err := repository.NewDatabase(dsn)
	if err != nil {
		t.Skipf("skipping integration test: database unavailable (%v)", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	return db
}

func insertTestProduct(t *testing.T, ctx context.Context, db *repository.Database) int {
	t.Helper()

	productID := 0
	sku := fmt.Sprintf("SCM-IT-%d", time.Now().UnixNano())
	err := db.QueryRowContext(
		ctx,
		`INSERT INTO product (sku, name, price, is_active) VALUES ($1, $2, $3, true) RETURNING id`,
		sku,
		"SCM Integration Product",
		99.99,
	).Scan(&productID)
	if err != nil {
		t.Fatalf("failed to create integration product: %v", err)
	}

	return productID
}

func lookupMissingProductID(t *testing.T, ctx context.Context, db *repository.Database) int {
	t.Helper()

	missingID := 0
	err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 0) + 1000000 FROM product`).Scan(&missingID)
	if err != nil {
		t.Fatalf("failed to compute missing product_id: %v", err)
	}

	return missingID
}

func cleanupReplenishFixtures(t *testing.T, ctx context.Context, db *repository.Database, poID, supplierID, productID int) {
	t.Helper()

	if poID > 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM purchase_order WHERE id = $1`, poID)
		if err != nil {
			t.Logf("cleanup warning: failed deleting purchase_order %d: %v", poID, err)
		}
	}
	if supplierID > 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM supplier WHERE id = $1`, supplierID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			t.Logf("cleanup warning: failed deleting supplier %d: %v", supplierID, err)
		}
	}
	if productID > 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM product WHERE id = $1`, productID)
		if err != nil {
			t.Logf("cleanup warning: failed deleting product %d: %v", productID, err)
		}
	}
}
