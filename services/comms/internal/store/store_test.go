package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/comms/models"
)

func TestChannelAndMessageThreads(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	ch, err := s.CreateChannel(ctx, "default", "general", "", "")
	if err != nil {
		t.Fatalf("create channel: %v", err)
	}
	if ch.Kind != models.KindSpace || ch.SpaceID != "default" {
		t.Fatalf("channel %+v", ch)
	}

	if _, err := s.CreateChannel(ctx, "default", "", "", ""); err == nil {
		t.Fatal("expected name required")
	}
	if _, err := s.CreateChannel(ctx, "default", "eng", "", models.KindTeam); err == nil {
		t.Fatal("expected team_id required")
	}
	teamCh, err := s.CreateChannel(ctx, "default", "eng", "team-1", "")
	if err != nil || teamCh.Kind != models.KindTeam {
		t.Fatalf("team channel: %v %#v", err, teamCh)
	}

	got, err := s.GetChannelInSpace(ctx, "default", ch.ID)
	if err != nil || got == nil || got.Name != "general" {
		t.Fatalf("get: %v %#v", err, got)
	}
	other, err := s.GetChannelInSpace(ctx, "other", ch.ID)
	if err != nil || other != nil {
		t.Fatalf("wrong space: %v %#v", err, other)
	}

	root, err := s.CreateMessage(ctx, ch.ID, "default", "u1", "hello", "", "", models.SourceUser, "")
	if err != nil {
		t.Fatalf("root: %v", err)
	}
	reply, err := s.CreateMessage(ctx, ch.ID, "default", "u2", "thread", root.ID, "", models.SourceUser, "")
	if err != nil || reply.ParentID != root.ID {
		t.Fatalf("reply: %v %#v", err, reply)
	}
	if _, err := s.CreateMessage(ctx, ch.ID, "default", "u1", "bad", "missing", "", models.SourceUser, ""); err == nil {
		t.Fatal("expected parent not found")
	}
	if _, err := s.CreateMessage(ctx, ch.ID, "default", "u1", "", "", "", models.SourceUser, ""); err == nil {
		t.Fatal("expected body or blob")
	}
	att, err := s.CreateMessage(ctx, ch.ID, "default", "u1", "", "", "blob-1", models.SourceUser, "")
	if err != nil || att.BlobID != "blob-1" {
		t.Fatalf("attachment: %v %#v", err, att)
	}

	all, err := s.ListMessages(ctx, MessageListFilter{ChannelID: ch.ID, Limit: 50})
	if err != nil || len(all) != 3 {
		t.Fatalf("list: %v %#v", err, all)
	}
	thread, err := s.ListMessages(ctx, MessageListFilter{ChannelID: ch.ID, ParentID: root.ID, Limit: 50})
	if err != nil || len(thread) != 1 || thread[0].ID != reply.ID {
		t.Fatalf("thread: %v %#v", err, thread)
	}

	first, err := s.CreateMessage(ctx, ch.ID, "default", "tg", "from telegram", "", "", models.SourceInteg, "tg:1")
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	dup, err := s.CreateMessage(ctx, ch.ID, "default", "tg", "from telegram again", "", "", models.SourceInteg, "tg:1")
	if err != nil || dup.ID != first.ID {
		t.Fatalf("idempotent ingest: %v %#v vs %#v", err, dup, first)
	}

	list, err := s.ListChannels(ctx, "default", "")
	if err != nil || len(list) != 2 {
		t.Fatalf("list channels: %v %#v", err, list)
	}
	onlyTeam, err := s.ListChannels(ctx, "default", "team-1")
	if err != nil || len(onlyTeam) != 1 || onlyTeam[0].ID != teamCh.ID {
		t.Fatalf("filter team: %v %#v", err, onlyTeam)
	}
}
