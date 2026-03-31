package domain

import "time"

// UploadResult contains stored file metadata after successful upload.
type UploadResult struct {
	FileID       string    `json:"file_id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"stored_name"`
	StoragePath  string    `json:"storage_path"`
	SizeBytes    int64     `json:"size_bytes"`
	MimeType     string    `json:"mime_type"`
	UploadedAt   time.Time `json:"uploaded_at"`
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
