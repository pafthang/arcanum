package store

import (
	"database/sql"
	"strings"
)

func migrateCommentBlob(db *sql.DB) error {
	_, err := db.Exec(`ALTER TABLE issue_comments ADD COLUMN blob_id TEXT NOT NULL DEFAULT ''`)
	if err != nil && !strings.Contains(err.Error(), "duplicate column") {
		return err
	}
	return nil
}
