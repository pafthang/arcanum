package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/integ/models"
)

func TestConnectorWebhookInbound(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	c, err := s.CreateConnector(ctx, "default", models.KindHook, "tg", "", "s3cr3t", map[string]any{"chat": "1"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if c.Status != models.StatusPending || !c.HasSecret || c.Secret != "s3cr3t" {
		t.Fatalf("connector %+v", c)
	}
	got, err := s.GetConnectorInSpace(ctx, "default", c.ID)
	if err != nil || got == nil || got.Name != "tg" {
		t.Fatalf("get: %v %#v", err, got)
	}
	other, err := s.GetConnectorInSpace(ctx, "other", c.ID)
	if err != nil || other != nil {
		t.Fatalf("wrong space: %v %#v", err, other)
	}

	st := models.StatusActive
	upd, err := s.UpdateConnector(ctx, c.ID, nil, &st, nil, nil)
	if err != nil || upd.Status != models.StatusActive {
		t.Fatalf("update: %v %#v", err, upd)
	}

	if _, err := s.CreateConnector(ctx, "default", "nope", "x", "", "", nil); err == nil {
		t.Fatal("expected invalid kind")
	}

	gh, err := s.CreateConnector(ctx, "default", models.KindGitHub, "gh", models.StatusActive, "ghsec", nil)
	if err != nil {
		t.Fatalf("github: %v", err)
	}
	repo, err := s.CreateRepo(ctx, gh.ID, "default", "kuayle", "arcanum", "42")
	if err != nil || repo.Owner != "kuayle" {
		t.Fatalf("repo: %v %#v", err, repo)
	}
	repos, err := s.ListRepos(ctx, gh.ID)
	if err != nil || len(repos) != 1 {
		t.Fatalf("list repos: %v %#v", err, repos)
	}

	h, err := s.CreateWebhook(ctx, "default", "https://example.test/hook", "whsec", []string{"issue.created"}, true)
	if err != nil {
		t.Fatalf("webhook: %v", err)
	}
	matched, err := s.ListActiveWebhooksForEvent(ctx, "default", "issue.created")
	if err != nil || len(matched) != 1 || matched[0].ID != h.ID {
		t.Fatalf("match: %v %#v", err, matched)
	}
	miss, err := s.ListActiveWebhooksForEvent(ctx, "default", "issue.updated")
	if err != nil || len(miss) != 0 {
		t.Fatalf("miss: %v %#v", err, miss)
	}

	del, err := s.RecordDelivery(ctx, "default", "webhook", h.ID, "issue.created", "{}", models.DeliverySent, "", 1, nowRFC3339())
	if err != nil || del.Status != models.DeliverySent {
		t.Fatalf("delivery: %v %#v", err, del)
	}
	dels, err := s.ListDeliveries(ctx, "default")
	if err != nil || len(dels) != 1 {
		t.Fatalf("list deliveries: %v %#v", err, dels)
	}

	ev, err := s.RecordInbound(ctx, "default", c.ID, "hook", "ext-1", map[string]any{"text": "ABC-12"}, []string{"ABC-12"})
	if err != nil || len(ev.IssueKeys) != 1 {
		t.Fatalf("inbound: %v %#v", err, ev)
	}

	list, err := s.ListConnectors(ctx, "default")
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %v %#v", err, list)
	}
	empty, err := s.ListConnectors(ctx, "missing")
	if err != nil || len(empty) != 0 {
		t.Fatalf("empty: %v %#v", err, empty)
	}
}
