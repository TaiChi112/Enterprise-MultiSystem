package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/user/pos-wms-mvp/services/ecm-api/internal/domain"
)

// Service contains ECM upload business logic.
type Service struct {
	uploadDir string
}

// NewService creates a new ECM service instance.
func NewService(uploadDir string) *Service {
	if strings.TrimSpace(uploadDir) == "" {
		uploadDir = "./uploads"
	}
	return &Service{uploadDir: uploadDir}
}

// SaveUpload validates file type and writes file to local upload directory.
func (s *Service) SaveUpload(fileHeader *multipart.FileHeader) (*domain.UploadResult, error) {
	if fileHeader == nil {
		return nil, fmt.Errorf("file is required")
	}

	if err := os.MkdirAll(s.uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to ensure upload directory: %w", err)
	}

	stream, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer stream.Close()

	headerBuf := make([]byte, 512)
	readN, readErr := stream.Read(headerBuf)
	if readErr != nil && readErr != io.EOF {
		return nil, fmt.Errorf("failed to inspect uploaded file: %w", readErr)
	}

	detectedMime := http.DetectContentType(headerBuf[:readN])
	if !isAllowedMimeType(detectedMime) {
		return nil, fmt.Errorf("unsupported file type: %s", detectedMime)
	}

	if _, err := stream.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to rewind uploaded file: %w", err)
	}

	fileID, err := newFileID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate file id: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	storedName := fileID + ext
	storagePath := filepath.Join(s.uploadDir, storedName)

	outFile, err := os.Create(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	sizeBytes, err := io.Copy(outFile, stream)
	if err != nil {
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}

	return &domain.UploadResult{
		FileID:       fileID,
		OriginalName: sanitizeFilename(fileHeader.Filename),
		StoredName:   storedName,
		StoragePath:  storagePath,
		SizeBytes:    sizeBytes,
		MimeType:     detectedMime,
		UploadedAt:   time.Now().UTC(),
	}, nil
}

func isAllowedMimeType(mimeType string) bool {
	if strings.HasPrefix(mimeType, "image/") {
		return true
	}
	return mimeType == "application/pdf"
}

func sanitizeFilename(name string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "." || base == string(filepath.Separator) || base == "" {
		return "unknown"
	}
	return base
}

func newFileID() (string, error) {
	randBytes := make([]byte, 8)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("file_%d_%s", time.Now().UTC().Unix(), hex.EncodeToString(randBytes)), nil
}
