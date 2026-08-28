package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/media/internal/objectstore"
	"github.com/pafthang/arcanum/services/media/models"
)

// Store is metadata SQLite plus blob files under dataDir/blobs.
type Store struct {
	db      *sql.DB
	dataDir string
	Objects objectstore.Backend
}

// OpenStore opens dataDir/media.db and ensures the blob directory.
func OpenStore(dataDir string) (*Store, error) {
	return OpenStoreBackend(dataDir, objectstore.NewFS(filepath.Join(dataDir, "blobs")))
}

// OpenStoreBackend opens metadata DB with an explicit object backend.
func OpenStoreBackend(dataDir string, objects objectstore.Backend) (*Store, error) {
	db, err := sqldb.Open(dataDir, "media")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if objects == nil {
		objects = objectstore.NewFS(filepath.Join(dataDir, "blobs"))
	}
	return &Store{db: db, dataDir: dataDir, Objects: objects}, nil
}

func (s *Store) objectKey(spaceID, id string) string {
	return strings.TrimSpace(spaceID) + "/" + strings.TrimSpace(id)
}

func migrate(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS blobs (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			filename TEXT NOT NULL,
			content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
			size INTEGER NOT NULL,
			sha256 TEXT NOT NULL,
			actor_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_blobs_space ON blobs(space_id, created_at)`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func (s *Store) blobPath(spaceID, id string) string {
	return filepath.Join(s.dataDir, "blobs", spaceID, id)
}

// Put writes bytes to the object backend and inserts metadata.
func (s *Store) Put(ctx context.Context, spaceID, filename, contentType, actorID string, data []byte) (*models.Blob, error) {
	spaceID = strings.TrimSpace(spaceID)
	filename = strings.TrimSpace(filename)
	if spaceID == "" {
		return nil, fmt.Errorf("space_id required")
	}
	if filename == "" {
		filename = "blob"
	}
	contentType = strings.TrimSpace(contentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	sum := sha256.Sum256(data)
	b := &models.Blob{
		ID:          idgen.New(),
		SpaceID:     spaceID,
		Filename:    filepath.Base(filename),
		ContentType: contentType,
		Size:        int64(len(data)),
		SHA256:      hex.EncodeToString(sum[:]),
		ActorID:     strings.TrimSpace(actorID),
		CreatedAt:   nowRFC3339(),
	}
	if s.Objects != nil {
		if err := s.Objects.Put(ctx, s.objectKey(spaceID, b.ID), contentType, data); err != nil {
			return nil, err
		}
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO blobs (id, space_id, filename, content_type, size, sha256, actor_id, created_at)
VALUES (?,?,?,?,?,?,?,?)`,
		b.ID, b.SpaceID, b.Filename, b.ContentType, b.Size, b.SHA256, b.ActorID, b.CreatedAt)
	if err != nil {
		if s.Objects != nil {
			_ = s.Objects.Delete(ctx, s.objectKey(spaceID, b.ID))
		}
		return nil, err
	}
	return b, nil
}

// GetMeta returns blob metadata in a space.
func (s *Store) GetMeta(ctx context.Context, spaceID, id string) (*models.Blob, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, space_id, filename, content_type, size, sha256, actor_id, created_at
FROM blobs WHERE id = ? AND space_id = ?`, strings.TrimSpace(id), strings.TrimSpace(spaceID))
	var b models.Blob
	err := row.Scan(&b.ID, &b.SpaceID, &b.Filename, &b.ContentType, &b.Size, &b.SHA256, &b.ActorID, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &b, nil
}

// ReadBytes returns file contents for a blob in a space.
func (s *Store) ReadBytes(ctx context.Context, spaceID, id string) (*models.Blob, []byte, error) {
	meta, err := s.GetMeta(ctx, spaceID, id)
	if err != nil || meta == nil {
		return meta, nil, err
	}
	if s.Objects == nil {
		return meta, nil, fmt.Errorf("object backend missing")
	}
	data, err := s.Objects.Get(ctx, s.objectKey(spaceID, id))
	if err != nil {
		return nil, nil, err
	}
	return meta, data, nil
}

// List returns blobs in a space, newest first.
func (s *Store) List(ctx context.Context, spaceID string) ([]models.Blob, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, filename, content_type, size, sha256, actor_id, created_at
FROM blobs WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Blob{}
	for rows.Next() {
		var b models.Blob
		if err := rows.Scan(&b.ID, &b.SpaceID, &b.Filename, &b.ContentType, &b.Size, &b.SHA256, &b.ActorID, &b.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}
