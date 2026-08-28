package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/space/models"
)

func registerTeams(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("teams_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListTeams(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/teams", "space", "teams.list")))

	must(svc.AddEndpoint("teams_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermTeamManage) {
			return
		}
		var in models.CreateTeamRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		t, err := d.Store.CreateTeam(req.Context(), spaceID, in.ParentID, in.Name)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, t)
	}), mini.Public("POST", "/api/spaces/{spaceId}/teams", "space", "teams.create")))

	must(svc.AddEndpoint("teams_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		teamID := strings.TrimSpace(mini.PathParam(req, "teamId"))
		if spaceID == "" || teamID == "" {
			httpx.Error(req, 400, "spaceId and teamId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		t, err := d.Store.GetTeam(req.Context(), spaceID, teamID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if t == nil {
			httpx.Error(req, 404, "Team not found.", nil)
			return
		}
		members, err := d.Store.ListTeamMembers(req.Context(), t.ID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"team": t, "members": members})
	}), mini.Public("GET", "/api/spaces/{spaceId}/teams/{teamId}", "space", "teams.get")))

	must(svc.AddEndpoint("team_members_add", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		teamID := strings.TrimSpace(mini.PathParam(req, "teamId"))
		if spaceID == "" || teamID == "" {
			httpx.Error(req, 400, "spaceId and teamId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermTeamManage) {
			return
		}
		t, err := d.Store.GetTeam(req.Context(), spaceID, teamID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if t == nil {
			httpx.Error(req, 404, "Team not found.", nil)
			return
		}
		var in models.AddTeamMemberRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		m, err := d.Store.GetMember(req.Context(), spaceID, in.UserID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if m == nil {
			httpx.Error(req, 400, "user is not a member of this space.", nil)
			return
		}
		tm, err := d.Store.AddTeamMember(req.Context(), teamID, in.UserID, in.Role)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, tm)
	}), mini.Public("POST", "/api/spaces/{spaceId}/teams/{teamId}/members", "space", "teams.members.add")))
}
