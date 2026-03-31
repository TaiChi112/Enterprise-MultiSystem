package service

import "testing"

func TestIsAllowedMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		want     bool
	}{
		{name: "jpeg image", mimeType: "image/jpeg", want: true},
		{name: "png image", mimeType: "image/png", want: true},
		{name: "pdf", mimeType: "application/pdf", want: true},
		{name: "zip", mimeType: "application/zip", want: false},
		{name: "text", mimeType: "text/plain", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isAllowedMimeType(tc.mimeType)
			if got != tc.want {
				t.Fatalf("isAllowedMimeType(%q)=%v, want %v", tc.mimeType, got, tc.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "normal", input: "receipt.pdf", expected: "receipt.pdf"},
		{name: "path traversal", input: "../../etc/passwd", expected: "passwd"},
		{name: "empty", input: "", expected: "unknown"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sanitizeFilename(tc.input)
			if got != tc.expected {
				t.Fatalf("sanitizeFilename(%q)=%q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
