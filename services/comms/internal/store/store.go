package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/comms/models"
)

// Store is the comms SQLite database.
type Store struct {
	db *sql.DB
}

// OpenStore opens dataDir/comms.db and migrates.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "comms")
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	return sqldb.Migrate(db,
		`CREATE TABLE IF NOT EXISTS channels (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			team_id TEXT NOT NULL DEFAULT '',
			name TEXT NOT NULL,
			kind TEXT NOT NULL CHECK (kind IN ('space','team','dm')),
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id TEXT PRIMARY KEY NOT NULL,
			channel_id TEXT NOT NULL,
			space_id TEXT NOT NULL,
			parent_id TEXT NOT NULL DEFAULT '',
			actor_id TEXT NOT NULL,
			body TEXT NOT NULL,
			blob_id TEXT NOT NULL DEFAULT '',
			source TEXT NOT NULL CHECK (source IN ('user','agent','integ')) DEFAULT 'user',
			external_ref TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY (channel_id) REFERENCES channels(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_space ON channels(space_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_channels_team ON channels(space_id, team_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_channel ON messages(channel_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_parent ON messages(channel_id, parent_id, created_at)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_messages_external ON messages(channel_id, external_ref) WHERE external_ref != ''`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// ValidKind reports whether kind is allowed.
func ValidKind(kind string) bool {
	switch kind {
	case models.KindSpace, models.KindTeam, models.KindDM:
		return true
	}
	return false
}

// ValidSource reports whether source is allowed.
func ValidSource(source string) bool {
	switch source {
	case models.SourceUser, models.SourceAgent, models.SourceInteg:
		return true
	}
	return false
}

func normalizeKind(kind, teamID string) (string, error) {
	kind = strings.TrimSpace(kind)
	if kind == "" {
		if teamID != "" {
			kind = models.KindTeam
		} else {
			kind = models.KindSpace
		}
	}
	if !ValidKind(kind) {
		return "", fmt.Errorf("invalid kind")
	}
	if kind == models.KindTeam && teamID == "" {
		return "", fmt.Errorf("team_id required for team channel")
	}
	return kind, nil
}

// CreateChannel inserts a channel.
func (s *Store) CreateChannel(ctx context.Context, spaceID, name, teamID, kind string) (*models.Channel, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	teamID = strings.TrimSpace(teamID)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	kind, err := normalizeKind(kind, teamID)
	if err != nil {
		return nil, err
	}
	now := nowRFC3339()
	ch := &models.Channel{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		TeamID:    teamID,
		Name:      name,
		Kind:      kind,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO channels (id, space_id, team_id, name, kind, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)`,
		ch.ID, ch.SpaceID, ch.TeamID, ch.Name, ch.Kind, ch.CreatedAt, ch.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func scanChannel(row *sql.Row) (*models.Channel, error) {
	var ch models.Channel
	err := row.Scan(&ch.ID, &ch.SpaceID, &ch.TeamID, &ch.Name, &ch.Kind, &ch.CreatedAt, &ch.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ch, nil
}

// GetChannel returns a channel by id.
func (s *Store) GetChannel(ctx context.Context, id string) (*models.Channel, error) {
	return scanChannel(s.db.QueryRowContext(ctx, `
SELECT id, space_id, team_id, name, kind, created_at, updated_at
FROM channels WHERE id = ?`, strings.TrimSpace(id)))
}

// GetChannelInSpace returns a channel if it belongs to spaceID.
func (s *Store) GetChannelInSpace(ctx context.Context, spaceID, id string) (*models.Channel, error) {
	ch, err := s.GetChannel(ctx, id)
	if err != nil || ch == nil {
		return ch, err
	}
	if ch.SpaceID != strings.TrimSpace(spaceID) {
		return nil, nil
	}
	return ch, nil
}

// ListChannels returns channels in a space, newest first. teamID filters when set.
func (s *Store) ListChannels(ctx context.Context, spaceID, teamID string) ([]models.Channel, error) {
	spaceID = strings.TrimSpace(spaceID)
	teamID = strings.TrimSpace(teamID)
	var (
		rows *sql.Rows
		err  error
	)
	if teamID == "" {
		rows, err = s.db.QueryContext(ctx, `
SELECT id, space_id, team_id, name, kind, created_at, updated_at
FROM channels WHERE space_id = ?
ORDER BY created_at DESC`, spaceID)
	} else {
		rows, err = s.db.QueryContext(ctx, `
SELECT id, space_id, team_id, name, kind, created_at, updated_at
FROM channels WHERE space_id = ? AND team_id = ?
ORDER BY created_at DESC`, spaceID, teamID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Channel{}
	for rows.Next() {
		var ch models.Channel
		if err := rows.Scan(&ch.ID, &ch.SpaceID, &ch.TeamID, &ch.Name, &ch.Kind, &ch.CreatedAt, &ch.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, ch)
	}
	return out, rows.Err()
}

// FindChannelByName returns the first channel with that name in the space.
func (s *Store) FindChannelByName(ctx context.Context, spaceID, name string) (*models.Channel, error) {
	return scanChannel(s.db.QueryRowContext(ctx, `
SELECT id, space_id, team_id, name, kind, created_at, updated_at
FROM channels WHERE space_id = ? AND name = ?
ORDER BY created_at ASC LIMIT 1`, strings.TrimSpace(spaceID), strings.TrimSpace(name)))
}

// CreateMessage inserts a message. Empty body is allowed only with blob_id (attachment-only).
func (s *Store) CreateMessage(ctx context.Context, channelID, spaceID, actorID, body, parentID, blobID, source, externalRef string) (*models.Message, error) {
	channelID = strings.TrimSpace(channelID)
	spaceID = strings.TrimSpace(spaceID)
	actorID = strings.TrimSpace(actorID)
	parentID = strings.TrimSpace(parentID)
	blobID = strings.TrimSpace(blobID)
	externalRef = strings.TrimSpace(externalRef)
	source = strings.TrimSpace(source)
	if source == "" {
		source = models.SourceUser
	}
	if !ValidSource(source) {
		return nil, fmt.Errorf("invalid source")
	}
	if channelID == "" || spaceID == "" || actorID == "" {
		return nil, fmt.Errorf("channel_id, space_id and actor_id required")
	}
	if strings.TrimSpace(body) == "" && blobID == "" {
		return nil, fmt.Errorf("body or blob_id required")
	}
	if parentID != "" {
		parent, err := s.GetMessage(ctx, parentID)
		if err != nil {
			return nil, err
		}
		if parent == nil || parent.ChannelID != channelID {
			return nil, fmt.Errorf("parent not found")
		}
	}
	msg := &models.Message{
		ID:          idgen.New(),
		ChannelID:   channelID,
		SpaceID:     spaceID,
		ParentID:    parentID,
		ActorID:     actorID,
		Body:        body,
		BlobID:      blobID,
		Source:      source,
		ExternalRef: externalRef,
		CreatedAt:   nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO messages (id, channel_id, space_id, parent_id, actor_id, body, blob_id, source, external_ref, created_at)
VALUES (?,?,?,?,?,?,?,?,?,?)`,
		msg.ID, msg.ChannelID, msg.SpaceID, msg.ParentID, msg.ActorID, msg.Body, msg.BlobID, msg.Source, msg.ExternalRef, msg.CreatedAt)
	if err != nil {
		if sqldb.IsConstraintError(err) && externalRef != "" {
			existing, gerr := s.GetMessageByExternal(ctx, channelID, externalRef)
			if gerr != nil {
				return nil, gerr
			}
			if existing != nil {
				return existing, nil
			}
		}
		return nil, err
	}
	return msg, nil
}

func scanMessage(row *sql.Row) (*models.Message, error) {
	var m models.Message
	err := row.Scan(&m.ID, &m.ChannelID, &m.SpaceID, &m.ParentID, &m.ActorID, &m.Body, &m.BlobID, &m.Source, &m.ExternalRef, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// GetMessage returns a message by id.
func (s *Store) GetMessage(ctx context.Context, id string) (*models.Message, error) {
	return scanMessage(s.db.QueryRowContext(ctx, `
SELECT id, channel_id, space_id, parent_id, actor_id, body, blob_id, source, external_ref, created_at
FROM messages WHERE id = ?`, strings.TrimSpace(id)))
}

// GetMessageByExternal returns a message by channel + external ref.
func (s *Store) GetMessageByExternal(ctx context.Context, channelID, externalRef string) (*models.Message, error) {
	return scanMessage(s.db.QueryRowContext(ctx, `
SELECT id, channel_id, space_id, parent_id, actor_id, body, blob_id, source, external_ref, created_at
FROM messages WHERE channel_id = ? AND external_ref = ?`, strings.TrimSpace(channelID), strings.TrimSpace(externalRef)))
}

// MessageListFilter selects messages in a channel.
type MessageListFilter struct {
	ChannelID string
	ParentID  string
	Before    string
	Limit     int
}

// ListMessages returns messages oldest-first within the window (chat history).
func (s *Store) ListMessages(ctx context.Context, f MessageListFilter) ([]models.Message, error) {
	if f.Limit <= 0 {
		f.Limit = 50
	}
	args := []any{strings.TrimSpace(f.ChannelID)}
	q := `
SELECT id, channel_id, space_id, parent_id, actor_id, body, blob_id, source, external_ref, created_at
FROM messages WHERE channel_id = ?`
	if f.ParentID != "" {
		q += ` AND parent_id = ?`
		args = append(args, strings.TrimSpace(f.ParentID))
	}
	if f.Before != "" {
		q += ` AND created_at < ?`
		args = append(args, strings.TrimSpace(f.Before))
	}
	q += ` ORDER BY created_at ASC LIMIT ?`
	args = append(args, f.Limit)

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Message{}
	for rows.Next() {
		var m models.Message
		if err := rows.Scan(&m.ID, &m.ChannelID, &m.SpaceID, &m.ParentID, &m.ActorID, &m.Body, &m.BlobID, &m.Source, &m.ExternalRef, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
