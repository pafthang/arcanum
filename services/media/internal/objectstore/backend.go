package objectstore

import "context"

// Backend stores blob bytes. Metadata stays in SQLite.
type Backend interface {
	Name() string
	Put(ctx context.Context, key, contentType string, data []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	// PresignGet returns a time-limited GET URL. Empty means caller should
	// mint a local HMAC URL instead.
	PresignGet(ctx context.Context, key, filename, contentType string, expiresSec int64) (string, error)
}
