package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/user/pos-wms-mvp/services/idp-api/internal/domain"
)

const simulatedDelay = 2 * time.Second

// Service contains IDP simulation business logic.
type Service struct{}

// NewService creates a new IDP service instance.
func NewService() *Service {
	return &Service{}
}

// SimulateExtraction mimics OCR/AI extraction from a file reference.
func (s *Service) SimulateExtraction(ctx context.Context, req *domain.ExtractRequest) (*domain.ExtractResult, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	fileID := strings.TrimSpace(req.FileID)
	filePath := strings.TrimSpace(req.FilePath)
	if fileID == "" && filePath == "" {
		return nil, fmt.Errorf("either file_id or file_path is required")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(simulatedDelay):
	}

	return &domain.ExtractResult{
		FileID:   fileID,
		FilePath: filePath,
		Extracted: domain.ExtractedData{
			DocumentType: "invoice",
			Amount:       1500.00,
			Currency:     "THB",
			VendorName:   "Demo Supplier Co., Ltd.",
		},
		ProcessedAt: time.Now().UTC(),
	}, nil
}
