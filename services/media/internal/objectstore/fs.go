package objectstore

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FS stores objects under root/{key}.
type FS struct {
	Root string
}

// NewFS creates a filesystem backend.
func NewFS(root string) *FS {
	return &FS{Root: root}
}

func (f *FS) Name() string { return "fs" }

func (f *FS) path(key string) string {
	key = filepath.Clean(strings.ReplaceAll(key, "\\", "/"))
	return filepath.Join(f.Root, filepath.FromSlash(key))
}

func (f *FS) Put(_ context.Context, key, _ string, data []byte) error {
	p := f.path(key)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

func (f *FS) Get(_ context.Context, key string) ([]byte, error) {
	b, err := os.ReadFile(f.path(key))
	if err != nil {
		return nil, fmt.Errorf("fs get: %w", err)
	}
	return b, nil
}

func (f *FS) Delete(_ context.Context, key string) error {
	_ = os.Remove(f.path(key))
	return nil
}

func (f *FS) PresignGet(context.Context, string, string, string, int64) (string, error) {
	return "", nil
}
