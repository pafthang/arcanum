package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/agents/models"
)

func TestRunSessionMemorySkill(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	r, err := s.CreateRun(ctx, "default", "agent-1", "issue-1", "hello")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.Status != models.StatusQueued {
		t.Fatalf("run %+v", r)
	}
	got, err := s.GetRunInSpace(ctx, "default", r.ID)
	if err != nil || got == nil || got.Input != "hello" {
		t.Fatalf("get: %v %#v", err, got)
	}
	other, err := s.GetRunInSpace(ctx, "other", r.ID)
	if err != nil || other != nil {
		t.Fatalf("wrong space: %v %#v", err, other)
	}

	running, err := s.MarkRunning(ctx, r.ID)
	if err != nil || running.Status != models.StatusRunning || running.StartedAt == "" {
		t.Fatalf("running: %v %#v", err, running)
	}
	done, err := s.FinishRun(ctx, r.ID, models.StatusSucceeded, "out", "")
	if err != nil || done.Status != models.StatusSucceeded || done.Output != "out" {
		t.Fatalf("finish: %v %#v", err, done)
	}

	if _, err := s.CreateRun(ctx, "default", "", "", ""); err == nil {
		t.Fatal("expected agent required")
	}

	r2, err := s.CreateRun(ctx, "default", "agent-1", "", "x")
	if err != nil {
		t.Fatal(err)
	}
	canc, err := s.CancelRun(ctx, r2.ID)
	if err != nil || canc.Status != models.StatusCancelling {
		t.Fatalf("cancel: %v %#v", err, canc)
	}

	sess, err := s.UpsertSession(ctx, r.ID, "default", "think", `{"ok":true}`)
	if err != nil || sess.Stage != "think" {
		t.Fatalf("session: %v %#v", err, sess)
	}
	sess2, err := s.UpsertSession(ctx, r.ID, "default", "act", `{"n":1}`)
	if err != nil || sess2.ID != sess.ID || sess2.Stage != "act" {
		t.Fatalf("upsert: %v %#v", err, sess2)
	}

	mem, err := s.PutMemory(ctx, "default", "agent-1", models.TierWorking, "k", "v1")
	if err != nil || mem.Value != "v1" {
		t.Fatalf("put: %v %#v", err, mem)
	}
	mem2, err := s.PutMemory(ctx, "default", "agent-1", models.TierWorking, "k", "v2")
	if err != nil || mem2.ID != mem.ID || mem2.Value != "v2" {
		t.Fatalf("put2: %v %#v", err, mem2)
	}
	mems, err := s.ListMemories(ctx, "default", "agent-1")
	if err != nil || len(mems) != 1 {
		t.Fatalf("list mem: %v %#v", err, mems)
	}

	sk, err := s.CreateSkill(ctx, "default", "summarize", "body")
	if err != nil {
		t.Fatalf("skill: %v", err)
	}
	skills, err := s.ListSkills(ctx, "default")
	if err != nil || len(skills) != 1 || skills[0].ID != sk.ID {
		t.Fatalf("list skills: %v %#v", err, skills)
	}

	list, err := s.ListRuns(ctx, "default")
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v %#v", err, list)
	}
}
