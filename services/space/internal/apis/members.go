package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/models"
)

func registerMembers(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("members_invite", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermMemberInvite) {
			return
		}
		var in models.InviteMemberRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		email := strings.ToLower(strings.TrimSpace(in.Email))
		role := strings.TrimSpace(in.Role)
		if role == "" {
			role = store.RoleMember
		}
		if email == "" {
			httpx.Error(req, 400, "email required.", nil)
			return
		}
		u, err := d.Store.GetUserByEmail(req.Context(), email)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if u == nil {
			httpx.Error(req, 404, "User not found. Register first.", nil)
			return
		}
		if existing, err := d.Store.GetMember(req.Context(), spaceID, u.ID); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		} else if existing != nil {
			httpx.JSON(req, 200, existing)
			return
		}
		m, err := d.Store.AddMember(req.Context(), spaceID, u.ID, role)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, m)
	}), mini.Public("POST", "/api/spaces/{spaceId}/members", "space", "members.invite")))

	must(svc.AddEndpoint("members_update", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		userID := strings.TrimSpace(mini.PathParam(req, "userId"))
		if spaceID == "" || userID == "" {
			httpx.Error(req, 400, "spaceId and userId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermMemberRole) {
			return
		}
		var in models.UpdateMemberRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		m, err := d.Store.UpdateMember(req.Context(), spaceID, userID, in.Role)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if m == nil {
			httpx.Error(req, 404, "Member not found.", nil)
			return
		}
		httpx.JSON(req, 200, m)
	}), mini.Public("PATCH", "/api/spaces/{spaceId}/members/{userId}", "space", "members.update")))
}
