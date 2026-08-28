package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/runtime/models"
)

func TestMachineCRUD(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	m, err := s.Create(ctx, "sp", "dev-1", "debian:bookworm-slim", "ag1", models.StatusRecorded)
	if err != nil || m.Name != "dev-1" {
		t.Fatalf("create: %v %#v", err, m)
	}
	got, err := s.GetInSpace(ctx, "sp", m.ID)
	if err != nil || got == nil {
		t.Fatalf("get: %v %#v", err, got)
	}
	miss, err := s.GetInSpace(ctx, "other", m.ID)
	if err != nil || miss != nil {
		t.Fatalf("space: %v %#v", err, miss)
	}
	stopped, err := s.SetStatus(ctx, "sp", m.ID, models.StatusStopped, "", "")
	if err != nil || stopped.Status != models.StatusStopped {
		t.Fatalf("stop: %v %#v", err, stopped)
	}
	list, err := s.List(ctx, "sp")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
}
