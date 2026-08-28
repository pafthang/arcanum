package store

import (
	"context"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/services/work/models"
)

// SetIssueLabels replaces labels on an issue. Empty ids clears them.
func (s *Store) SetIssueLabels(ctx context.Context, spaceID, issueID string, labelIDs []string) error {
	spaceID = strings.TrimSpace(spaceID)
	issueID = strings.TrimSpace(issueID)
	iss, err := s.GetIssueInSpace(ctx, spaceID, issueID)
	if err != nil {
		return err
	}
	if iss == nil {
		return fmt.Errorf("issue not found")
	}
	seen := map[string]struct{}{}
	ids := make([]string, 0, len(labelIDs))
	for _, id := range labelIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		lb, err := s.GetLabel(ctx, id)
		if err != nil {
			return err
		}
		if lb == nil || lb.SpaceID != spaceID {
			return fmt.Errorf("label not in this space")
		}
		ids = append(ids, id)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `DELETE FROM issue_labels WHERE issue_id = ?`, issueID); err != nil {
		return err
	}
	for _, id := range ids {
		if _, err := tx.ExecContext(ctx, `INSERT INTO issue_labels (issue_id, label_id) VALUES (?,?)`, issueID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// LabelsForIssues returns labels grouped by issue id.
func (s *Store) LabelsForIssues(ctx context.Context, issueIDs []string) (map[string][]models.Label, error) {
	out := map[string][]models.Label{}
	if len(issueIDs) == 0 {
		return out, nil
	}
	args := make([]any, 0, len(issueIDs))
	ph := make([]string, 0, len(issueIDs))
	for _, id := range issueIDs {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		args = append(args, id)
		ph = append(ph, "?")
	}
	if len(args) == 0 {
		return out, nil
	}
	q := `
SELECT il.issue_id, l.id, l.space_id, l.name, l.color, l.created_at
FROM issue_labels il
JOIN labels l ON l.id = il.label_id
WHERE il.issue_id IN (` + strings.Join(ph, ",") + `)
ORDER BY l.name ASC`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var issueID string
		var lb models.Label
		if err := rows.Scan(&issueID, &lb.ID, &lb.SpaceID, &lb.Name, &lb.Color, &lb.CreatedAt); err != nil {
			return nil, err
		}
		out[issueID] = append(out[issueID], lb)
	}
	return out, rows.Err()
}

// FillIssueLabels sets Labels on a single issue.
func (s *Store) FillIssueLabels(ctx context.Context, iss *models.Issue) error {
	if iss == nil {
		return nil
	}
	by, err := s.LabelsForIssues(ctx, []string{iss.ID})
	if err != nil {
		return err
	}
	iss.Labels = by[iss.ID]
	if iss.Labels == nil {
		iss.Labels = []models.Label{}
	}
	return nil
}

// AttachLabels fills Issue.Labels for the given issues.
func (s *Store) AttachLabels(ctx context.Context, issues []models.Issue) error {
	if len(issues) == 0 {
		return nil
	}
	ids := make([]string, len(issues))
	for i := range issues {
		ids[i] = issues[i].ID
		if issues[i].Labels == nil {
			issues[i].Labels = []models.Label{}
		}
	}
	by, err := s.LabelsForIssues(ctx, ids)
	if err != nil {
		return err
	}
	for i := range issues {
		if lbs, ok := by[issues[i].ID]; ok {
			issues[i].Labels = lbs
		}
	}
	return nil
}
