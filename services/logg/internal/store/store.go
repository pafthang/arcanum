package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/logg/models"
	_ "modernc.org/sqlite" // sqlite driver
)

// Store manages the logg SQLite database.
type Store struct {
	db *sql.DB
}

// LogEntry represents a single log line.
type LogEntry struct {
	ID      int64  `json:"id"`
	Service string `json:"service"`
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"message"`
	Raw     string `json:"raw"` // Raw JSON of the log for extended attributes
}

// OpenStore opens or initializes the SQLite database for the logg service.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "logg")
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &Store{db: db}, nil
}

func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		service TEXT NOT NULL,
		time TEXT NOT NULL,
		level TEXT NOT NULL,
		message TEXT NOT NULL,
		raw TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_logs_time ON logs(time);
	CREATE INDEX IF NOT EXISTS idx_logs_service ON logs(service);
	CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);

	CREATE TABLE IF NOT EXISTS activity (
		id TEXT PRIMARY KEY NOT NULL,
		space_id TEXT NOT NULL,
		target_type TEXT NOT NULL DEFAULT '',
		target_id TEXT NOT NULL DEFAULT '',
		actor_id TEXT NOT NULL DEFAULT '',
		type TEXT NOT NULL,
		summary TEXT NOT NULL DEFAULT '',
		payload_json TEXT NOT NULL DEFAULT '{}',
		created TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS activity_target_created_idx ON activity (target_type, target_id, created DESC);
	CREATE INDEX IF NOT EXISTS activity_space_created_idx ON activity (space_id, created DESC);
	`
	_, err := db.Exec(schema)
	return err
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying sql.DB.
func (s *Store) DB() *sql.DB {
	return s.db
}

// ──────────────────────────────────────────────────────────────────────────────
// Logs
// ──────────────────────────────────────────────────────────────────────────────

// InsertLog inserts a new log entry.
func (s *Store) InsertLog(service, time, level, message, raw string) error {
	_, err := s.db.Exec(`INSERT INTO logs (service, time, level, message, raw) VALUES (?, ?, ?, ?, ?)`, service, time, level, message, raw)
	return err
}

// ListLogs retrieves a paginated list of logs, optionally filtered by service and level.
func (s *Store) ListLogs(service, level string, limit, offset int) ([]LogEntry, error) {
	query := `SELECT id, service, time, level, message, raw FROM logs WHERE 1=1`
	args := []any{}

	if service != "" {
		query += ` AND service = ?`
		args = append(args, service)
	}
	if level != "" {
		query += ` AND level = ?`
		args = append(args, level)
	}

	query += ` ORDER BY time DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []LogEntry
	for rows.Next() {
		var l LogEntry
		if err := rows.Scan(&l.ID, &l.Service, &l.Time, &l.Level, &l.Message, &l.Raw); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// ──────────────────────────────────────────────────────────────────────────────
// logg.Activity
// ──────────────────────────────────────────────────────────────────────────────

// AppendActivity inserts a durable activity row.
func (s *Store) AppendActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	if a == nil {
		return nil, fmt.Errorf("activity required")
	}
	a.SpaceID = strings.TrimSpace(a.SpaceID)
	a.Type = strings.TrimSpace(a.Type)
	if a.SpaceID == "" || a.Type == "" {
		return nil, fmt.Errorf("spaceId and type required")
	}
	if a.Payload == nil {
		a.Payload = map[string]any{}
	}
	raw, err := json.Marshal(a.Payload)
	if err != nil {
		return nil, err
	}
	_, err = s.db.ExecContext(ctx, `
INSERT INTO activity (id, space_id, target_type, target_id, actor_id, type, summary, payload_json, created)
VALUES (?,?,?,?,?,?,?,?,?)`,
		a.ID, a.SpaceID, a.TargetType, a.TargetID, a.ActorID, a.Type, a.Summary, string(raw), a.Created,
	)
	if err != nil {
		return nil, err
	}
	return a, nil
}

// ListTargetActivity returns recent activity for a specific target (newest first).
func (s *Store) ListTargetActivity(ctx context.Context, spaceID, targetType, targetID string, limit int) ([]models.Activity, error) {
	spaceID = strings.TrimSpace(spaceID)
	targetID = strings.TrimSpace(targetID)
	if spaceID == "" || targetID == "" {
		return nil, fmt.Errorf("spaceId and targetId required")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, target_type, target_id, actor_id, type, summary, payload_json, created
FROM activity
WHERE space_id = ? AND target_type = ? AND target_id = ?
ORDER BY created DESC
LIMIT ?`, spaceID, targetType, targetID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanActivities(rows)
}

// ListTeamActivity returns team activity newest first, with optional filters + pagination.
func (s *Store) ListTeamActivity(ctx context.Context, f models.ActivityListFilter, page, perPage int) ([]models.Activity, int, error) {
	f.SpaceID = strings.TrimSpace(f.SpaceID)
	if f.SpaceID == "" {
		return nil, 0, fmt.Errorf("spaceId required")
	}
	if page < 1 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 50
	}
	if perPage > 200 {
		perPage = 200
	}

	where := []string{"space_id = ?"}
	args := []any{f.SpaceID}

	if tt := strings.TrimSpace(f.TargetType); tt != "" {
		where = append(where, "target_type = ?")
		args = append(args, tt)
	}
	if tid := strings.TrimSpace(f.TargetID); tid != "" {
		where = append(where, "target_id = ?")
		args = append(args, tid)
	}
	if actor := strings.TrimSpace(f.ActorID); actor != "" {
		where = append(where, "actor_id = ?")
		args = append(args, actor)
	}
	if typ := strings.TrimSpace(f.Type); typ != "" {
		// "task" / "project" → prefix; "task.created" → exact
		if strings.Contains(typ, ".") {
			where = append(where, "type = ?")
			args = append(args, typ)
		} else {
			where = append(where, "(type = ? OR type LIKE ?)")
			args = append(args, typ, typ+".%")
		}
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		where = append(where, "LOWER(summary) LIKE ?")
		args = append(args, "%"+strings.ToLower(q)+"%")
	}

	clause := strings.Join(where, " AND ")
	var total int
	if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM activity WHERE "+clause, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * perPage
	queryArgs := append(append([]any{}, args...), perPage, offset)
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, target_type, target_id, actor_id, type, summary, payload_json, created
FROM activity
WHERE `+clause+`
ORDER BY created DESC
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items, err := scanActivities(rows)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func scanActivities(rows *sql.Rows) ([]models.Activity, error) {
	var out []models.Activity
	for rows.Next() {
		var a models.Activity
		var payload string
		if err := rows.Scan(
			&a.ID, &a.SpaceID, &a.TargetType, &a.TargetID, &a.ActorID, &a.Type, &a.Summary, &payload, &a.Created,
		); err != nil {
			return nil, err
		}
		if payload == "" {
			payload = "{}"
		}
		_ = json.Unmarshal([]byte(payload), &a.Payload)
		if a.Payload == nil {
			a.Payload = map[string]any{}
		}
		out = append(out, a)
	}
	if out == nil {
		out = []models.Activity{}
	}
	return out, rows.Err()
}
