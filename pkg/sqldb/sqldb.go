// Package sqldb opens SQLite databases with sensible defaults for services.
package sqldb

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

// Open opens (or creates) a SQLite database under dataDir/name.db.
// Applies WAL and busy timeout pragmas.
func Open(dataDir, name string) (*sql.DB, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir data dir: %w", err)
	}
	name = strings.TrimSuffix(name, ".db")
	path := filepath.Join(dataDir, name+".db")

	// modernc driver: path with query pragmas
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite single-writer friendly default per service
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping %s: %w", path, err)
	}
	return db, nil
}

// Migrate runs sequential SQL statements (each must be complete).
func Migrate(db *sql.DB, statements ...string) error {
	for i, stmt := range statements {
		if strings.TrimSpace(stmt) == "" {
			continue
		}
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration step %d: %w\nSQL: %s", i, err, stmt)
		}
	}
	return nil
}

// IsConstraintError checks if err is a UNIQUE constraint violation.
func IsConstraintError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "UNIQUE constraint failed")
}
