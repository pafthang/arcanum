package store

import (
	"context"
	"strings"

	"github.com/pafthang/arcanum/services/agents/models"
)

// SearchMemories returns memories whose key or value contains q (case-insensitive).
func (s *Store) SearchMemories(ctx context.Context, spaceID, agentID, q string) ([]models.Memory, error) {
	q = strings.TrimSpace(q)
	if q == "" {
		return s.ListMemories(ctx, spaceID, agentID)
	}
	like := "%" + q + "%"
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, agent_id, tier, key, value, updated_at
FROM memories
WHERE space_id = ? AND agent_id = ?
  AND (key LIKE ? COLLATE NOCASE OR value LIKE ? COLLATE NOCASE)
ORDER BY updated_at DESC
LIMIT 20`, strings.TrimSpace(spaceID), strings.TrimSpace(agentID), like, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Memory{}
	for rows.Next() {
		var m models.Memory
		if err := rows.Scan(&m.ID, &m.SpaceID, &m.AgentID, &m.Tier, &m.Key, &m.Value, &m.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
