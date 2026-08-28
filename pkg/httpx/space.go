package httpx

import (
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/pkg/mini"
)

// Standard space roles (higher rank = more power). Mirrors auth service.
const (
	RoleViewer = "viewer"
	RoleMember = "member"
	RoleAdmin  = "admin"
	RoleOwner  = "owner"
)

// RequireSpaceRead ensures an authenticated caller with space viewer+ (or platform admin).
// Returns the effective space filter: empty means all spaces (platform admin + allSpaces/no space).
// Writes 401/403 on failure.
func RequireSpaceRead(req mini.Request) (spaceID string, ok bool) {
	tc := SpaceContext(req)
	if !tc.IsPlatform && tc.UserID == "" {
		Error(req, 401, "The request requires valid authorization token.", nil)
		return "", false
	}
	if !tc.IsPlatform && !HasMinSpaceRole(tc.SpaceRole, RoleViewer) {
		Error(req, 403, "Space viewer role required.", nil)
		return "", false
	}
	spaceID = EffectiveSpaceID(req)
	if spaceID == "" && !tc.IsPlatform {
		Error(req, 403, "Active space required. Use POST /api/auth/switch-space.", nil)
		return "", false
	}
	if spaceID == "" {
		spaceID = queryWorkspaceID(req)
	}
	return spaceID, true
}

// RequireSpaceWrite ensures space member+ (or platform admin).
// Platform admin may return an empty spaceID (caller reads space from body/query on create).
// Non-admin always returns the JWT active space. Writes 401/403 on failure.
func RequireSpaceWrite(req mini.Request) (spaceID string, ok bool) {
	tc := SpaceContext(req)
	if !tc.IsPlatform && tc.UserID == "" {
		Error(req, 401, "The request requires valid authorization token.", nil)
		return "", false
	}
	if tc.IsPlatform {
		spaceID = EffectiveSpaceID(req)
		if spaceID == "" {
			spaceID = queryWorkspaceID(req)
		}
		return spaceID, true
	}
	if !HasMinSpaceRole(tc.SpaceRole, RoleMember) {
		Error(req, 403, "Space member role required.", nil)
		return "", false
	}
	if tc.SpaceID == "" {
		Error(req, 403, "Active space required. Use POST /api/auth/switch-space.", nil)
		return "", false
	}
	return tc.SpaceID, true
}

// RequireSpace is a general gate for minRole (viewer|member|admin|owner).
func RequireSpace(req mini.Request, minRole string) (spaceID string, ok bool) {
	minRole = strings.ToLower(strings.TrimSpace(minRole))
	if minRole == "" || minRole == RoleViewer {
		return RequireSpaceRead(req)
	}
	tc := SpaceContext(req)
	if !tc.IsPlatform && tc.UserID == "" {
		Error(req, 401, "The request requires valid authorization token.", nil)
		return "", false
	}
	if tc.IsPlatform {
		spaceID = EffectiveSpaceID(req)
		if spaceID == "" {
			spaceID = queryWorkspaceID(req)
		}
		return spaceID, true
	}
	if !HasMinSpaceRole(tc.SpaceRole, minRole) {
		Error(req, 403, fmt.Sprintf("Space %s role required.", minRole), nil)
		return "", false
	}
	if tc.SpaceID == "" {
		Error(req, 403, "Active space required. Use POST /api/auth/switch-space.", nil)
		return "", false
	}
	return tc.SpaceID, true
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// queryWorkspaceID returns ?spaceId|space (platform override helpers).
func queryWorkspaceID(req mini.Request) string {
	return firstNonEmpty(
		Query(req, "spaceId"),
		Query(req, "space"),
	)
}

// PathSpaceID returns workspace id from path params {spaceId} or {space}.
func PathSpaceID(req mini.Request) string {
	return firstNonEmpty(
		mini.PathParam(req, "spaceId"),
		mini.PathParam(req, "space"),
	)
}

// RequireSpacePath resolves scope from path {spaceId}/{space} and validates JWT.
// Non-platform callers: JWT active space must equal the path space (switch-space first).
// Platform admin may address any space via the path (path required).
// write=true requires member+; write=false requires viewer+.
func RequireSpacePath(req mini.Request, write bool) (spaceID string, ok bool) {
	pathSpace := PathSpaceID(req)
	if pathSpace == "" {
		pathSpace = AuthSpaceID(req)
	}
	if pathSpace == "" {
		Error(req, 400, "spaceId path required.", nil)
		return "", false
	}
	var jwtSpace string
	if write {
		jwtSpace, ok = RequireSpaceWrite(req)
	} else {
		jwtSpace, ok = RequireSpaceRead(req)
	}
	if !ok {
		return "", false
	}
	tc := SpaceContext(req)
	if tc.IsPlatform {
		return pathSpace, true
	}
	if jwtSpace != "" && jwtSpace != pathSpace {
		Error(req, 403, "Path space does not match active workspace. Use POST /api/auth/switch-space.", nil)
		return "", false
	}
	return pathSpace, true
}
