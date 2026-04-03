package domain

import "time"

const (
	PurchaseOrderStatusDraft       = "draft"
	PurchaseOrderStatusApproved    = "approved"
	PurchaseOrderStatusTransmitted = "transmitted"
	PurchaseOrderStatusCancelled   = "cancelled"
)

// Supplier represents a supplier profile in SCM.
type Supplier struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	Contact   *string   `db:"contact" json:"contact"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// PurchaseOrder represents a replenishment order in SCM.
type PurchaseOrder struct {
	ID         int       `db:"id" json:"id"`
	SupplierID int       `db:"supplier_id" json:"supplier_id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	Quantity   int       `db:"quantity" json:"quantity"`
	Status     string    `db:"status" json:"status"`
	UnitCost   float64   `db:"unit_cost" json:"unit_cost,omitempty"`
	TotalCost  float64   `db:"total_cost" json:"total_cost,omitempty"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type CreateSupplierRequest struct {
	Name    string `json:"name" validate:"required,min=1,max=255"`
	Contact string `json:"contact" validate:"omitempty,max=255"`
}

type UpdateSupplierRequest struct {
	Name    *string `json:"name" validate:"omitempty,min=1,max=255"`
	Contact *string `json:"contact" validate:"omitempty,max=255"`
}

type ReplenishRequest struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity,omitempty"`
}

type EDITransmitRequest struct {
	PurchaseOrderID int    `json:"purchase_order_id"`
	SupplierID      int    `json:"supplier_id"`
	ProductID       int    `json:"product_id"`
	Quantity        int    `json:"quantity"`
	Status          string `json:"status"`
}

type EDITransmitAck struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ReplenishResponse struct {
	PurchaseOrder *PurchaseOrder `json:"purchase_order"`
	EDI           EDITransmitAck `json:"edi"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
