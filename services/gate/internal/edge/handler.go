package edge

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	loggmodels "github.com/pafthang/arcanum/services/logg/models"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/gate/internal/auth"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
)

// ClaimMap maps JWT claim name → outbound header (from CLAIM_HEADERS).
// Always injects X-Mini-Subject from sub.
type ClaimMap map[string]string

// Handler is the NATS-backed public HTTP edge.
type Handler struct {
	NC        *nats.Conn
	Table     *Table
	Auth      []auth.Authenticator // try in order for required/optional
	Claims    ClaimMap
	DefaultTO time.Duration
	MaxBody   int64
	Client    *mini.Client
	Log       *slog.Logger
	Logg      *loggclient.Client
}

// ServeHTTP implements http.Handler for discovered public routes + catalog.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.Table == nil {
		http.Error(w, "gate not ready", http.StatusServiceUnavailable)
		return
	}

	path := r.URL.Path
	// Built-in edge endpoints
	switch {
	case path == "/_catalog" && (r.Method == http.MethodGet || r.Method == http.MethodHead):
		h.serveCatalog(w, r)
		return
	case path == "/healthz" || path == "/readyz":
		// Should be registered before this handler; keep as fallback.
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
		return
	}

	// Catalog WebSocket routes: NATS Subscribe/Publish bridge.
	if isWebSocketUpgrade(r) {
		h.serveWS(w, r)
		return
	}

	route, params, ok := h.Table.Match(r.Method, path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Per-route auth
	mode := strings.ToLower(strings.TrimSpace(route.Auth))
	if mode == "" || mode == "public" {
		mode = mini.AuthRequired
	}
	if mode == "public" {
		mode = mini.AuthNone
	}

	var ident *auth.Identity
	switch mode {
	case mini.AuthNone:
		// skip
	case mini.AuthOptional:
		ident, _ = h.tryAuth(r)
	default: // required
		var err error
		ident, err = h.requireAuth(r)
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	maxBody := h.MaxBody
	if maxBody <= 0 {
		maxBody = 32 << 20
	}
	if route.MaxBody > 0 && route.MaxBody < maxBody {
		maxBody = route.MaxBody
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			http.Error(w, "payload too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	hdrs := mini.Headers{}
	nh := nats.Header(hdrs)
	// Forward selected inbound headers (Set = MIME-canonical keys)
	if ct := r.Header.Get("Content-Type"); ct != "" {
		nh.Set("Content-Type", ct)
	}
	if rid := r.Header.Get("X-Request-ID"); rid != "" {
		nh.Set("X-Request-ID", rid)
	}
	// Query string for httpx.QueryValues / Query
	if q := r.URL.RawQuery; q != "" {
		nh.Set("X-Query-String", q)
		for k, vs := range r.URL.Query() {
			if len(vs) > 0 {
				nh.Set("X-Mini-Query-"+k, vs[0])
			}
		}
	}
	hdrs = mini.Headers(nh)

	// Multipart routes: gate unwraps form fields → X-Mini-Form-* and file → body.
	// Without this, services receive raw multipart bytes and lose folder/filename.
	if route.Multipart {
		file, form, isMP, merr := parseMultipartBody(body, r.Header.Get("Content-Type"))
		if merr != nil {
			http.Error(w, "bad multipart body", http.StatusBadRequest)
			return
		}
		if isMP {
			body = file.Data
			applyMultipartHeaders(hdrs, file, form)
		}
	}

	// Path params
	hdrs = mini.ApplyPathParams(hdrs, params)
	// Auth claims → headers
	if ident != nil {
		injectClaims(hdrs, ident, h.Claims)
	}

	timeout := h.DefaultTO
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	if route.Timeout > 0 {
		timeout = route.Timeout
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()

	client := h.Client
	if client == nil {
		c, err := mini.NewClient(h.NC, mini.WithClientTimeout(timeout))
		if err != nil {
			http.Error(w, "gate misconfigured", http.StatusInternalServerError)
			return
		}
		client = c
	}

	res, err := client.Request(ctx, route.Subject, body, hdrs)
	if err != nil {
		var se *mini.ServiceError
		if errors.As(err, &se) {
			status := se.HTTPStatus()
			// Prefer explicit status header if present
			if s := se.Headers.Get(mini.StatusHeader); s != "" {
				if n, e := strconv.Atoi(s); e == nil && n >= 100 && n <= 599 {
					status = n
				}
			}
			copyResponseHeaders(w.Header(), se.Headers)
			w.Header().Set("Content-Type", contentType(se.Headers, se.Data))
			w.WriteHeader(status)
			if len(se.Data) > 0 {
				_, _ = w.Write(se.Data)
			} else if se.Description != "" {
				_, _ = w.Write([]byte(se.Description))
			}
			return
		}
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, nats.ErrTimeout) {
			http.Error(w, "upstream timeout", http.StatusGatewayTimeout)
			return
		}
		if errors.Is(err, nats.ErrNoResponders) {
			http.Error(w, "no responders", http.StatusBadGateway)
			return
		}
		if h.Log != nil {
			h.Log.Warn("edge request failed", "subject", route.Subject, "err", err)
		}
		http.Error(w, "bad gate", http.StatusBadGateway)
		return
	}

	status := http.StatusOK
	if s := res.Headers.Get(mini.StatusHeader); s != "" {
		if n, e := strconv.Atoi(s); e == nil && n >= 100 && n <= 599 {
			status = n
		}
	}
	copyResponseHeaders(w.Header(), res.Headers)
	w.Header().Set("Content-Type", contentType(res.Headers, res.Data))
	w.WriteHeader(status)
	if len(res.Data) > 0 && r.Method != http.MethodHead {
		_, _ = w.Write(res.Data)
	}
}

func (h *Handler) serveCatalog(w http.ResponseWriter, r *http.Request) {
	routes := h.Table.List()
	f := mini.CatalogFilter{
		Service: r.URL.Query().Get("service"),
		Kind:    r.URL.Query().Get("kind"),
		Auth:    r.URL.Query().Get("auth"),
		Tag:     r.URL.Query().Get("tag"),
		Q:       r.URL.Query().Get("q"),
	}
	cat := mini.BuildCatalogFiltered(routes, f)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ETag", `"`+cat.Revision+`"`)
	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}
	_ = json.NewEncoder(w).Encode(cat)
}

func (h *Handler) tryAuth(r *http.Request) (*auth.Identity, error) {
	var lastErr error
	for _, a := range h.Auth {
		if a == nil {
			continue
		}
		id, err := a.Authenticate(r)
		if err == nil {
			if ident, ok := id.(*auth.Identity); ok {
				return ident, nil
			}
		}
		if err != auth.ErrUnauthorized && err.Error() != "missing authorization header" && err.Error() != "missing access_token" && err.Error() != "missing client certificate" {
			lastErr = err
		}
	}
	if lastErr != nil && h.Logg != nil {
		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		h.Logg.AppendActivityAsync(&loggmodels.Activity{
			Type:    "gate.auth.failed",
			Summary: "Authentication failed",
			Payload: map[string]any{
				"path": r.URL.Path,
				"ip":   ip,
				"err":  lastErr.Error(),
			},
		})
	}
	return nil, auth.ErrUnauthorized
}

func (h *Handler) requireAuth(r *http.Request) (*auth.Identity, error) {
	return h.tryAuth(r)
}

func injectClaims(h mini.Headers, id *auth.Identity, claimMap ClaimMap) {
	if id == nil || h == nil {
		return
	}
	nh := nats.Header(h)
	set := func(k, v string) {
		if v != "" {
			nh.Set(k, v)
		}
	}
	set("X-Mini-Subject", id.Subject)
	// Default claim → header map if none configured
	cm := claimMap
	if len(cm) == 0 {
		cm = ClaimMap{
			"typ":           "X-Mini-Auth-Type",
			"platform_role": "X-Mini-Platform-Role",
			"space_id":      "X-Mini-Space-Id",
			"space_role":    "X-Mini-Space-Role",
			"email":         "X-Mini-Email",
			"role":          "X-Mini-Role",
			"tv":            "X-Mini-Token-Version",
		}
	}
	for claim, header := range cm {
		if v, ok := id.Claims[claim]; ok && v != nil {
			set(header, claimString(v))
		}
	}
	if nh.Get("X-Mini-Token-Version") == "" {
		if v, ok := id.Claims["tv"]; ok {
			set("X-Mini-Token-Version", claimString(v))
		} else if v, ok := id.Claims["jti"]; ok {
			set("X-Mini-Token-Version", claimString(v))
		}
	}
}

func claimString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		// JWT numeric dates sometimes appear as float; skip non-int claims
		if t == float64(int64(t)) {
			return strconv.FormatInt(int64(t), 10)
		}
		return strconv.FormatFloat(t, 'f', -1, 64)
	case bool:
		if t {
			return "true"
		}
		return "false"
	case json.Number:
		return t.String()
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		s := string(b)
		if len(s) >= 2 && s[0] == '"' {
			return strings.Trim(s, `"`)
		}
		return s
	}
}

func copyResponseHeaders(dst http.Header, src mini.Headers) {
	for k, vs := range src {
		// Strip internal NATS service headers from client response
		switch k {
		case mini.StatusHeader, mini.ErrorHeader, mini.ErrorCodeHeader:
			continue
		}
		if strings.HasPrefix(k, "Nats-") || strings.HasPrefix(k, "Nats-Service") {
			continue
		}
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}

func contentType(h mini.Headers, data []byte) string {
	if ct := h.Get("Content-Type"); ct != "" {
		return ct
	}
	if len(data) > 0 && (data[0] == '{' || data[0] == '[') {
		return "application/json"
	}
	return "application/octet-stream"
}

// ParseClaimHeaders parses "claim:Header,claim2:Header2" env format.
func ParseClaimHeaders(raw string) ClaimMap {
	out := ClaimMap{}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		i := strings.IndexByte(part, ':')
		if i <= 0 {
			continue
		}
		claim := strings.TrimSpace(part[:i])
		header := strings.TrimSpace(part[i+1:])
		if claim != "" && header != "" {
			out[claim] = header
		}
	}
	return out
}
