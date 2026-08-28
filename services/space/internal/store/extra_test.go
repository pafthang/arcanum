package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/pafthang/arcanum/pkg/passwd"
)

func TestUpdateMemberAndTeamsList(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "data"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	ctx := context.Background()
	hash, _ := passwd.Hash("x")
	u, err := s.CreateUser(ctx, "m@t.local", hash, ActorUser, false)
	if err != nil {
		t.Fatal(err)
	}
	sp, err := s.CreateSpace(ctx, "", "s")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddMember(ctx, sp.ID, u.ID, RoleOwner); err != nil {
		t.Fatal(err)
	}
	if _, err := s.UpdateMember(ctx, sp.ID, u.ID, RoleMember); err == nil {
		t.Fatal("expected last owner demote to fail")
	}
	u2, err := s.CreateUser(ctx, "n@t.local", hash, ActorUser, false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.AddMember(ctx, sp.ID, u2.ID, RoleOwner); err != nil {
		t.Fatal(err)
	}
	m, err := s.UpdateMember(ctx, sp.ID, u.ID, RoleMember)
	if err != nil || m.Role != RoleMember {
		t.Fatalf("update: %v %#v", err, m)
	}
	team, err := s.CreateTeam(ctx, sp.ID, "", "core")
	if err != nil {
		t.Fatal(err)
	}
	list, err := s.ListTeams(ctx, sp.ID)
	if err != nil || len(list) != 1 || list[0].ID != team.ID {
		t.Fatalf("list teams: %v %#v", err, list)
	}
	renamed, err := s.UpdateTeam(ctx, sp.ID, team.ID, "core-2")
	if err != nil || renamed.Name != "core-2" {
		t.Fatalf("rename team: %v %#v", err, renamed)
	}
	if err := s.RemoveMember(ctx, sp.ID, u.ID); err != nil {
		t.Fatalf("remove member: %v", err)
	}
	if err := s.RemoveMember(ctx, sp.ID, u2.ID); err == nil {
		t.Fatal("expected last owner remove to fail")
	}
}
