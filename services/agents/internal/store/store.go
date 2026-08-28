package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/agents/models"
)

// Store is the agents SQLite database.
type Store struct {
	db *sql.DB
}

// OpenStore opens dataDir/agents.db and migrates.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "agents")
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
		`CREATE TABLE IF NOT EXISTS runs (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			issue_id TEXT NOT NULL DEFAULT '',
			agent_id TEXT NOT NULL,
			status TEXT NOT NULL,
			input TEXT NOT NULL DEFAULT '',
			output TEXT NOT NULL DEFAULT '',
			error TEXT NOT NULL DEFAULT '',
			started_at TEXT NOT NULL DEFAULT '',
			finished_at TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY NOT NULL,
			run_id TEXT NOT NULL UNIQUE,
			space_id TEXT NOT NULL,
			stage TEXT NOT NULL DEFAULT '',
			payload TEXT NOT NULL DEFAULT '',
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			FOREIGN KEY (run_id) REFERENCES runs(id)
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			agent_id TEXT NOT NULL,
			tier TEXT NOT NULL,
			key TEXT NOT NULL,
			value TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			UNIQUE (space_id, agent_id, key)
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			name TEXT NOT NULL,
			body TEXT NOT NULL,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_space ON runs(space_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_runs_agent ON runs(agent_id, created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_memories_agent ON memories(space_id, agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_skills_space ON skills(space_id, created_at)`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

// ValidStatus reports whether a run status is allowed.
func ValidStatus(status string) bool {
	switch status {
	case models.StatusQueued, models.StatusRunning, models.StatusCancelling,
		models.StatusCancelled, models.StatusSucceeded, models.StatusFailed:
		return true
	}
	return false
}

func terminal(status string) bool {
	switch status {
	case models.StatusCancelled, models.StatusSucceeded, models.StatusFailed:
		return true
	}
	return false
}

func scanRun(scan func(dest ...any) error) (*models.Run, error) {
	var r models.Run
	err := scan(&r.ID, &r.SpaceID, &r.IssueID, &r.AgentID, &r.Status, &r.Input, &r.Output, &r.Error, &r.StartedAt, &r.FinishedAt, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// CreateRun inserts a queued run.
func (s *Store) CreateRun(ctx context.Context, spaceID, agentID, issueID, input string) (*models.Run, error) {
	spaceID = strings.TrimSpace(spaceID)
	agentID = strings.TrimSpace(agentID)
	if spaceID == "" || agentID == "" {
		return nil, fmt.Errorf("space_id and agent_id required")
	}
	now := nowRFC3339()
	r := &models.Run{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		IssueID:   strings.TrimSpace(issueID),
		AgentID:   agentID,
		Status:    models.StatusQueued,
		Input:     input,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO runs (id, space_id, issue_id, agent_id, status, input, output, error, started_at, finished_at, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		r.ID, r.SpaceID, r.IssueID, r.AgentID, r.Status, r.Input, r.Output, r.Error, r.StartedAt, r.FinishedAt, r.CreatedAt, r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// GetRun returns a run by id.
func (s *Store) GetRun(ctx context.Context, id string) (*models.Run, error) {
	return scanRun(s.db.QueryRowContext(ctx, `
SELECT id, space_id, issue_id, agent_id, status, input, output, error, started_at, finished_at, created_at, updated_at
FROM runs WHERE id = ?`, strings.TrimSpace(id)).Scan)
}

// GetRunInSpace returns a run if it belongs to spaceID.
func (s *Store) GetRunInSpace(ctx context.Context, spaceID, id string) (*models.Run, error) {
	r, err := s.GetRun(ctx, id)
	if err != nil || r == nil {
		return r, err
	}
	if r.SpaceID != strings.TrimSpace(spaceID) {
		return nil, nil
	}
	return r, nil
}

// ListRuns returns runs in a space, newest first.
func (s *Store) ListRuns(ctx context.Context, spaceID string) ([]models.Run, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, issue_id, agent_id, status, input, output, error, started_at, finished_at, created_at, updated_at
FROM runs WHERE space_id = ?
ORDER BY created_at DESC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Run{}
	for rows.Next() {
		r, err := scanRun(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, *r)
	}
	return out, rows.Err()
}

// MarkRunning sets status to running unless already cancelling/terminal.
func (s *Store) MarkRunning(ctx context.Context, id string) (*models.Run, error) {
	cur, err := s.GetRun(ctx, id)
	if err != nil || cur == nil {
		return cur, err
	}
	if cur.Status == models.StatusCancelling || terminal(cur.Status) {
		return cur, nil
	}
	now := nowRFC3339()
	if cur.StartedAt == "" {
		cur.StartedAt = now
	}
	cur.Status = models.StatusRunning
	cur.UpdatedAt = now
	_, err = s.db.ExecContext(ctx, `
UPDATE runs SET status=?, started_at=?, updated_at=? WHERE id=?`,
		cur.Status, cur.StartedAt, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

// FinishRun sets a terminal status, output and error.
func (s *Store) FinishRun(ctx context.Context, id, status, output, errText string) (*models.Run, error) {
	if !terminal(status) {
		return nil, fmt.Errorf("invalid finish status")
	}
	cur, err := s.GetRun(ctx, id)
	if err != nil || cur == nil {
		return cur, err
	}
	now := nowRFC3339()
	cur.Status = status
	cur.Output = output
	cur.Error = errText
	cur.FinishedAt = now
	cur.UpdatedAt = now
	_, err = s.db.ExecContext(ctx, `
UPDATE runs SET status=?, output=?, error=?, finished_at=?, updated_at=? WHERE id=?`,
		cur.Status, cur.Output, cur.Error, cur.FinishedAt, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

// CancelRun marks a non-terminal run as cancelling.
func (s *Store) CancelRun(ctx context.Context, id string) (*models.Run, error) {
	cur, err := s.GetRun(ctx, id)
	if err != nil || cur == nil {
		return cur, err
	}
	if terminal(cur.Status) {
		return cur, nil
	}
	now := nowRFC3339()
	cur.Status = models.StatusCancelling
	cur.UpdatedAt = now
	_, err = s.db.ExecContext(ctx, `
UPDATE runs SET status=?, updated_at=? WHERE id=?`, cur.Status, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func scanSession(scan func(dest ...any) error) (*models.Session, error) {
	var sess models.Session
	err := scan(&sess.ID, &sess.RunID, &sess.SpaceID, &sess.Stage, &sess.Payload, &sess.CreatedAt, &sess.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sess, nil
}

// GetSession returns the session for a run.
func (s *Store) GetSession(ctx context.Context, runID string) (*models.Session, error) {
	return scanSession(s.db.QueryRowContext(ctx, `
SELECT id, run_id, space_id, stage, payload, created_at, updated_at
FROM sessions WHERE run_id = ?`, strings.TrimSpace(runID)).Scan)
}

// UpsertSession writes pipeline payload for a run.
func (s *Store) UpsertSession(ctx context.Context, runID, spaceID, stage, payload string) (*models.Session, error) {
	runID = strings.TrimSpace(runID)
	spaceID = strings.TrimSpace(spaceID)
	if runID == "" || spaceID == "" {
		return nil, fmt.Errorf("run_id and space_id required")
	}
	now := nowRFC3339()
	cur, err := s.GetSession(ctx, runID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		cur = &models.Session{
			ID:        idgen.New(),
			RunID:     runID,
			SpaceID:   spaceID,
			Stage:     stage,
			Payload:   payload,
			CreatedAt: now,
			UpdatedAt: now,
		}
		_, err = s.db.ExecContext(ctx, `
INSERT INTO sessions (id, run_id, space_id, stage, payload, created_at, updated_at)
VALUES (?,?,?,?,?,?,?)`,
			cur.ID, cur.RunID, cur.SpaceID, cur.Stage, cur.Payload, cur.CreatedAt, cur.UpdatedAt)
		if err != nil {
			return nil, err
		}
		return cur, nil
	}
	cur.Stage = stage
	cur.Payload = payload
	cur.UpdatedAt = now
	_, err = s.db.ExecContext(ctx, `
UPDATE sessions SET stage=?, payload=?, updated_at=? WHERE id=?`,
		cur.Stage, cur.Payload, cur.UpdatedAt, cur.ID)
	if err != nil {
		return nil, err
	}
	return cur, nil
}

func validTier(tier string) bool {
	switch tier {
	case models.TierWorking, models.TierEpisodic, models.TierSemantic:
		return true
	}
	return false
}

// ListMemories returns memories for an agent in a space.
func (s *Store) ListMemories(ctx context.Context, spaceID, agentID string) ([]models.Memory, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, agent_id, tier, key, value, updated_at
FROM memories WHERE space_id = ? AND agent_id = ?
ORDER BY key ASC`, strings.TrimSpace(spaceID), strings.TrimSpace(agentID))
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

// PutMemory upserts a memory row by (space, agent, key).
func (s *Store) PutMemory(ctx context.Context, spaceID, agentID, tier, key, value string) (*models.Memory, error) {
	spaceID = strings.TrimSpace(spaceID)
	agentID = strings.TrimSpace(agentID)
	key = strings.TrimSpace(key)
	tier = strings.TrimSpace(tier)
	if spaceID == "" || agentID == "" || key == "" {
		return nil, fmt.Errorf("space_id, agent_id and key required")
	}
	if tier == "" {
		tier = models.TierWorking
	}
	if !validTier(tier) {
		return nil, fmt.Errorf("invalid memory tier")
	}
	now := nowRFC3339()
	_, err := s.db.ExecContext(ctx, `
INSERT INTO memories (id, space_id, agent_id, tier, key, value, updated_at)
VALUES (?,?,?,?,?,?,?)
ON CONFLICT (space_id, agent_id, key) DO UPDATE SET
	tier=excluded.tier, value=excluded.value, updated_at=excluded.updated_at`,
		idgen.New(), spaceID, agentID, tier, key, value, now)
	if err != nil {
		return nil, err
	}
	var m models.Memory
	err = s.db.QueryRowContext(ctx, `
SELECT id, space_id, agent_id, tier, key, value, updated_at
FROM memories WHERE space_id = ? AND agent_id = ? AND key = ?`, spaceID, agentID, key).
		Scan(&m.ID, &m.SpaceID, &m.AgentID, &m.Tier, &m.Key, &m.Value, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListSkills returns skills in a space.
func (s *Store) ListSkills(ctx context.Context, spaceID string) ([]models.Skill, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, space_id, name, body, created_at
FROM skills WHERE space_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Skill{}
	for rows.Next() {
		var sk models.Skill
		if err := rows.Scan(&sk.ID, &sk.SpaceID, &sk.Name, &sk.Body, &sk.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sk)
	}
	return out, rows.Err()
}

// CreateSkill inserts a skill.
func (s *Store) CreateSkill(ctx context.Context, spaceID, name, body string) (*models.Skill, error) {
	spaceID = strings.TrimSpace(spaceID)
	name = strings.TrimSpace(name)
	if spaceID == "" || name == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	sk := &models.Skill{
		ID:        idgen.New(),
		SpaceID:   spaceID,
		Name:      name,
		Body:      body,
		CreatedAt: nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO skills (id, space_id, name, body, created_at) VALUES (?,?,?,?,?)`,
		sk.ID, sk.SpaceID, sk.Name, sk.Body, sk.CreatedAt)
	if err != nil {
		return nil, err
	}
	return sk, nil
}
