package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/work/models"
)

func migrateLabels(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS labels (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			color TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_labels_space_name ON labels(space_id, name)`,
		`CREATE TABLE IF NOT EXISTS issue_labels (
			issue_id TEXT NOT NULL,
			label_id TEXT NOT NULL,
			PRIMARY KEY (issue_id, label_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_issue_labels_label ON issue_labels(label_id)`,
	)
}

// CreateLabel inserts a space label. Name is unique per space.
func (s *Store) CreateLabel(ctx context.Context, spaceID, name, color string) (*models.Label, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	lb := &models.Label{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		Name:      name,
		Color:     strings.TrimSpace(color),
		CreatedAt: nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO labels (id, space_id, name, color, created_at) VALUES (?,?,?,?,?)`,
		lb.ID, lb.SpaceID, lb.Name, lb.Color, lb.CreatedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, fmt.Errorf("label name already exists")
		}
		return nil, err
	}
	return lb, nil
}

// ListLabels returns labels in a space, by name.
func (s *Store) ListLabels(ctx context.Context, spaceID string) ([]models.Label, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, color, created_at
FROM labels WHERE space_id = ?
ORDER BY name ASC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Label{}
	for rows.Next() {
		var lb models.Label
		if err := rows.Scan(&lb.ID, &lb.SpaceID, &lb.Name, &lb.Color, &lb.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, lb)
	}
	return out, rows.Err()
}

// GetLabel returns a label by id.
func (s *Store) GetLabel(ctx context.Context, id string) (*models.Label, error) {
	var lb models.Label
	err := s.db.QueryRowContext(ctx, `
SELECT id, space_id, name, color, created_at FROM labels WHERE id = ?`, strings.TrimSpace(id)).
		Scan(&lb.ID, &lb.SpaceID, &lb.Name, &lb.Color, &lb.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &lb, nil
}
