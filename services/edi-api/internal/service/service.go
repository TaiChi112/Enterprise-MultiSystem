package service

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/user/pos-wms-mvp/services/edi-api/internal/domain"
)

// Service contains EDI transformation and transmission simulation logic.
type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) TransformAndTransmit(internalPO *domain.InternalPOPayload) (*domain.EDITransmitResponse, error) {
	ediPayload := domain.ExternalEDIPayload{
		DocumentType:   "PO_850",
		MessageID:      fmt.Sprintf("EDI-PO-%d-%d", internalPO.PurchaseOrderID, time.Now().Unix()),
		TradingPartner: fmt.Sprintf("VENDOR-%d", internalPO.SupplierID),
		PurchaseOrder:  internalPO.PurchaseOrderID,
		GeneratedAt:    time.Now().UTC(),
		Items: []domain.ExternalEDIItem{
			{
				LineNumber: 1,
				ProductRef: fmt.Sprintf("PRODUCT-%d", internalPO.ProductID),
				Quantity:   internalPO.Quantity,
			},
		},
	}

	// Simulate outbound B2B transmission side effect.
	log.Printf("TRANSMITTED TO VENDOR successfully | po_id=%d | partner=%s", internalPO.PurchaseOrderID, ediPayload.TradingPartner)

	return &domain.EDITransmitResponse{
		Success:          true,
		Message:          "PO transformed and transmitted successfully",
		ExternalDocument: ediPayload,
	}, nil
}

func ValidateInternalPO(internalPO *domain.InternalPOPayload) error {
	if internalPO.PurchaseOrderID <= 0 {
		return fmt.Errorf("purchase_order_id must be greater than 0")
	}
	if internalPO.SupplierID <= 0 {
		return fmt.Errorf("supplier_id must be greater than 0")
	}
	if internalPO.ProductID <= 0 {
		return fmt.Errorf("product_id must be greater than 0")
	}
	if internalPO.Quantity <= 0 {
		return fmt.Errorf("quantity must be greater than 0")
	}
	if strings.TrimSpace(internalPO.Status) == "" {
		return fmt.Errorf("status is required")
	}
	return nil
}
