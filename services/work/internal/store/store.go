package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/work/models"
)

// Store is the work SQLite database.
type Store struct {
	db *sql.DB
}

// OpenStore opens dataDir/work.db and migrates.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "work")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateLabels(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateIssueExtra(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateCommentBlob(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateCycles(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateProjects(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateViews(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS issues (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			title TEXT NOT NULL,
			body TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL CHECK (status IN ('open','started','done')),
			assignee_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS issue_comments (
			id TEXT PRIMARY KEY NOT NULL,
			issue_id TEXT NOT NULL,
			actor_id TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (issue_id) REFERENCES issues(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_issues_space ON issues(space_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_issues_assignee ON issues(assignee_id)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_issue ON issue_comments(issue_id, created_at)`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// ValidStatus reports whether status is allowed.
func ValidStatus(status string) bool {
	switch status {
	case models.StatusOpen, models.StatusStarted, models.StatusDone:
		return true
	}
	return false
}

// CreateIssue inserts an issue. Default status is open.
func (s *Store) CreateIssue(ctx context.Context, spaceID, title, body, status, assigneeID string) (*models.Issue, error) {
	spaceID = strings.TrimSpace(spaceID)
	title = strings.TrimSpace(title)
	if spaceID == "" || title == "" {
		return nil, fmt.Errorf("space_id and title required")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = models.StatusOpen
	}
	if !ValidStatus(status) {
		return nil, fmt.Errorf("invalid status")
	}
	now := nowRFC3339()
	iss := &models.Issue{
		ID:         idgen.New(),
		SpaceID:    spaceID,
		Title:      title,
		Body:       body,
		Status:     status,
		AssigneeID: strings.TrimSpace(assigneeID),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO issues (id, space_id, title, body, status, assignee_id, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?)`,
		iss.ID, iss.SpaceID, iss.Title, iss.Body, iss.Status, iss.AssigneeID, iss.CreatedAt, iss.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return iss, nil
}

func scanIssue(row *sql.Row) (*models.Issue, error) {
	var iss models.Issue
	err := row.Scan(&iss.ID, &iss.SpaceID, &iss.Title, &iss.Body, &iss.Status, &iss.AssigneeID, &iss.CreatedAt, &iss.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &iss, nil
}

// GetIssue returns an issue by id.
func (s *Store) GetIssue(ctx context.Context, id string) (*models.Issue, error) {
	iss, err := scanIssue(s.db.QueryRowContext(ctx, `
SELECT id, space_id, title, body, status, assignee_id, created_at, updated_at
FROM issues WHERE id = ?`, strings.TrimSpace(id)))
	if err != nil || iss == nil {
		return iss, err
	}
	if err := s.HydrateIssue(ctx, iss); err != nil {
		return nil, err
	}
	return iss, nil
}

// GetIssueInSpace returns an issue if it belongs to spaceID.
func (s *Store) GetIssueInSpace(ctx context.Context, spaceID, id string) (*models.Issue, error) {
	iss, err := s.GetIssue(ctx, id)
	if err != nil || iss == nil {
		return iss, err
	}
	if iss.SpaceID != strings.TrimSpace(spaceID) {
		return nil, nil
	}
	return iss, nil
}

// ListIssues returns issues in a space, newest first.
func (s *Store) ListIssues(ctx context.Context, spaceID string) ([]models.Issue, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, title, body, status, assignee_id, created_at, updated_at
FROM issues WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Issue{}
	for rows.Next() {
		var iss models.Issue
		if err := rows.Scan(&iss.ID, &iss.SpaceID, &iss.Title, &iss.Body, &iss.Status, &iss.AssigneeID, &iss.CreatedAt, &iss.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, iss)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := s.HydrateIssue(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// UpdateIssue applies a partial update. nil fields are left unchanged.
func (s *Store) UpdateIssue(ctx context.Context, id string, title, body, status, assigneeID *string) (*models.Issue, error) {
	cur, err := s.GetIssue(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, nil
	}
	if title != nil {
		t := strings.TrimSpace(*title)
		if t == "" {
			return nil, fmt.Errorf("title required")
		}
		cur.Title = t
	}
	if body != nil {
		cur.Body = *body
	}
	if status != nil {
		st := strings.TrimSpace(*status)
		if !ValidStatus(st) {
			return nil, fmt.Errorf("invalid status")
		}
		cur.Status = st
	}
	if assigneeID != nil {
		cur.AssigneeID = strings.TrimSpace(*assigneeID)
	}
	cur.UpdatedAt = nowRFC3339()
	_, err = s.db.ExecContext(ctx, `
UPDATE issues SET title=?, body=?, status=?, assignee_id=?, updated_at=? WHERE id=?`,
		cur.Title, cur.Body, cur.Status, cur.AssigneeID, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

// AddComment appends a comment to an issue.
func (s *Store) AddComment(ctx context.Context, issueID, actorID, body, blobID string) (*models.Comment, error) {
	issueID = strings.TrimSpace(issueID)
	actorID = strings.TrimSpace(actorID)
	body = strings.TrimSpace(body)
	blobID = strings.TrimSpace(blobID)
	if issueID == "" || actorID == "" || (body == "" && blobID == "") {
		return nil, fmt.Errorf("issue_id, actor_id and body or blob_id required")
	}
	iss, err := s.GetIssue(ctx, issueID)
	if err != nil {
		return nil, err
	}
	if iss == nil {
		return nil, fmt.Errorf("issue not found")
	}
	c := &models.Comment{
		ID:        idgen.New(),
		IssueID:   issueID,
		ActorID:   actorID,
		Body:      body,
		BlobID:    blobID,
		CreatedAt: nowRFC3339(),
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO issue_comments (id, issue_id, actor_id, body, blob_id, created_at) VALUES (?,?,?,?,?,?)`,
		c.ID, c.IssueID, c.ActorID, c.Body, c.BlobID, c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ListComments returns comments for an issue, oldest first.
func (s *Store) ListComments(ctx context.Context, issueID string) ([]models.Comment, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, issue_id, actor_id, body, blob_id, created_at
FROM issue_comments WHERE issue_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(issueID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Comment{}
	for rows.Next() {
		var c models.Comment
		if err := rows.Scan(&c.ID, &c.IssueID, &c.ActorID, &c.Body, &c.BlobID, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
