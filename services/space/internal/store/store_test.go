package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/pkg/passwd"
)

func TestStoreUserSpaceMember(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenStore(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer s.Close()
	ctx := context.Background()

	hash, err := passwd.Hash("secret")
	if err != nil {
		t.Fatal(err)
	}
	u, err := s.CreateUser(ctx, "Ada@Kuayle.local", hash, ActorUser, false)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	got, err := s.GetUserByEmail(ctx, "ada@kuayle.local")
	if err != nil || got == nil {
		t.Fatalf("get by email: %v %#v", err, got)
	}
	if got.ID != u.ID || got.Email != "ada@kuayle.local" || got.PasswordHash == "" {
		t.Fatalf("unexpected user %+v", got)
	}
	byID, err := s.GetUser(ctx, u.ID)
	if err != nil || byID == nil || byID.ID != u.ID {
		t.Fatalf("get user: %v %#v", err, byID)
	}

	sp, err := s.CreateSpace(ctx, "", "alpha")
	if err != nil {
		t.Fatalf("create space: %v", err)
	}
	if _, err := s.AddMember(ctx, sp.ID, u.ID, RoleOwner); err != nil {
		t.Fatalf("add member: %v", err)
	}
	m, err := s.GetMember(ctx, sp.ID, u.ID)
	if err != nil || m == nil || m.Role != RoleOwner {
		t.Fatalf("get member: %v %#v", err, m)
	}
	list, err := s.ListSpacesForUser(ctx, u.ID)
	if err != nil || len(list) != 1 || list[0].ID != sp.ID || list[0].Role != RoleOwner {
		t.Fatalf("list for user: %v %#v", err, list)
	}
	members, err := s.ListMembers(ctx, sp.ID)
	if err != nil || len(members) != 1 {
		t.Fatalf("list members: %v %#v", err, members)
	}

	team, err := s.CreateTeam(ctx, sp.ID, "", "core")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}
	if _, err := s.AddTeamMember(ctx, team.ID, u.ID, RoleMember); err != nil {
		t.Fatalf("add team member: %v", err)
	}
	if _, err := s.CreateAPIKey(ctx, u.ID, "hash-only"); err != nil {
		t.Fatalf("api key: %v", err)
	}
}

func TestSeedIdempotent(t *testing.T) {
	dir := t.TempDir()
	s, err := OpenStore(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := Seed(s, "admin"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := Seed(s, "admin"); err != nil {
		t.Fatalf("seed again: %v", err)
	}
	ctx := context.Background()
	sp, err := s.GetSpace(ctx, DefaultSpaceID)
	if err != nil || sp == nil || sp.Name != "default" {
		t.Fatalf("default space: %v %#v", err, sp)
	}
	u, err := s.GetUserByEmail(ctx, "admin@kuayle.local")
	if err != nil || u == nil || !u.PlatformAdmin || u.Actor != ActorUser {
		t.Fatalf("admin: %v %#v", err, u)
	}
	if !passwd.Verify("admin", u.PasswordHash) {
		t.Fatal("seed hash does not verify admin")
	}
	m, err := s.GetMember(ctx, DefaultSpaceID, u.ID)
	if err != nil || m == nil || m.Role != RoleOwner {
		t.Fatalf("owner: %v %#v", err, m)
	}
}
