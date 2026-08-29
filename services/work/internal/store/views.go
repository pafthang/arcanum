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

func migrateViews(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS work_views (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			query TEXT NOT NULL DEFAULT '',
			icon TEXT NOT NULL DEFAULT '',
			created_by TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_views_space ON work_views(space_id, created_at)`,
	)
}

// CreateView creates a saved view filter in a space.
func (s *Store) CreateView(ctx context.Context, spaceID, name, description, query, icon, createdBy string) (*models.View, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	v := &models.View{
		ID:          idgen.New(),
		SpaceID:     spaceID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Query:       strings.TrimSpace(query),
		Icon:        strings.TrimSpace(icon),
		CreatedBy:   strings.TrimSpace(createdBy),
		CreatedAt:   nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO work_views (id, space_id, name, description, query, icon, created_by, created_at)
VALUES (?,?,?,?,?,?,?,?)`,
		v.ID, v.SpaceID, v.Name, v.Description, v.Query, v.Icon, v.CreatedBy, v.CreatedAt)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// ListViews lists saved views in a space.
func (s *Store) ListViews(ctx context.Context, spaceID string) ([]models.View, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, description, query, icon, created_by, created_at
FROM work_views WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.View{}
	for rows.Next() {
		var v models.View
		if err := rows.Scan(&v.ID, &v.SpaceID, &v.Name, &v.Description, &v.Query, &v.Icon, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// DeleteView removes a saved view.
func (s *Store) DeleteView(ctx context.Context, spaceID, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM work_views WHERE space_id = ? AND id = ?`,
		strings.TrimSpace(spaceID), strings.TrimSpace(id))
	return err
}
