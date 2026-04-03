package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/domain"
	"github.com/user/pos-wms-mvp/services/scm-api/internal/repository"
)

// Service holds repository dependencies and business logic.
type Service struct {
	repo       *repository.Database
	ediAPIURL  string
	client     *http.Client
	defaultSID int
	defaultQty int
	defaultSN  string
	defaultSC  string
}

func NewService(repo *repository.Database) *Service {
	return &Service{
		repo:       repo,
		ediAPIURL:  getEnv("EDI_API_URL", "http://localhost:4005/edi/transmit"),
		client:     &http.Client{Timeout: 10 * time.Second},
		defaultSID: getEnvInt("DEFAULT_SUPPLIER_ID", 1),
		defaultQty: getEnvInt("DEFAULT_REPLENISH_QTY", 50),
		defaultSN:  getEnv("DEFAULT_SUPPLIER_NAME", "Default Supplier"),
		defaultSC:  getEnv("DEFAULT_SUPPLIER_CONTACT", "default.supplier@example.com"),
	}
}

func (s *Service) CreateSupplier(ctx context.Context, req *domain.CreateSupplierRequest) (*domain.Supplier, error) {
	supplier := &domain.Supplier{
		Name:    strings.TrimSpace(req.Name),
		Contact: stringPtr(strings.TrimSpace(req.Contact)),
	}

	if err := s.repo.CreateSupplier(ctx, supplier); err != nil {
		return nil, err
	}

	return supplier, nil
}

func (s *Service) GetSupplier(ctx context.Context, id int) (*domain.Supplier, error) {
	return s.repo.GetSupplierByID(ctx, id)
}

func (s *Service) GetSuppliers(ctx context.Context, limit, offset int) ([]*domain.Supplier, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetSuppliers(ctx, limit, offset)
}

func (s *Service) UpdateSupplier(ctx context.Context, id int, req *domain.UpdateSupplierRequest) (*domain.Supplier, error) {
	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		req.Name = &name
	}
	if req.Contact != nil {
		contact := strings.TrimSpace(*req.Contact)
		req.Contact = &contact
	}

	return s.repo.UpdateSupplier(ctx, id, req)
}

func (s *Service) DeleteSupplier(ctx context.Context, id int) error {
	return s.repo.DeleteSupplier(ctx, id)
}

func (s *Service) Replenish(ctx context.Context, req *domain.ReplenishRequest) (*domain.ReplenishResponse, error) {
	if req.ProductID <= 0 {
		return nil, fmt.Errorf("product_id must be greater than 0")
	}

	quantity := req.Quantity
	if quantity <= 0 {
		quantity = s.defaultQty
	}
	if quantity <= 0 {
		quantity = 1
	}

	supplierID, err := s.ensureDefaultSupplier(ctx)
	if err != nil {
		return nil, err
	}

	po := &domain.PurchaseOrder{
		SupplierID: supplierID,
		ProductID:  req.ProductID,
		Quantity:   quantity,
		Status:     domain.PurchaseOrderStatusApproved,
	}

	if err := s.repo.CreatePurchaseOrder(ctx, po); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23503" && pqErr.Constraint == "fk_purchase_order_product" {
			return nil, fmt.Errorf("%w: product_id=%d", ErrInvalidReplenishProduct, req.ProductID)
		}
		return nil, err
	}

	ack, err := s.transmitToEDI(ctx, po)
	if err != nil {
		return nil, fmt.Errorf("failed to transmit purchase order to edi: %w", err)
	}

	if err := s.repo.UpdatePurchaseOrderStatus(ctx, po.ID, domain.PurchaseOrderStatusTransmitted); err != nil {
		return nil, err
	}
	po.Status = domain.PurchaseOrderStatusTransmitted

	return &domain.ReplenishResponse{PurchaseOrder: po, EDI: *ack}, nil
}

func (s *Service) GetPurchaseOrders(ctx context.Context, limit, offset int) ([]*domain.PurchaseOrder, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.GetPurchaseOrders(ctx, limit, offset)
}

func (s *Service) ensureDefaultSupplier(ctx context.Context) (int, error) {
	if s.defaultSID > 0 {
		supplier, err := s.repo.GetSupplierByID(ctx, s.defaultSID)
		if err == nil && supplier != nil {
			return supplier.ID, nil
		}
	}

	newSupplier := &domain.Supplier{
		Name:    strings.TrimSpace(s.defaultSN),
		Contact: stringPtr(strings.TrimSpace(s.defaultSC)),
	}
	if newSupplier.Name == "" {
		newSupplier.Name = "Default Supplier"
	}

	if err := s.repo.CreateSupplier(ctx, newSupplier); err != nil {
		return 0, fmt.Errorf("failed to ensure default supplier: %w", err)
	}

	return newSupplier.ID, nil
}

func (s *Service) transmitToEDI(ctx context.Context, po *domain.PurchaseOrder) (*domain.EDITransmitAck, error) {
	body := domain.EDITransmitRequest{
		PurchaseOrderID: po.ID,
		SupplierID:      po.SupplierID,
		ProductID:       po.ProductID,
		Quantity:        po.Quantity,
		Status:          po.Status,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.ediAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("edi returned non-success status: %d", resp.StatusCode)
	}

	ack := &domain.EDITransmitAck{}
	if err := json.NewDecoder(resp.Body).Decode(ack); err != nil {
		return nil, err
	}

	return ack, nil
}

func stringPtr(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
