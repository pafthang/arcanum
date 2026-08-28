package apis

import (
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/models"
)

func requireMember(req mini.Request, d *Deps, spaceID string) bool {
	tc := httpx.SpaceContext(req)
	if tc.UserID == "" && !tc.IsPlatform {
		httpx.Error(req, 401, "The request requires valid authorization token.", nil)
		return false
	}
	if tc.IsPlatform {
		return true
	}
	if tc.UserID == "" {
		httpx.Error(req, 401, "The request requires valid authorization token.", nil)
		return false
	}
	m, err := d.Store.GetMember(req.Context(), spaceID, tc.UserID)
	if err != nil {
		httpx.Error(req, 500, err.Error(), nil)
		return false
	}
	if m == nil {
		httpx.Error(req, 403, "Not a member of this space.", nil)
		return false
	}
	return true
}

func memberRole(req mini.Request, d *Deps, spaceID string) (string, bool) {
	tc := httpx.SpaceContext(req)
	if tc.IsPlatform {
		return store.RoleOwner, true
	}
	if !requireMember(req, d, spaceID) {
		return "", false
	}
	m, err := d.Store.GetMember(req.Context(), spaceID, tc.UserID)
	if err != nil || m == nil {
		httpx.Error(req, 403, "Not a member of this space.", nil)
		return "", false
	}
	return m.Role, true
}

func requirePerm(req mini.Request, d *Deps, spaceID, perm string) bool {
	role, ok := memberRole(req, d, spaceID)
	if !ok {
		return false
	}
	if !models.RoleHas(role, perm) {
		httpx.Error(req, 403, "Insufficient space role.", nil)
		return false
	}
	return true
}
