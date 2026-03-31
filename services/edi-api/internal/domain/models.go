package domain

import "time"

// InternalPOPayload is the internal SCM payload received by EDI service.
type InternalPOPayload struct {
	PurchaseOrderID int    `json:"purchase_order_id"`
	SupplierID      int    `json:"supplier_id"`
	ProductID       int    `json:"product_id"`
	Quantity        int    `json:"quantity"`
	Status          string `json:"status"`
}

// ExternalEDIItem simulates one line item in standardized EDI payload.
type ExternalEDIItem struct {
	LineNumber int    `json:"line_number"`
	ProductRef string `json:"product_ref"`
	Quantity   int    `json:"quantity"`
}

// ExternalEDIPayload is a simulated vendor-facing EDI document.
type ExternalEDIPayload struct {
	DocumentType   string            `json:"document_type"`
	MessageID      string            `json:"message_id"`
	TradingPartner string            `json:"trading_partner"`
	PurchaseOrder  int               `json:"purchase_order"`
	GeneratedAt    time.Time         `json:"generated_at"`
	Items          []ExternalEDIItem `json:"items"`
}

type EDITransmitResponse struct {
	Success          bool               `json:"success"`
	Message          string             `json:"message"`
	ExternalDocument ExternalEDIPayload `json:"external_document"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
