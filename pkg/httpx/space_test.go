package httpx

import (
	"context"
	"testing"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
)

type stubReq struct {
	h       nats.Header
	errCode string
}

func (r *stubReq) Respond([]byte, ...mini.RespondOpt) error  { return nil }
func (r *stubReq) RespondJSON(any, ...mini.RespondOpt) error { return nil }
func (r *stubReq) Error(code, _ string, _ []byte, _ ...mini.RespondOpt) error {
	r.errCode = code
	return nil
}
func (r *stubReq) Data() []byte             { return nil }
func (r *stubReq) Headers() mini.Headers    { return mini.Headers(r.h) }
func (r *stubReq) Subject() string          { return "test" }
func (r *stubReq) Reply() string            { return "" }
func (r *stubReq) Context() context.Context { return context.Background() }
func (r *stubReq) PathParam(string) string  { return "" }

func reqWith(headers map[string]string) *stubReq {
	h := nats.Header{}
	for k, v := range headers {
		h.Set(k, v)
	}
	return &stubReq{h: h}
}

func TestRequireSpaceRead_member(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "member",
	})
	spaceID, ok := RequireSpaceRead(r)
	if !ok || spaceID != "space-a" {
		t.Fatalf("got spaceID=%q ok=%v err=%s", spaceID, ok, r.errCode)
	}
}

func TestRequireSpaceRead_unauthenticated(t *testing.T) {
	r := reqWith(nil)
	_, ok := RequireSpaceRead(r)
	if ok || r.errCode != "401" {
		t.Fatalf("want 401, ok=%v code=%s", ok, r.errCode)
	}
}

func TestRequireSpaceRead_noActiveSpace(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Role": "viewer",
	})
	_, ok := RequireSpaceRead(r)
	if ok || r.errCode != "403" {
		t.Fatalf("want 403, ok=%v code=%s", ok, r.errCode)
	}
}

func TestRequireSpaceWrite_viewerForbidden(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "viewer",
	})
	_, ok := RequireSpaceWrite(r)
	if ok || r.errCode != "403" {
		t.Fatalf("want 403, ok=%v code=%s", ok, r.errCode)
	}
}

func TestRequireSpaceWrite_member(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "member",
	})
	spaceID, ok := RequireSpaceWrite(r)
	if !ok || spaceID != "space-a" {
		t.Fatalf("got spaceID=%q ok=%v err=%s", spaceID, ok, r.errCode)
	}
}

func TestRequireSpaceWrite_platformEmptyOK(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":       "admin-1",
		"X-Mini-Auth-Type":     "admin",
		"X-Mini-Platform-Role": "platform_admin",
	})
	spaceID, ok := RequireSpaceWrite(r)
	if !ok {
		t.Fatalf("platform write should ok, code=%s", r.errCode)
	}
	if spaceID != "" {
		t.Fatalf("want empty spaceID, got %q", spaceID)
	}
}

func TestRequireSpace_admin(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "admin",
	})
	spaceID, ok := RequireSpace(r, RoleAdmin)
	if !ok || spaceID != "space-a" {
		t.Fatalf("got spaceID=%q ok=%v err=%s", spaceID, ok, r.errCode)
	}
	r2 := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "member",
	})
	_, ok = RequireSpace(r2, RoleAdmin)
	if ok || r2.errCode != "403" {
		t.Fatalf("member should not pass admin gate, ok=%v code=%s", ok, r2.errCode)
	}
}

func TestRequireSpacePath_match(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":       "user-1",
		"X-Mini-Space-Id":      "space-a",
		"X-Mini-Space-Role":    "member",
		"X-Mini-Param-spaceId": "space-a",
	})
	id, ok := RequireSpacePath(r, true)
	if !ok || id != "space-a" {
		t.Fatalf("got spaceID=%q ok=%v err=%s", id, ok, r.errCode)
	}
}

func TestRequireSpacePath_mismatch(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":       "user-1",
		"X-Mini-Space-Id":      "space-a",
		"X-Mini-Space-Role":    "member",
		"X-Mini-Param-spaceId": "space-b",
	})
	_, ok := RequireSpacePath(r, true)
	if ok || r.errCode != "403" {
		t.Fatalf("want 403 mismatch, ok=%v code=%s", ok, r.errCode)
	}
}

func TestRequireSpacePath_missingPathFallsBackToJWT(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-a",
		"X-Mini-Space-Role": "member",
	})
	id, ok := RequireSpacePath(r, false)
	if !ok || id != "space-a" {
		t.Fatalf("want JWT fallback space-a, got spaceID=%q ok=%v code=%s", id, ok, r.errCode)
	}
}

func TestRequireSpacePath_missingPathAndJWT(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Role": "member",
	})
	_, ok := RequireSpacePath(r, false)
	if ok || r.errCode != "400" {
		t.Fatalf("want 400 when path+JWT space missing, ok=%v code=%s", ok, r.errCode)
	}
}

func TestAuthSpaceID_spaceHeader(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":  "user-1",
		"X-Mini-Space-Id": "space-b",
	})
	if got := AuthSpaceID(r); got != "space-b" {
		t.Fatalf("want space-b, got %q", got)
	}
}

func TestAuthSpaceRole_spaceHeader(t *testing.T) {
	r := reqWith(map[string]string{
		"X-Mini-Subject":    "user-1",
		"X-Mini-Space-Id":   "space-b",
		"X-Mini-Space-Role": "admin",
	})
	if got := AuthSpaceRole(r); got != "admin" {
		t.Fatalf("want admin, got %q", got)
	}
}
