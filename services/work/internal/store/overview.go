package store

import (
	"context"
	"strings"

	"github.com/pafthang/arcanum/services/work/models"
)

// Overview aggregates issue counts for a space.
func (s *Store) Overview(ctx context.Context, spaceID string) (*models.Overview, error) {
	spaceID = strings.TrimSpace(spaceID)
	ov := &models.Overview{
		ByStatus: map[string]int{
			models.StatusOpen:    0,
			models.StatusStarted: 0,
			models.StatusDone:    0,
		},
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT status, COUNT(*), SUM(CASE WHEN assignee_id != '' THEN 1 ELSE 0 END)
FROM issues WHERE space_id = ?
GROUP BY status`, spaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var n, assigned int
		if err := rows.Scan(&status, &n, &assigned); err != nil {
			return nil, err
		}
		ov.ByStatus[status] = n
		ov.Issues += n
		ov.Assigned += assigned
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ov.Unassigned = ov.Issues - ov.Assigned
	if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM issue_comments c
JOIN issues i ON i.id = c.issue_id
WHERE i.space_id = ?`, spaceID).Scan(&ov.Comments); err != nil {
		return nil, err
	}
	return ov, nil
}
