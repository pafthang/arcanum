package proxy

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// UploadConfig configures file uploads.
type UploadConfig struct {
	MaxSize      int64    // max size in bytes
	AllowedTypes []string // MIME types; empty = any
	AllowedExts  []string // extensions (with dot); empty = any
}

// DefaultUploadConfig is a sensible default.
func DefaultUploadConfig() UploadConfig {
	return UploadConfig{
		MaxSize: 32 << 20, // 32 MiB
		AllowedTypes: []string{
			"image/jpeg", "image/png", "image/gif", "image/webp",
			"application/pdf",
			"text/plain",
			"application/json",
		},
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf", ".txt", ".json"},
	}
}

// ValidateUpload checks Content-Type, extension, and size.
func ValidateUpload(r *http.Request, cfg UploadConfig) error {
	if cfg.MaxSize > 0 && r.ContentLength > cfg.MaxSize {
		return fmt.Errorf("file too large: %d > %d", r.ContentLength, cfg.MaxSize)
	}

	ct := r.Header.Get("Content-Type")
	if len(cfg.AllowedTypes) > 0 && ct != "" {
		mainType := strings.Split(ct, ";")[0]
		mainType = strings.TrimSpace(mainType)
		ok := false
		for _, t := range cfg.AllowedTypes {
			if t == mainType {
				ok = true
				break
			}
		}
		if !ok {
			return fmt.Errorf("content-type not allowed: %s", mainType)
		}
	}

	return nil
}

// StreamUpload reads the request body with a size limit into dst.
func StreamUpload(r *http.Request, dst io.Writer, maxSize int64) (int64, error) {
	if maxSize <= 0 {
		maxSize = 32 << 20
	}
	limited := io.LimitReader(r.Body, maxSize+1)
	n, err := io.Copy(dst, limited)
	if err != nil {
		return n, err
	}
	if n > maxSize {
		return n, fmt.Errorf("file too large")
	}
	return n, nil
}

// ExtAllowed checks the file extension.
func ExtAllowed(filename string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	ext := strings.ToLower(filepath.Ext(filename))
	for _, a := range allowed {
		if strings.ToLower(a) == ext {
			return true
		}
	}
	return false
}
