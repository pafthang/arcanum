package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/agents/internal/store"
	"github.com/pafthang/arcanum/services/agents/models"
)

func TestExecuteSucceeds(t *testing.T) {
	s, err := store.OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	run, err := s.CreateRun(ctx, "default", "agent-1", "iss", "ping")
	if err != nil {
		t.Fatal(err)
	}
	r := &Runner{Store: s}
	out, err := r.Execute(ctx, run)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != models.StatusSucceeded || out.Output == "" {
		t.Fatalf("run %+v", out)
	}
	sess, err := s.GetSession(ctx, run.ID)
	if err != nil || sess == nil || sess.Stage != StageSummarize {
		t.Fatalf("session: %v %#v", err, sess)
	}
}

func TestExecuteCancel(t *testing.T) {
	s, err := store.OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	run, err := s.CreateRun(ctx, "default", "agent-1", "", "x")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.CancelRun(ctx, run.ID); err != nil {
		t.Fatal(err)
	}
	out, err := (&Runner{Store: s}).Execute(ctx, run)
	if err != nil {
		t.Fatal(err)
	}
	if out.Status != models.StatusCancelled {
		t.Fatalf("status %s", out.Status)
	}
}
