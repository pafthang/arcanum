package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/work/models"
)

func TestIssueCRUDAndComments(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	iss, err := s.CreateIssue(ctx, "default", "First", "body", "", "u1")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if iss.Status != models.StatusOpen || iss.SpaceID != "default" || iss.AssigneeID != "u1" {
		t.Fatalf("issue %+v", iss)
	}

	got, err := s.GetIssueInSpace(ctx, "default", iss.ID)
	if err != nil || got == nil || got.Title != "First" {
		t.Fatalf("get: %v %#v", err, got)
	}
	other, err := s.GetIssueInSpace(ctx, "other", iss.ID)
	if err != nil || other != nil {
		t.Fatalf("wrong space: %v %#v", err, other)
	}

	started := models.StatusStarted
	title := "Renamed"
	upd, err := s.UpdateIssue(ctx, iss.ID, &title, nil, &started, nil)
	if err != nil || upd.Title != "Renamed" || upd.Status != models.StatusStarted {
		t.Fatalf("update: %v %#v", err, upd)
	}

	if _, err := s.CreateIssue(ctx, "default", "", "", "", ""); err == nil {
		t.Fatal("expected title required")
	}
	if _, err := s.CreateIssue(ctx, "default", "x", "", "nope", ""); err == nil {
		t.Fatal("expected invalid status")
	}

	c, err := s.AddComment(ctx, iss.ID, "u1", "hello", "")
	if err != nil || c.Body != "hello" {
		t.Fatalf("comment: %v %#v", err, c)
	}
	comments, err := s.ListComments(ctx, iss.ID)
	if err != nil || len(comments) != 1 {
		t.Fatalf("list comments: %v %#v", err, comments)
	}

	list, err := s.ListIssues(ctx, "default")
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %#v", err, list)
	}
	empty, err := s.ListIssues(ctx, "missing")
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty list: %v %#v", err, empty)
	}

	ov, err := s.Overview(ctx, "default")
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if ov.Issues != 1 || ov.ByStatus[models.StatusStarted] != 1 || ov.Assigned != 1 || ov.Comments != 1 {
		t.Fatalf("overview %+v", ov)
	}
}
