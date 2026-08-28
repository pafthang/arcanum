package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/services/work/models"
)

func migrateIssueExtra(db *sql.DB) error {
	stmts := []string{
		`ALTER TABLE issues ADD COLUMN priority TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE issues ADD COLUMN due_at TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE issues ADD COLUMN parent_id TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS issue_assignees (
			issue_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			PRIMARY KEY (issue_id, user_id)
		)`,
		`CREATE TABLE IF NOT EXISTS issue_relations (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			from_id TEXT NOT NULL,
			to_id TEXT NOT NULL,
			kind TEXT NOT NULL CHECK (kind IN ('blocks','blocked','duplicate','related')),
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_issue_relations_from ON issue_relations(from_id)`,
		`CREATE INDEX IF NOT EXISTS idx_issue_relations_to ON issue_relations(to_id)`,
	}
	for _, q := range stmts {
		if _, err := db.Exec(q); err != nil {
			if !strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
				return err
			}
		}
	}
	return nil
}

// SetIssueFields updates priority, due date and parent.
func (s *Store) SetIssueFields(ctx context.Context, id, priority, dueAt, parentID string) error {
	_, err := s.db.ExecContext(ctx, `
UPDATE issues SET priority=?, due_at=?, parent_id=?, updated_at=? WHERE id=?`,
		strings.TrimSpace(priority), strings.TrimSpace(dueAt), strings.TrimSpace(parentID), nowRFC3339(), strings.TrimSpace(id))
	return err
}

// ReplaceAssignees replaces extra assignees for an issue (primary assignee_id stays on the row).
func (s *Store) ReplaceAssignees(ctx context.Context, issueID string, userIDs []string) error {
	issueID = strings.TrimSpace(issueID)
	if _, err := s.db.ExecContext(ctx, `DELETE FROM issue_assignees WHERE issue_id=?`, issueID); err != nil {
		return err
	}
	for _, uid := range userIDs {
		uid = strings.TrimSpace(uid)
		if uid == "" {
			continue
		}
		if _, err := s.db.ExecContext(ctx, `INSERT OR IGNORE INTO issue_assignees (issue_id, user_id) VALUES (?,?)`, issueID, uid); err != nil {
			return err
		}
	}
	return nil
}

// ListAssignees returns extra assignee ids.
func (s *Store) ListAssignees(ctx context.Context, issueID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT user_id FROM issue_assignees WHERE issue_id=?`, strings.TrimSpace(issueID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// AddRelation links two issues.
func (s *Store) AddRelation(ctx context.Context, spaceID, fromID, toID, kind string) (*models.IssueRelation, error) {
	kind = strings.TrimSpace(kind)
	switch kind {
	case "blocks", "blocked", "duplicate", "related":
	default:
		return nil, fmt.Errorf("invalid relation kind")
	}
	r := &models.IssueRelation{
		ID:        idgen.New(),
		SpaceID:   strings.TrimSpace(spaceID),
		FromID:    strings.TrimSpace(fromID),
		ToID:      strings.TrimSpace(toID),
		Kind:      kind,
		CreatedAt: nowRFC3339(),
	}
	if r.FromID == "" || r.ToID == "" || r.FromID == r.ToID {
		return nil, fmt.Errorf("fromId and toId required and distinct")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO issue_relations (id, space_id, from_id, to_id, kind, created_at) VALUES (?,?,?,?,?,?)`,
		r.ID, r.SpaceID, r.FromID, r.ToID, r.Kind, r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ListRelations returns relations where the issue is from or to.
func (s *Store) ListRelations(ctx context.Context, issueID string) ([]models.IssueRelation, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, from_id, to_id, kind, created_at
FROM issue_relations WHERE from_id=? OR to_id=?
ORDER BY created_at ASC`, strings.TrimSpace(issueID), strings.TrimSpace(issueID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.IssueRelation{}
	for rows.Next() {
		var r models.IssueRelation
		if err := rows.Scan(&r.ID, &r.SpaceID, &r.FromID, &r.ToID, &r.Kind, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// HydrateIssue fills extra fields.
func (s *Store) HydrateIssue(ctx context.Context, iss *models.Issue) error {
	if iss == nil {
		return nil
	}
	var priority, dueAt, parent sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT priority, due_at, parent_id FROM issues WHERE id=?`, iss.ID).
		Scan(&priority, &dueAt, &parent)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			return nil
		}
		return err
	}
	iss.Priority = priority.String
	iss.DueAt = dueAt.String
	iss.ParentID = parent.String
	ids, err := s.ListAssignees(ctx, iss.ID)
	if err != nil {
		return err
	}
	iss.AssigneeIDs = ids
	rels, err := s.ListRelations(ctx, iss.ID)
	if err != nil {
		return err
	}
	iss.Relations = rels
	return s.FillIssueLabels(ctx, iss)
}
