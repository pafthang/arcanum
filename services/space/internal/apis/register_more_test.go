package apis

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/pkg/passwd"
	"github.com/pafthang/arcanum/services/space/internal/config"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/models"
)

func TestRegisterAndSwitch(t *testing.T) {
	s, err := store.OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	d := &Deps{Store: s, Config: config.Defaults()}
	out, code, err := registerUser(context.Background(), d, models.RegisterRequest{
		Email: "a@b.local", Password: "pass", Name: "alpha",
	})
	if err != nil || code != 201 || out.Token == "" || out.Space == nil {
		t.Fatalf("register: code=%d err=%v out=%#v", code, err, out)
	}
	_, code, err = registerUser(context.Background(), d, models.RegisterRequest{
		Email: "a@b.local", Password: "pass",
	})
	if err == nil || code != 409 {
		t.Fatalf("expected 409, got %d %v", code, err)
	}
	hash, _ := passwd.Hash("x")
	u2, err := s.CreateUser(context.Background(), "c@d.local", hash, store.ActorUser, false)
	if err != nil {
		t.Fatal(err)
	}
	sp2, err := s.CreateSpace(context.Background(), "", "other")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddMember(context.Background(), sp2.ID, out.User.ID, store.RoleMember); err != nil {
		t.Fatal(err)
	}
	sw, code, err := switchSpace(context.Background(), d, out.User.ID, sp2.ID)
	if err != nil || code != 200 || sw.SpaceRole != store.RoleMember || sw.Space.ID != sp2.ID {
		t.Fatalf("switch: code=%d err=%v %#v", code, err, sw)
	}
	_, code, err = switchSpace(context.Background(), d, u2.ID, out.Space.ID)
	if err == nil || code != 403 {
		t.Fatalf("expected 403, got %d %v", code, err)
	}
}
