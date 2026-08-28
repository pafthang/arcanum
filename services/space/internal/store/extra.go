package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/services/space/models"
)

// UpdateMember changes a membership role.
func (s *Store) UpdateMember(ctx context.Context, spaceID, userID, role string) (*models.Member, error) {
	role = strings.TrimSpace(role)
	if !validSpaceRole(role) {
		return nil, fmt.Errorf("invalid space role")
	}
	cur, err := s.GetMember(ctx, spaceID, userID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, nil
	}
	if cur.Role == RoleOwner && role != RoleOwner {
		n, err := s.countOwners(ctx, spaceID)
		if err != nil {
			return nil, err
		}
		if n <= 1 {
			return nil, fmt.Errorf("cannot demote the last owner")
		}
	}
	_, err = s.db.ExecContext(ctx, `
UPDATE space_members SET role = ? WHERE space_id = ? AND user_id = ?`,
		role, strings.TrimSpace(spaceID), strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	cur.Role = role
	return cur, nil
}

func (s *Store) countOwners(ctx context.Context, spaceID string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM space_members WHERE space_id = ? AND role = ?`,
		strings.TrimSpace(spaceID), RoleOwner).Scan(&n)
	return n, err
}

// GetTeam returns a team in a space.
func (s *Store) GetTeam(ctx context.Context, spaceID, teamID string) (*models.Team, error) {
	var t models.Team
	var parent sql.NullString
	err := s.db.QueryRowContext(ctx, `
SELECT id, space_id, parent_id, name, created_at FROM teams WHERE id = ? AND space_id = ?`,
		strings.TrimSpace(teamID), strings.TrimSpace(spaceID)).
		Scan(&t.ID, &t.SpaceID, &parent, &t.Name, &t.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if parent.Valid {
		t.ParentID = parent.String
	}
	return &t, nil
}

// ListTeams returns teams in a space.
func (s *Store) ListTeams(ctx context.Context, spaceID string) ([]models.Team, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, parent_id, name, created_at
FROM teams WHERE space_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Team{}
	for rows.Next() {
		var t models.Team
		var parent sql.NullString
		if err := rows.Scan(&t.ID, &t.SpaceID, &parent, &t.Name, &t.CreatedAt); err != nil {
			return nil, err
		}
		if parent.Valid {
			t.ParentID = parent.String
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ListTeamMembers returns members of a team.
func (s *Store) ListTeamMembers(ctx context.Context, teamID string) ([]models.TeamMember, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT team_id, user_id, role FROM team_members WHERE team_id = ?`, strings.TrimSpace(teamID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.TeamMember{}
	for rows.Next() {
		var m models.TeamMember
		if err := rows.Scan(&m.TeamID, &m.UserID, &m.Role); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// ListAPIKeys returns keys owned by a user.
func (s *Store) ListAPIKeys(ctx context.Context, userID string) ([]models.APIKey, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, user_id, created_at FROM api_keys WHERE user_id = ? ORDER BY created_at ASC`,
		strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.APIKey{}
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// ListAPIKeysInSpace returns keys whose owner is a member of the space.
func (s *Store) ListAPIKeysInSpace(ctx context.Context, spaceID string) ([]models.APIKey, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT k.id, k.user_id, k.created_at
FROM api_keys k
JOIN space_members m ON m.user_id = k.user_id
WHERE m.space_id = ?
ORDER BY k.created_at ASC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.APIKey{}
	for rows.Next() {
		var k models.APIKey
		if err := rows.Scan(&k.ID, &k.UserID, &k.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// RemoveMember deletes a membership. Last owner cannot be removed.
func (s *Store) RemoveMember(ctx context.Context, spaceID, userID string) error {
	cur, err := s.GetMember(ctx, spaceID, userID)
	if err != nil {
		return err
	}
	if cur == nil {
		return fmt.Errorf("member not found")
	}
	if cur.Role == RoleOwner {
		n, err := s.countOwners(ctx, spaceID)
		if err != nil {
			return err
		}
		if n <= 1 {
			return fmt.Errorf("cannot remove the last owner")
		}
	}
	_, err = s.db.ExecContext(ctx, `
DELETE FROM space_members WHERE space_id = ? AND user_id = ?`,
		strings.TrimSpace(spaceID), strings.TrimSpace(userID))
	return err
}

// UpdateTeam renames a team in a space.
func (s *Store) UpdateTeam(ctx context.Context, spaceID, teamID, name string) (*models.Team, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	t, err := s.GetTeam(ctx, spaceID, teamID)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	_, err = s.db.ExecContext(ctx, `UPDATE teams SET name = ? WHERE id = ? AND space_id = ?`,
		name, strings.TrimSpace(teamID), strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	t.Name = name
	return t, nil
}

// GetAPIKeyByHash looks up a key by stored hash.
func (s *Store) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	var k models.APIKey
	err := s.db.QueryRowContext(ctx, `
SELECT id, user_id, created_at FROM api_keys WHERE key_hash = ?`, strings.TrimSpace(keyHash)).
		Scan(&k.ID, &k.UserID, &k.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &k, nil
}
