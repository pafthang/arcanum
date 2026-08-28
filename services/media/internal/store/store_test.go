package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestPutGetList(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	b, err := s.Put(ctx, "sp1", "note.txt", "text/plain", "u1", []byte("hello"))
	if err != nil || b == nil || b.Size != 5 || b.SHA256 == "" {
		t.Fatalf("put: %v %#v", err, b)
	}
	got, data, err := s.ReadBytes(ctx, "sp1", b.ID)
	if err != nil || got == nil || string(data) != "hello" {
		t.Fatalf("read: %v %#v %q", err, got, data)
	}
	miss, err := s.GetMeta(ctx, "other", b.ID)
	if err != nil || miss != nil {
		t.Fatalf("other space: %v %#v", err, miss)
	}
	list, err := s.List(ctx, "sp1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
}
