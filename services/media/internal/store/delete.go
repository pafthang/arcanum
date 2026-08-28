package store

import (
	"context"
	"os"
)

// Delete removes metadata and the file. Missing blob is not an error.
func (s *Store) Delete(ctx context.Context, spaceID, id string) error {
	meta, err := s.GetMeta(ctx, spaceID, id)
	if err != nil {
		return err
	}
	if meta == nil {
		return nil
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM blobs WHERE id = ? AND space_id = ?`, meta.ID, meta.SpaceID); err != nil {
		return err
	}
	_ = os.Remove(s.blobPath(meta.SpaceID, meta.ID))
	return nil
}
