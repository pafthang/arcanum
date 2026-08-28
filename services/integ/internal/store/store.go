package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/integ/models"
)

// Store is the integ SQLite database.
type Store struct {
	db *sql.DB
}

// OpenStore opens dataDir/integ.db and migrates.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "integ")
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
		`CREATE TABLE IF NOT EXISTS connectors (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			kind TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			config_json TEXT NOT NULL DEFAULT '{}',
			secret TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS github_repos (
			id TEXT PRIMARY KEY NOT NULL,
			connector_id TEXT NOT NULL,
			space_id TEXT NOT NULL,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			installation_id TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			FOREIGN KEY (connector_id) REFERENCES connectors(id)
		)`,
		`CREATE TABLE IF NOT EXISTS webhooks (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			url TEXT NOT NULL,
			secret TEXT NOT NULL DEFAULT '',
			events_json TEXT NOT NULL DEFAULT '[]',
			active INTEGER NOT NULL DEFAULT 1,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS deliveries (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			target_type TEXT NOT NULL,
			target_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			payload TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL,
			attempts INTEGER NOT NULL DEFAULT 1,
			last_error TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			delivered_at TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE TABLE IF NOT EXISTS inbound_events (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			connector_id TEXT NOT NULL,
			source TEXT NOT NULL,
			external_id TEXT NOT NULL DEFAULT '',
			payload_json TEXT NOT NULL DEFAULT '{}',
			issue_keys_json TEXT NOT NULL DEFAULT '[]',
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_connectors_space ON connectors(space_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_repos_connector ON github_repos(connector_id)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_repos_unique ON github_repos(connector_id, owner, name)`,
		`CREATE INDEX IF NOT EXISTS idx_webhooks_space ON webhooks(space_id)`,
		`CREATE INDEX IF NOT EXISTS idx_deliveries_space ON deliveries(space_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_inbound_connector ON inbound_events(connector_id, created_at)`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func marshalJSON(v any) string {
	if v == nil {
		return "{}"
	}
	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}

func unmarshalMap(raw string) map[string]any {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func unmarshalStrings(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []string{}
	}
	var out []string
	if err := json.Unmarshal([]byte(raw), &out); err != nil || out == nil {
		return []string{}
	}
	return out
}

func scanConnector(scan func(dest ...any) error) (*models.Connector, error) {
	var c models.Connector
	var cfg, secret string
	err := scan(&c.ID, &c.SpaceID, &c.Kind, &c.Name, &c.Status, &cfg, &secret, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.Config = unmarshalMap(cfg)
	c.Secret = secret
	c.HasSecret = secret != ""
	return &c, nil
}

func publicConnector(c *models.Connector) *models.Connector {
	if c == nil {
		return nil
	}
	out := *c
	out.Secret = ""
	return &out
}

// CreateConnector inserts a connector. Secret is stored as given.
func (s *Store) CreateConnector(ctx context.Context, spaceID, kind, name, status, secret string, cfg map[string]any) (*models.Connector, error) {
	spaceID = strings.TrimSpace(spaceID)
	kind = strings.TrimSpace(kind)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	if !models.ValidKind(kind) {
		return nil, fmt.Errorf("invalid kind")
	}
	status = strings.TrimSpace(status)
	if status == "" {
		status = models.StatusPending
	}
	if !models.ValidStatus(status) {
		return nil, fmt.Errorf("invalid status")
	}
	now := nowRFC3339()
	c := &models.Connector{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		Kind:      kind,
		Name:      name,
		Status:    status,
		Config:    cfg,
		Secret:    secret,
		HasSecret: secret != "",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if c.Config == nil {
		c.Config = map[string]any{}
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO connectors (id, space_id, kind, name, status, config_json, secret, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?)`,
		c.ID, c.SpaceID, c.Kind, c.Name, c.Status, marshalJSON(c.Config), c.Secret, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// GetConnector returns a connector by id (includes secret).
func (s *Store) GetConnector(ctx context.Context, id string) (*models.Connector, error) {
	return scanConnector(s.db.QueryRowContext(ctx, `
SELECT id, space_id, kind, name, status, config_json, secret, created_at, updated_at
FROM connectors WHERE id = ?`, strings.TrimSpace(id)).Scan)
}

// GetConnectorInSpace returns a connector scoped to a space.
func (s *Store) GetConnectorInSpace(ctx context.Context, spaceID, id string) (*models.Connector, error) {
	c, err := s.GetConnector(ctx, id)
	if err != nil || c == nil {
		return c, err
	}
	if c.SpaceID != strings.TrimSpace(spaceID) {
		return nil, nil
	}
	return c, nil
}

// ListConnectors returns connectors in a space, newest first.
func (s *Store) ListConnectors(ctx context.Context, spaceID string) ([]models.Connector, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, kind, name, status, config_json, secret, created_at, updated_at
FROM connectors WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Connector{}
	for rows.Next() {
		c, err := scanConnector(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

// UpdateConnector applies a partial update.
func (s *Store) UpdateConnector(ctx context.Context, id string, name, status, secret *string, cfg *map[string]any) (*models.Connector, error) {
	cur, err := s.GetConnector(ctx, id)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, nil
	}
	if name != nil {
		n := strings.TrimSpace(*name)
		if n == "" {
			return nil, fmt.Errorf("name required")
		}
		cur.Name = n
	}
	if status != nil {
		st := strings.TrimSpace(*status)
		if !models.ValidStatus(st) {
			return nil, fmt.Errorf("invalid status")
		}
		cur.Status = st
	}
	if cfg != nil {
		cur.Config = *cfg
		if cur.Config == nil {
			cur.Config = map[string]any{}
		}
	}
	if secret != nil {
		cur.Secret = *secret
		cur.HasSecret = cur.Secret != ""
	}
	cur.UpdatedAt = nowRFC3339()
	_, err = s.db.ExecContext(ctx, `
UPDATE connectors SET name=?, status=?, config_json=?, secret=?, updated_at=? WHERE id=?`,
		cur.Name, cur.Status, marshalJSON(cur.Config), cur.Secret, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

// CreateRepo links a GitHub repository to a github connector.
func (s *Store) CreateRepo(ctx context.Context, connectorID, spaceID, owner, name, installationID string) (*models.Repo, error) {
	owner = strings.TrimSpace(owner)
	name = strings.TrimSpace(name)
	if strings.TrimSpace(connectorID) == "" || strings.TrimSpace(spaceID) == "" || owner == "" || name == "" {
		return nil, fmt.Errorf("connector_id, space_id, owner and name required")
	}
	r := &models.Repo{
		ID:             idgen.New(),
		ConnectorID:    strings.TrimSpace(connectorID),
		SpaceID:        strings.TrimSpace(spaceID),
		Owner:          owner,
		Name:           name,
		InstallationID: strings.TrimSpace(installationID),
		CreatedAt:      nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO github_repos (id, connector_id, space_id, owner, name, installation_id, created_at)
VALUES (?,?,?,?,?,?,?)`,
		r.ID, r.ConnectorID, r.SpaceID, r.Owner, r.Name, r.InstallationID, r.CreatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// ListRepos returns repos for a connector.
func (s *Store) ListRepos(ctx context.Context, connectorID string) ([]models.Repo, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, connector_id, space_id, owner, name, installation_id, created_at
FROM github_repos WHERE connector_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(connectorID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Repo{}
	for rows.Next() {
		var r models.Repo
		if err := rows.Scan(&r.ID, &r.ConnectorID, &r.SpaceID, &r.Owner, &r.Name, &r.InstallationID, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CreateWebhook inserts an outbound webhook.
func (s *Store) CreateWebhook(ctx context.Context, spaceID, url, secret string, events []string, active bool) (*models.Webhook, error) {
	spaceID = strings.TrimSpace(spaceID)
	url = strings.TrimSpace(url)
	if spaceID == "" || url == "" {
		return nil, fmt.Errorf("space_id and url required")
	}
	if events == nil {
		events = []string{}
	}
	h := &models.Webhook{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		URL:       url,
		Events:    events,
		Active:    active,
		Secret:    secret,
		HasSecret: secret != "",
		CreatedAt: nowRFC3339(),
	}
	activeInt := 0
	if active {
		activeInt = 1
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO webhooks (id, space_id, url, secret, events_json, active, created_at)
VALUES (?,?,?,?,?,?,?)`,
		h.ID, h.SpaceID, h.URL, h.Secret, marshalJSON(h.Events), activeInt, h.CreatedAt)
	if err != nil {
		return nil, err
	}
	return h, nil
}

func scanWebhook(scan func(dest ...any) error) (*models.Webhook, error) {
	var h models.Webhook
	var events string
	var active int
	err := scan(&h.ID, &h.SpaceID, &h.URL, &h.Secret, &events, &active, &h.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	h.Events = unmarshalStrings(events)
	h.Active = active != 0
	h.HasSecret = h.Secret != ""
	return &h, nil
}

// ListWebhooks returns outbound webhooks for a space.
func (s *Store) ListWebhooks(ctx context.Context, spaceID string) ([]models.Webhook, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, url, secret, events_json, active, created_at
FROM webhooks WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Webhook{}
	for rows.Next() {
		h, err := scanWebhook(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *h)
	}
	return out, rows.Err()
}

func eventMatches(events []string, typ string) bool {
	if len(events) == 0 {
		return true
	}
	for _, e := range events {
		e = strings.TrimSpace(e)
		if e == "" || e == "*" || e == typ {
			return true
		}
	}
	return false
}

// ListActiveWebhooksForEvent returns active webhooks subscribed to typ.
func (s *Store) ListActiveWebhooksForEvent(ctx context.Context, spaceID, typ string) ([]models.Webhook, error) {
	all, err := s.ListWebhooks(ctx, spaceID)
	if err != nil {
		return nil, err
	}
	out := []models.Webhook{}
	for _, h := range all {
		if !h.Active {
			continue
		}
		if eventMatches(h.Events, typ) {
			out = append(out, h)
		}
	}
	return out, nil
}

// RecordDelivery stores an outbound delivery attempt.
func (s *Store) RecordDelivery(ctx context.Context, spaceID, targetType, targetID, eventType, payload, status, lastErr string, attempts int, deliveredAt string) (*models.Delivery, error) {
	if attempts < 1 {
		attempts = 1
	}
	d := &models.Delivery{
		ID:          idgen.New(),
		SpaceID:     strings.TrimSpace(spaceID),
		TargetType:  strings.TrimSpace(targetType),
		TargetID:    strings.TrimSpace(targetID),
		EventType:   strings.TrimSpace(eventType),
		Status:      status,
		Attempts:    attempts,
		LastError:   lastErr,
		CreatedAt:   nowRFC3339(),
		DeliveredAt: deliveredAt,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO deliveries (id, space_id, target_type, target_id, event_type, payload, status, attempts, last_error, created_at, delivered_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		d.ID, d.SpaceID, d.TargetType, d.TargetID, d.EventType, payload, d.Status, d.Attempts, d.LastError, d.CreatedAt, d.DeliveredAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// ListDeliveries returns outbound attempts for a space, newest first.
func (s *Store) ListDeliveries(ctx context.Context, spaceID string) ([]models.Delivery, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, target_type, target_id, event_type, status, attempts, last_error, created_at, delivered_at
FROM deliveries WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Delivery{}
	for rows.Next() {
		var d models.Delivery
		if err := rows.Scan(&d.ID, &d.SpaceID, &d.TargetType, &d.TargetID, &d.EventType, &d.Status, &d.Attempts, &d.LastError, &d.CreatedAt, &d.DeliveredAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

// RecordInbound stores a normalized inbound payload.
func (s *Store) RecordInbound(ctx context.Context, spaceID, connectorID, source, externalID string, payload map[string]any, issueKeys []string) (*models.InboundEvent, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	if issueKeys == nil {
		issueKeys = []string{}
	}
	ev := &models.InboundEvent{
		ID:          idgen.New(),
		SpaceID:     strings.TrimSpace(spaceID),
		ConnectorID: strings.TrimSpace(connectorID),
		Source:      strings.TrimSpace(source),
		ExternalID:  strings.TrimSpace(externalID),
		Payload:     payload,
		IssueKeys:   issueKeys,
		CreatedAt:   nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO inbound_events (id, space_id, connector_id, source, external_id, payload_json, issue_keys_json, created_at)
VALUES (?,?,?,?,?,?,?,?)`,
		ev.ID, ev.SpaceID, ev.ConnectorID, ev.Source, ev.ExternalID, marshalJSON(ev.Payload), marshalJSON(ev.IssueKeys), ev.CreatedAt)
	if err != nil {
		return nil, err
	}
	return ev, nil
}

// PublicConnector is the HTTP view (secret stripped).
func PublicConnector(c *models.Connector) *models.Connector {
	return publicConnector(c)
}

// PublicWebhook is the HTTP view (secret stripped).
func PublicWebhook(h *models.Webhook) *models.Webhook {
	if h == nil {
		return nil
	}
	out := *h
	out.Secret = ""
	return &out
}
