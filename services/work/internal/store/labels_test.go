package store

import (
	"context"
	"path/filepath"
	"testing"
)

func TestLabelsOnIssue(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	iss, err := s.CreateIssue(ctx, "default", "Tagged", "", "", "")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	bug, err := s.CreateLabel(ctx, "default", "bug", "#ff0000")
	if err != nil {
		t.Fatalf("label: %v", err)
	}
	if _, err := s.CreateLabel(ctx, "default", "bug", ""); err == nil {
		t.Fatal("expected unique name")
	}
	if _, err := s.CreateLabel(ctx, "other", "bug", ""); err != nil {
		t.Fatalf("other space: %v", err)
	}

	if err := s.SetIssueLabels(ctx, "default", iss.ID, []string{bug.ID}); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.SetIssueLabels(ctx, "default", iss.ID, []string{"missing"}); err == nil {
		t.Fatal("expected foreign label reject")
	}

	by, err := s.LabelsForIssues(ctx, []string{iss.ID})
	if err != nil || len(by[iss.ID]) != 1 || by[iss.ID][0].Name != "bug" {
		t.Fatalf("for issues: %v %#v", err, by)
	}

	list, err := s.ListLabels(ctx, "default")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}

	if err := s.SetIssueFields(ctx, iss.ID, "high", "2026-09-01", ""); err != nil {
		t.Fatalf("fields: %v", err)
	}
	got, err := s.GetIssue(ctx, iss.ID)
	if err != nil || got.Priority != "high" || got.DueAt != "2026-09-01" || len(got.Labels) != 1 {
		t.Fatalf("hydrate: %v %#v", err, got)
	}

	if err := s.SetIssueLabels(ctx, "default", iss.ID, nil); err != nil {
		t.Fatalf("clear: %v", err)
	}
	by, err = s.LabelsForIssues(ctx, []string{iss.ID})
	if err != nil || len(by[iss.ID]) != 0 {
		t.Fatalf("cleared: %v %#v", err, by)
	}
}
