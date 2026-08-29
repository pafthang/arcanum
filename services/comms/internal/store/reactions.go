package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/comms/models"
)

func migrateReactions(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS comms_reactions (
			id TEXT PRIMARY KEY NOT NULL,
			message_id TEXT NOT NULL,
			space_id TEXT NOT NULL,
			actor_id TEXT NOT NULL,
			emoji TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (message_id) REFERENCES messages(id)
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_unique ON comms_reactions(message_id, actor_id, emoji)`,
		`CREATE INDEX IF NOT EXISTS idx_reactions_message ON comms_reactions(message_id)`,
	)
}

// AddReaction adds an emoji reaction to a message.
func (s *Store) AddReaction(ctx context.Context, messageID, spaceID, actorID, emoji string) (*models.Reaction, error) {
	messageID = strings.TrimSpace(messageID)
	spaceID = strings.TrimSpace(spaceID)
	actorID = strings.TrimSpace(actorID)
	emoji = strings.TrimSpace(emoji)
	if messageID == "" || spaceID == "" || actorID == "" || emoji == "" {
		return nil, fmt.Errorf("message_id, space_id, actor_id and emoji required")
	}
	r := &models.Reaction{
		ID:        idgen.New(),
		MessageID: messageID,
		SpaceID:   spaceID,
		ActorID:   actorID,
		Emoji:     emoji,
		CreatedAt: nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO comms_reactions (id, message_id, space_id, actor_id, emoji, created_at)
VALUES (?,?,?,?,?,?)`,
		r.ID, r.MessageID, r.SpaceID, r.ActorID, r.Emoji, r.CreatedAt)
	if err != nil {
		if sqldb.IsConstraintError(err) {
			return r, nil // Idempotent reaction
		}
		return nil, err
	}
	return r, nil
}

// ListReactions lists reactions for a message.
func (s *Store) ListReactions(ctx context.Context, messageID string) ([]models.Reaction, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, message_id, space_id, actor_id, emoji, created_at
FROM comms_reactions WHERE message_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(messageID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Reaction{}
	for rows.Next() {
		var r models.Reaction
		if err := rows.Scan(&r.ID, &r.MessageID, &r.SpaceID, &r.ActorID, &r.Emoji, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// RemoveReaction removes a user's emoji reaction.
func (s *Store) RemoveReaction(ctx context.Context, messageID, actorID, emoji string) error {
	_, err := s.db.ExecContext(ctx, `
DELETE FROM comms_reactions WHERE message_id = ? AND actor_id = ? AND emoji = ?`,
		strings.TrimSpace(messageID), strings.TrimSpace(actorID), strings.TrimSpace(emoji))
	return err
}
