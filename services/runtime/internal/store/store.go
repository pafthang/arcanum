package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/runtime/models"
)

// Store is the runtime SQLite database.
type Store struct {
	db *sql.DB
}

// OpenStore opens dataDir/runtime.db.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "runtime")
	if err != nil {
		return nil, err
	}
	if err := sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS machines (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			image TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			docker_id TEXT NOT NULL DEFAULT '',
			agent_id TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_machines_space ON machines(space_id, created_at)`,
	); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

// Close closes the database.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339Nano) }

func scan(scanFn func(dest ...any) error) (*models.Machine, error) {
	var m models.Machine
	err := scanFn(&m.ID, &m.SpaceID, &m.Name, &m.Image, &m.Status, &m.DockerID, &m.AgentID, &m.Error, &m.CreatedAt, &m.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// Create inserts a machine row.
func (s *Store) Create(ctx context.Context, spaceID, name, image, agentID, status string) (*models.Machine, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	if status == "" {
		status = models.StatusRecorded
	}
	if image == "" {
		image = "debian:bookworm-slim"
	}
	now := nowRFC3339()
	m := &models.Machine{
		ID: idgen.New(), SpaceID: spaceID, Name: name, Image: image,
		Status: status, AgentID: strings.TrimSpace(agentID), CreatedAt: now, UpdatedAt: now,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO machines (id, space_id, name, image, status, docker_id, agent_id, error, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		m.ID, m.SpaceID, m.Name, m.Image, m.Status, m.DockerID, m.AgentID, m.Error, m.CreatedAt, m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetInSpace returns a machine in a space.
func (s *Store) GetInSpace(ctx context.Context, spaceID, id string) (*models.Machine, error) {
	return scan(s.db.QueryRowContext(ctx, `
SELECT id, space_id, name, image, status, docker_id, agent_id, error, created_at, updated_at
FROM machines WHERE id = ? AND space_id = ?`, strings.TrimSpace(id), strings.TrimSpace(spaceID)).Scan)
}

// List returns machines in a space.
func (s *Store) List(ctx context.Context, spaceID string) ([]models.Machine, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, image, status, docker_id, agent_id, error, created_at, updated_at
FROM machines WHERE space_id = ? ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Machine{}
	for rows.Next() {
		m, err := scan(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

// SetStatus updates status and optional docker id / error.
func (s *Store) SetStatus(ctx context.Context, spaceID, id, status, dockerID, errText string) (*models.Machine, error) {
	cur, err := s.GetInSpace(ctx, spaceID, id)
	if err != nil || cur == nil {
		return cur, err
	}
	cur.Status = status
	if dockerID != "" {
		cur.DockerID = dockerID
	}
	cur.Error = errText
	cur.UpdatedAt = nowRFC3339()
	_, err = s.db.ExecContext(ctx, `
UPDATE machines SET status=?, docker_id=?, error=?, updated_at=? WHERE id=? AND space_id=?`,
		cur.Status, cur.DockerID, cur.Error, cur.UpdatedAt, cur.ID, cur.SpaceID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}
