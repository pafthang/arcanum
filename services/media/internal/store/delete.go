package store

import "context"

// Delete removes metadata and the object. Missing blob is not an error.
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
	if s.Objects != nil {
		_ = s.Objects.Delete(ctx, s.objectKey(meta.SpaceID, meta.ID))
	}
	return nil
}
