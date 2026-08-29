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

func migrateProjects(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS work_projects (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			key TEXT NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'planned',
			lead_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_projects_space ON work_projects(space_id, created_at)`,
	)
}

// CreateProject inserts a new project.
func (s *Store) CreateProject(ctx context.Context, spaceID, name, key, description, status, leadID string) (*models.Project, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "planned"
	}
	now := nowRFC3339()
	p := &models.Project{
		ID:          idgen.New(),
		SpaceID:     spaceID,
		Name:        name,
		Key:         strings.ToUpper(strings.TrimSpace(key)),
		Description: strings.TrimSpace(description),
		Status:      status,
		LeadID:      strings.TrimSpace(leadID),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO work_projects (id, space_id, name, key, description, status, lead_id, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?)`,
		p.ID, p.SpaceID, p.Name, p.Key, p.Description, p.Status, p.LeadID, p.CreatedAt, p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// GetProject fetches a project by id in a space.
func (s *Store) GetProject(ctx context.Context, spaceID, id string) (*models.Project, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, space_id, name, key, description, status, lead_id, created_at, updated_at
FROM work_projects WHERE space_id = ? AND id = ?`, strings.TrimSpace(spaceID), strings.TrimSpace(id))
	var p models.Project
	err := row.Scan(&p.ID, &p.SpaceID, &p.Name, &p.Key, &p.Description, &p.Status, &p.LeadID, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListProjects lists projects in a space.
func (s *Store) ListProjects(ctx context.Context, spaceID string) ([]models.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, key, description, status, lead_id, created_at, updated_at
FROM work_projects WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Project{}
	for rows.Next() {
		var p models.Project
		if err := rows.Scan(&p.ID, &p.SpaceID, &p.Name, &p.Key, &p.Description, &p.Status, &p.LeadID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
