package apis

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/services/space/internal/config"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/models"
)

func TestLoginSeedAdmin(t *testing.T) {
	dir := t.TempDir()
	s, err := store.OpenStore(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := store.Seed(s, "admin"); err != nil {
		t.Fatal(err)
	}
	d := &Deps{Store: s, Config: config.Defaults()}
	out, code, err := login(context.Background(), d, models.LoginRequest{
		Email: "admin@kuayle.local", Password: "admin",
	})
	if err != nil || code != 200 {
		t.Fatalf("login: code=%d err=%v", code, err)
	}
	if out.Token == "" || out.User.Email != "admin@kuayle.local" || !out.User.PlatformAdmin {
		t.Fatalf("unexpected login %#v", out)
	}
	if out.Space == nil || out.Space.ID != store.DefaultSpaceID || out.SpaceRole != store.RoleOwner {
		t.Fatalf("space claim %#v", out)
	}
	_, code, err = login(context.Background(), d, models.LoginRequest{
		Email: "admin@kuayle.local", Password: "wrong",
	})
	if err == nil || code != 401 {
		t.Fatalf("expected 401, got code=%d err=%v", code, err)
	}
}
