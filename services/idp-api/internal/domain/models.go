package domain

import "time"

// ExtractRequest specifies a target file reference for processing.
type ExtractRequest struct {
	FileID   string `json:"file_id"`
	FilePath string `json:"file_path"`
}

// ExtractedData is the simulated AI/OCR structured output.
type ExtractedData struct {
	DocumentType string  `json:"document_type"`
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	VendorName   string  `json:"vendor_name"`
}

// ExtractResult is the processing result envelope payload.
type ExtractResult struct {
	FileID      string        `json:"file_id,omitempty"`
	FilePath    string        `json:"file_path,omitempty"`
	Extracted   ExtractedData `json:"extracted_data"`
	ProcessedAt time.Time     `json:"processed_at"`
}

// SuccessResponse is a standard success envelope.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse is a standard error envelope.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
}
