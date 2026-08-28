// Package httpx helpers map mini request/response to JSON API conventions.
package httpx

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/pafthang/arcanum/pkg/mini"
)

// APIError is a structured error body for gate clients.
type APIError struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data,omitempty"`
}

// JSON responds with status and payload.
func JSON(req mini.Request, status int, v any) {
	_ = req.RespondJSON(v, mini.WithStatus(status))
}

// NoContent responds with empty body and status.
func NoContent(req mini.Request, status int) {
	_ = req.Respond(nil, mini.WithStatus(status))
}

// Error sends a JSON API error.
func Error(req mini.Request, status int, message string, data map[string]any) {
	if message == "" {
		message = httpStatusText(status)
	}
	body, _ := json.Marshal(APIError{Code: status, Message: message, Data: data})
	code := fmt.Sprintf("%d", status)
	_ = req.Error(code, message, body, mini.WithStatus(status))
}

// BindJSON unmarshals request body into dest.
func BindJSON(req mini.Request, dest any) error {
	data := req.Data()
	if len(data) == 0 {
		return fmt.Errorf("empty body")
	}
	return json.Unmarshal(data, dest)
}

// QueryValues parses gate-forwarded query string (X-Query-String).
func QueryValues(req mini.Request) url.Values {
	if v := req.Headers().Get("X-Query-String"); v != "" {
		q, err := url.ParseQuery(v)
		if err == nil {
			return q
		}
	}
	return url.Values{}
}

// Query returns a single query parameter.
// Order: X-Mini-Query-*, X-Query-String, form headers.
func Query(req mini.Request, name string) string {
	if v := req.Headers().Get("X-Mini-Query-" + name); v != "" {
		return v
	}
	if v := QueryValues(req).Get(name); v != "" {
		return v
	}
	return ""
}

// QueryInt returns query param as int or def.
func QueryInt(req mini.Request, name string, def int) int {
	v := Query(req, name)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

// PageParams returns page (1-based) and perPage from query string.
func PageParams(req mini.Request) (page, perPage int) {
	page = QueryInt(req, "page", 1)
	perPage = QueryInt(req, "perPage", 30)
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 30
	}
	if perPage > 500 {
		perPage = 500
	}
	return page, perPage
}

// AuthSubject returns gate-injected subject (JWT sub).
func AuthSubject(req mini.Request) string {
	return req.Headers().Get("X-Mini-Subject")
}

// AuthSpaceID returns active workspace scope from JWT claim headers (space_id).
// AuthSpaceID returns active workspace scope from JWT claim headers (space_id).
func AuthSpaceID(req mini.Request) string {
	return req.Headers().Get("X-Mini-Space-Id")
}

// AuthSpaceRole returns workspace role claim (owner|admin|member|viewer).
func AuthSpaceRole(req mini.Request) string {
	return strings.ToLower(req.Headers().Get("X-Mini-Space-Role"))
}

// AuthPlatformRole returns platform_role claim ("" | platform_admin).
func AuthPlatformRole(req mini.Request) string {
	if v := req.Headers().Get("X-Mini-Platform-Role"); v != "" {
		return strings.ToLower(v)
	}
	return strings.ToLower(req.Headers().Get("X-Mini-Role"))
}

// AuthType returns admin|user|"" from claim headers.
func AuthType(req mini.Request) string {
	t := strings.ToLower(req.Headers().Get("X-Mini-Auth-Type"))
	if t != "" {
		return t
	}
	// fallback: some gateways put typ in X-Mini-Typ
	return strings.ToLower(req.Headers().Get("X-Mini-Typ"))
}

// Header returns first header value.
func Header(req mini.Request, key string) string {
	return req.Headers().Get(key)
}

// IsPlatformAdmin reports platform-level superuser (not team role "admin").
// JWT issues typ=admin and role/platform_role=platform_admin for these users.
func IsPlatformAdmin(req mini.Request) bool {
	if AuthType(req) == "admin" {
		return true
	}
	role := AuthPlatformRole(req)
	return role == "platform_admin"
}

// IsAdmin is an alias of IsPlatformAdmin.
// Team role "admin" must NOT pass this check.
func IsAdmin(req mini.Request) bool {
	return IsPlatformAdmin(req)
}

// IsAuthenticated reports any JWT subject present.
func IsAuthenticated(req mini.Request) bool {
	return AuthSubject(req) != ""
}

// RequireAdmin returns false and writes 401 if not platform admin.
// When no auth headers at all (dev without gate JWT), allow if allowDevOpen.
func RequireAdmin(req mini.Request, allowDevOpen bool) bool {
	if IsPlatformAdmin(req) {
		return true
	}
	if allowDevOpen && !IsAuthenticated(req) && AuthType(req) == "" {
		return true
	}
	Error(req, 401, "The request requires valid admin authorization token.", nil)
	return false
}

// Space role ranks (higher = more power). Mirrors auth service.
var teamRoleRank = map[string]int{
	"viewer": 1,
	"member": 2,
	"admin":  3,
	"owner":  4,
}

// HasMinSpaceRole reports whether have >= need.
func HasMinSpaceRole(have, need string) bool {
	return teamRoleRank[strings.ToLower(have)] >= teamRoleRank[strings.ToLower(need)]
}

// AuthSpaceContext is the active multi-tenant scope from JWT headers.
type AuthSpaceContext struct {
	UserID       string
	SpaceID      string
	SpaceRole    string
	PlatformRole string
	IsPlatform   bool
	Email        string
}

// SpaceContext extracts auth + workspace scope from gate claim headers.
func SpaceContext(req mini.Request) AuthSpaceContext {
	return AuthSpaceContext{
		UserID:       AuthSubject(req),
		SpaceID:      AuthSpaceID(req),
		SpaceRole:    AuthSpaceRole(req),
		PlatformRole: AuthPlatformRole(req),
		IsPlatform:   IsPlatformAdmin(req),
		Email:        Header(req, "X-Mini-Email"),
	}
}

// EffectiveSpaceID returns the workspace (space) to scope data to.
// Platform admin:
//   - ?allSpaces=1 → empty (see all workspaces)
//   - ?spaceId=… → override active workspace
//   - else JWT space_id (may be empty = all)
//
// Non-admin: JWT workspace id (empty → caller should 403 on space-scoped resources).
func EffectiveSpaceID(req mini.Request) string {
	tc := SpaceContext(req)
	if tc.IsPlatform {
		if v := Query(req, "allSpaces"); v == "1" || v == "true" {
			return ""
		}
		if q := Query(req, "spaceId"); q != "" {
			return q
		}
	}
	return tc.SpaceID
}

func httpStatusText(code int) string {
	switch code {
	case 400:
		return "Something went wrong while processing your request."
	case 401:
		return "The request requires valid authentication credentials."
	case 403:
		return "You are not allowed to perform this request."
	case 404:
		return "The requested resource wasn't found."
	default:
		return "Something went wrong while processing your request."
	}
}
