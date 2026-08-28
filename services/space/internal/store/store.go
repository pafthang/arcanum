package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/passwd"
	"github.com/pafthang/arcanum/pkg/sqldb"
	"github.com/pafthang/arcanum/services/space/models"
)

const (
	ActorUser  = "user"
	ActorAgent = "agent"

	RoleOwner  = "owner"
	RoleAdmin  = "admin"
	RoleMember = "member"
	RoleGuest  = "guest"

	DefaultSpaceID = "default"
)

// Store is the space SQLite database.
type Store struct {
	db *sql.DB
}

// UserRecord is a user row including the password hash (never serialized).
type UserRecord struct {
	models.User
	PasswordHash string
}

// OpenStore opens dataDir/space.db and migrates.
func OpenStore(dataDir string) (*Store, error) {
	db, err := sqldb.Open(dataDir, "space")
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
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY NOT NULL,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			actor TEXT NOT NULL CHECK (actor IN ('user','agent')),
			platform_admin INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS spaces (
			id TEXT PRIMARY KEY NOT NULL,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS space_members (
			space_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL CHECK (role IN ('owner','admin','member','guest')),
			created_at TEXT NOT NULL,
			PRIMARY KEY (space_id, user_id),
			FOREIGN KEY (space_id) REFERENCES spaces(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS teams (
			id TEXT PRIMARY KEY NOT NULL,
			space_id TEXT NOT NULL,
			parent_id TEXT,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (space_id) REFERENCES spaces(id)
		)`,
		`CREATE TABLE IF NOT EXISTS team_members (
			team_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			role TEXT NOT NULL,
			PRIMARY KEY (team_id, user_id),
			FOREIGN KEY (team_id) REFERENCES teams(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY NOT NULL,
			user_id TEXT NOT NULL,
			key_hash TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_space_members_user ON space_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_teams_space ON teams(space_id)`,
		`CREATE INDEX IF NOT EXISTS idx_api_keys_user ON api_keys(user_id)`,
	)
}

// Close closes the database.
func (s *Store) Close() error {
	return s.db.Close()
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func validActor(a string) bool {
	return a == ActorUser || a == ActorAgent
}

func validSpaceRole(r string) bool {
	switch r {
	case RoleOwner, RoleAdmin, RoleMember, RoleGuest:
		return true
	}
	return false
}

// CreateUser inserts a user. passwordHash must already be argon2id encoded.
func (s *Store) CreateUser(ctx context.Context, email, passwordHash, actor string, platformAdmin bool) (*UserRecord, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	actor = strings.TrimSpace(actor)
	if email == "" {
		return nil, fmt.Errorf("email required")
	}
	if !validActor(actor) {
		return nil, fmt.Errorf("actor must be user or agent")
	}
	u := &UserRecord{
		User: models.User{
			ID:            idgen.New(),
			Email:         email,
			Actor:         actor,
			PlatformAdmin: platformAdmin,
			CreatedAt:     nowRFC3339(),
		},
		PasswordHash: passwordHash,
	}
	admin := 0
	if platformAdmin {
		admin = 1
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO users (id, email, password_hash, actor, platform_admin, created_at)
VALUES (?,?,?,?,?,?)`, u.ID, u.Email, u.PasswordHash, u.Actor, admin, u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// GetUser returns a user by id.
func (s *Store) GetUser(ctx context.Context, id string) (*UserRecord, error) {
	return s.scanUser(s.db.QueryRowContext(ctx, `
SELECT id, email, password_hash, actor, platform_admin, created_at
FROM users WHERE id = ?`, strings.TrimSpace(id)))
}

// GetUserByEmail returns a user by email (case-insensitive).
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*UserRecord, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	return s.scanUser(s.db.QueryRowContext(ctx, `
SELECT id, email, password_hash, actor, platform_admin, created_at
FROM users WHERE email = ?`, email))
}

func (s *Store) scanUser(row *sql.Row) (*UserRecord, error) {
	var u UserRecord
	var admin int
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Actor, &admin, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.PlatformAdmin = admin != 0
	return &u, nil
}

// CreateSpace inserts a space. id may be empty (generated) or fixed (seed default).
func (s *Store) CreateSpace(ctx context.Context, id, name string) (*models.Space, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name required")
	}
	id = strings.TrimSpace(id)
	if id == "" {
		id = idgen.New()
	}
	now := nowRFC3339()
	sp := &models.Space{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO spaces (id, name, created_at, updated_at) VALUES (?,?,?,?)`,
		sp.ID, sp.Name, sp.CreatedAt, sp.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return sp, nil
}

// GetSpace returns a space by id.
func (s *Store) GetSpace(ctx context.Context, id string) (*models.Space, error) {
	var sp models.Space
	err := s.db.QueryRowContext(ctx, `
SELECT id, name, created_at, updated_at FROM spaces WHERE id = ?`, strings.TrimSpace(id)).
		Scan(&sp.ID, &sp.Name, &sp.CreatedAt, &sp.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sp, nil
}

// ListSpacesForUser returns spaces the user belongs to, with role.
func (s *Store) ListSpacesForUser(ctx context.Context, userID string) ([]models.SpaceWithRole, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT s.id, s.name, s.created_at, s.updated_at, m.role
FROM spaces s
JOIN space_members m ON m.space_id = s.id
WHERE m.user_id = ?
ORDER BY s.created_at ASC`, strings.TrimSpace(userID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.SpaceWithRole{}
	for rows.Next() {
		var item models.SpaceWithRole
		if err := rows.Scan(&item.ID, &item.Name, &item.CreatedAt, &item.UpdatedAt, &item.Role); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// AddMember adds a space membership.
func (s *Store) AddMember(ctx context.Context, spaceID, userID, role string) (*models.Member, error) {
	role = strings.TrimSpace(role)
	if !validSpaceRole(role) {
		return nil, fmt.Errorf("invalid space role")
	}
	m := &models.Member{
		SpaceID:   strings.TrimSpace(spaceID),
		UserID:    strings.TrimSpace(userID),
		Role:      role,
		CreatedAt: nowRFC3339(),
	}
	if m.SpaceID == "" || m.UserID == "" {
		return nil, fmt.Errorf("space_id and user_id required")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO space_members (space_id, user_id, role, created_at) VALUES (?,?,?,?)`,
		m.SpaceID, m.UserID, m.Role, m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// GetMember returns a membership or nil.
func (s *Store) GetMember(ctx context.Context, spaceID, userID string) (*models.Member, error) {
	var m models.Member
	err := s.db.QueryRowContext(ctx, `
SELECT space_id, user_id, role, created_at
FROM space_members WHERE space_id = ? AND user_id = ?`,
		strings.TrimSpace(spaceID), strings.TrimSpace(userID)).
		Scan(&m.SpaceID, &m.UserID, &m.Role, &m.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &m, nil
}

// ListMembers returns members of a space.
func (s *Store) ListMembers(ctx context.Context, spaceID string) ([]models.Member, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT space_id, user_id, role, created_at
FROM space_members WHERE space_id = ?
ORDER BY created_at ASC`, strings.TrimSpace(spaceID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []models.Member{}
	for rows.Next() {
		var m models.Member
		if err := rows.Scan(&m.SpaceID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// CreateTeam inserts a nested team.
func (s *Store) CreateTeam(ctx context.Context, spaceID, parentID, name string) (*models.Team, error) {
	name = strings.TrimSpace(name)
	if name == "" || strings.TrimSpace(spaceID) == "" {
		return nil, fmt.Errorf("space_id and name required")
	}
	t := &models.Team{
		ID:        idgen.New(),
		SpaceID:   strings.TrimSpace(spaceID),
		ParentID:  strings.TrimSpace(parentID),
		Name:      name,
		CreatedAt: nowRFC3339(),
	}
	var parent any
	if t.ParentID != "" {
		parent = t.ParentID
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO teams (id, space_id, parent_id, name, created_at) VALUES (?,?,?,?,?)`,
		t.ID, t.SpaceID, parent, t.Name, t.CreatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// AddTeamMember attaches a user to a team.
func (s *Store) AddTeamMember(ctx context.Context, teamID, userID, role string) (*models.TeamMember, error) {
	role = strings.TrimSpace(role)
	if role == "" {
		role = RoleMember
	}
	tm := &models.TeamMember{
		TeamID: strings.TrimSpace(teamID),
		UserID: strings.TrimSpace(userID),
		Role:   role,
	}
	if tm.TeamID == "" || tm.UserID == "" {
		return nil, fmt.Errorf("team_id and user_id required")
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO team_members (team_id, user_id, role) VALUES (?,?,?)`,
		tm.TeamID, tm.UserID, tm.Role)
	if err != nil {
		return nil, err
	}
	return tm, nil
}

// CreateAPIKey stores only the key hash.
func (s *Store) CreateAPIKey(ctx context.Context, userID, keyHash string) (*models.APIKey, error) {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(keyHash) == "" {
		return nil, fmt.Errorf("user_id and key_hash required")
	}
	k := &models.APIKey{
		ID:        idgen.New(),
		UserID:    strings.TrimSpace(userID),
		CreatedAt: nowRFC3339(),
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO api_keys (id, user_id, key_hash, created_at) VALUES (?,?,?,?)`,
		k.ID, k.UserID, keyHash, k.CreatedAt)
	if err != nil {
		return nil, err
	}
	return k, nil
}

// Seed creates platform admin + default space when the default space is missing.
func Seed(s *Store, password string) error {
	if s == nil {
		return fmt.Errorf("store required")
	}
	ctx := context.Background()
	existing, err := s.GetSpace(ctx, DefaultSpaceID)
	if err != nil {
		return fmt.Errorf("seed lookup default space: %w", err)
	}
	if existing != nil {
		return nil
	}
	if strings.TrimSpace(password) == "" {
		password = "admin"
	}
	hash, err := passwd.Hash(password)
	if err != nil {
		return fmt.Errorf("seed hash: %w", err)
	}

	const adminEmail = "admin@kuayle.local"
	u, err := s.GetUserByEmail(ctx, adminEmail)
	if err != nil {
		return fmt.Errorf("seed lookup admin: %w", err)
	}
	if u == nil {
		u, err = s.CreateUser(ctx, adminEmail, hash, ActorUser, true)
		if err != nil {
			return fmt.Errorf("seed create admin: %w", err)
		}
	}
	sp, err := s.CreateSpace(ctx, DefaultSpaceID, "default")
	if err != nil {
		return fmt.Errorf("seed create space: %w", err)
	}
	if _, err := s.AddMember(ctx, sp.ID, u.ID, RoleOwner); err != nil {
		return fmt.Errorf("seed membership: %w", err)
	}
	return nil
}
