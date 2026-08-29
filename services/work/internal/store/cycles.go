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

func migrateCycles(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS work_cycles (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT 'upcoming',
			start_date TEXT NOT NULL DEFAULT '',
			end_date TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cycles_space ON work_cycles(space_id, created_at)`,
	)
}

// CreateCycle creates a new sprint/cycle in a space.
func (s *Store) CreateCycle(ctx context.Context, spaceID, name, description, status, startDate, endDate string) (*models.Cycle, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = "upcoming"
	}
	now := nowRFC3339()
	c := &models.Cycle{
		ID:          idgen.New(),
		SpaceID:     spaceID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Status:      status,
		StartDate:   strings.TrimSpace(startDate),
		EndDate:     strings.TrimSpace(endDate),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO work_cycles (id, space_id, name, description, status, start_date, end_date, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?)`,
		c.ID, c.SpaceID, c.Name, c.Description, c.Status, c.StartDate, c.EndDate, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetCycle fetches a cycle by id in a space.
func (s *Store) GetCycle(ctx context.Context, spaceID, id string) (*models.Cycle, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, space_id, name, description, status, start_date, end_date, created_at, updated_at
FROM work_cycles WHERE space_id = ? AND id = ?`, strings.TrimSpace(spaceID), strings.TrimSpace(id))
	var c models.Cycle
	err := row.Scan(&c.ID, &c.SpaceID, &c.Name, &c.Description, &c.Status, &c.StartDate, &c.EndDate, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// ListCycles lists all cycles in a space.
func (s *Store) ListCycles(ctx context.Context, spaceID string) ([]models.Cycle, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, description, status, start_date, end_date, created_at, updated_at
FROM work_cycles WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Cycle{}
	for rows.Next() {
		var c models.Cycle
		if err := rows.Scan(&c.ID, &c.SpaceID, &c.Name, &c.Description, &c.Status, &c.StartDate, &c.EndDate, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
