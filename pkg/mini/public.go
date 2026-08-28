package mini

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Metadata keys used to declare public HTTP exposure of an endpoint.
// The gate discovers endpoints via $SRV.INFO and only routes those
// marked with MetaPublic=true and a valid HTTP method/path.
const (
	MetaPublic     = "mini.public"      // "true" to expose via gate
	MetaHTTPMethod = "mini.http.method" // GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
	MetaHTTPPath   = "mini.http.path"   // e.g. /v1/orders
	MetaTimeout    = "mini.timeout"     // optional Go duration, e.g. 3s
	// MetaAuth controls gate auth for this route when global auth is configured:
	//   "required" (default) — must authenticate
	//   "optional"           — try auth; allow anonymous
	//   "none" / "public"    — skip auth entirely
	MetaAuth = "mini.auth"
	// MetaOpenAPIRequest / MetaOpenAPIResponse are optional JSON Schema objects
	// (as compact JSON strings) embedded into gate OpenAPI output.
	MetaOpenAPIRequest  = "mini.openapi.request"
	MetaOpenAPIResponse = "mini.openapi.response"
	MetaOpenAPISummary  = "mini.openapi.summary"

	// Upload / body handling (HTTP only).
	// MetaMaxBody is a size limit for this route (e.g. "32MB", "64m", "1048576").
	// Overrides gate default MaxBodyBytes for non-stream single-shot requests.
	MetaMaxBody = "mini.http.max_body"
	// MetaMultipart when true: gate parses multipart/form-data; form fields
	// become X-Mini-Form-* headers and the first file becomes the body.
	MetaMultipart = "mini.http.multipart"
	// MetaStream when true: gate streams the body to the service as chunked
	// NATS messages (begin/data/end). Use mini.NewStreamHandler on the service.
	MetaStream = "mini.http.stream"
	// MetaStreamChunk optional chunk size (e.g. "256KB"). Default 512KiB.
	MetaStreamChunk = "mini.http.stream_chunk"
	// MetaPolicyCEL is a CEL expression evaluated by the gate after auth.
	// Example: `params.tenant == string(claims.tenant)`
	MetaPolicyCEL = "mini.policy.cel"

	// MetaModule tags an endpoint with the logical module name inside a composite
	// process (e.g. "cron" inside exec, "memo" inside platform). Optional.
	MetaModule = "mini.module"
)

// Auth mode values for MetaAuth / PublicRoute.Auth.
const (
	AuthRequired = "required"
	AuthOptional = "optional"
	AuthNone     = "none"
)

// PublicRoute describes a service endpoint exposed to the outside world
// through the gate (HTTP request/reply or WebSocket real-time).
type PublicRoute struct {
	// Kind is "http" (default) or "ws".
	Kind string `json:"kind,omitempty"`

	// Service is the NATS micro service name.
	Service string `json:"service"`

	// ServiceID is the instance id that advertised this route (informational).
	ServiceID string `json:"service_id,omitempty"`

	// Version is the service SemVer.
	Version string `json:"version,omitempty"`

	// Endpoint is the endpoint name within the service.
	Endpoint string `json:"endpoint"`

	// Subject is the NATS subject the gate should request (HTTP)
	// or the subscribe subject (WS).
	Subject string `json:"subject"`

	// Method is the HTTP method (uppercase), or "WS" for WebSocket routes.
	Method string `json:"method"`

	// Path is the public HTTP/WS path pattern (must start with /).
	// Supports {param} placeholders, e.g. /v1/orders/{id}.
	Path string `json:"path"`

	// Timeout is an optional per-route request timeout (HTTP only).
	Timeout time.Duration `json:"timeout,omitempty"`

	// Auth is required|optional|none (see MetaAuth). Empty means required when
	// the gate has auth configured.
	Auth string `json:"auth,omitempty"`

	// MaxBody is an optional per-route HTTP body size limit (bytes). 0 = gate default.
	MaxBody int64 `json:"max_body,omitempty"`

	// Multipart enables multipart/form-data parsing at the gate.
	Multipart bool `json:"multipart,omitempty"`

	// Stream enables chunked NATS upload protocol for large bodies.
	Stream bool `json:"stream,omitempty"`

	// StreamChunk is the preferred chunk size for Stream mode (bytes). 0 = default.
	StreamChunk int `json:"stream_chunk,omitempty"`

	// Subscribe is NATS → WS subject template (WS only).
	Subscribe string `json:"subscribe,omitempty"`

	// Publish is WS → NATS subject template (WS only; empty = read-only).
	Publish string `json:"publish,omitempty"`

	// Metadata is the full endpoint metadata map.
	Metadata map[string]string `json:"metadata,omitempty"`

	// pattern is compiled from Path (not serialized).
	pattern *PathPattern `json:"-"`
}

// RouteKey returns a stable routing key "METHOD PATH".
func (r PublicRoute) RouteKey() string {
	return r.Method + " " + r.Path
}

// Pattern returns the compiled path pattern (lazy, cached on the route).
func (r *PublicRoute) Pattern() *PathPattern {
	if r == nil {
		return nil
	}
	if r.pattern != nil {
		return r.pattern
	}
	p, err := ParsePathPattern(r.Path)
	if err != nil {
		return nil
	}
	r.pattern = p
	return p
}

// MatchPath matches an HTTP path against this route's pattern.
func (r *PublicRoute) MatchPath(path string) (map[string]string, bool) {
	p := r.Pattern()
	if p == nil {
		return nil, false
	}
	return p.Match(path)
}

// WithPublicHTTP marks an endpoint as publicly reachable via the gate
// with the given HTTP method and path.
func WithPublicHTTP(method, path string) EndpointOpt {
	return func(e *endpointOpts) error {
		method = strings.ToUpper(strings.TrimSpace(method))
		path = strings.TrimSpace(path)
		if err := validatePublicHTTP(method, path); err != nil {
			return fmt.Errorf("%w: %s", ErrConfigValidation, err)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaPublic] = "true"
		e.metadata[MetaTransport] = TransportHTTP
		e.metadata[MetaHTTPMethod] = method
		e.metadata[MetaHTTPPath] = path
		return nil
	}
}

// WithPublicTimeout sets the suggested gate timeout for a public endpoint.
func WithPublicTimeout(d time.Duration) EndpointOpt {
	return func(e *endpointOpts) error {
		if d <= 0 {
			return fmt.Errorf("%w: public timeout must be positive", ErrConfigValidation)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaTimeout] = d.String()
		return nil
	}
}

// WithPublicAuth sets mini.auth metadata (required|optional|none).
func WithPublicAuth(mode string) EndpointOpt {
	return func(e *endpointOpts) error {
		mode = normalizeAuthMode(mode)
		if mode == "" {
			return fmt.Errorf("%w: auth mode must be required, optional, or none", ErrConfigValidation)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaAuth] = mode
		return nil
	}
}

// WithOpenAPISchema attaches optional request/response JSON Schema (raw JSON) for OpenAPI.
func WithOpenAPISchema(requestSchemaJSON, responseSchemaJSON string) EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		if requestSchemaJSON != "" {
			e.metadata[MetaOpenAPIRequest] = requestSchemaJSON
		}
		if responseSchemaJSON != "" {
			e.metadata[MetaOpenAPIResponse] = responseSchemaJSON
		}
		return nil
	}
}

// WithOpenAPISummary sets a short summary for OpenAPI/Catalog documentation.
func WithOpenAPISummary(summary string) EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaOpenAPISummary] = summary
		return nil
	}
}

// WithPublicMaxBody sets mini.http.max_body (bytes). Use for larger-than-default single-shot uploads.
func WithPublicMaxBody(bytes int64) EndpointOpt {
	return func(e *endpointOpts) error {
		if bytes <= 0 {
			return fmt.Errorf("%w: max body must be positive", ErrConfigValidation)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaMaxBody] = fmt.Sprintf("%d", bytes)
		return nil
	}
}

// WithPublicMultipart marks the public route so the gate parses multipart/form-data.
func WithPublicMultipart() EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaMultipart] = "true"
		return nil
	}
}

// WithPublicStream enables chunked NATS streaming uploads for this public route.
// Pair with NewStreamHandler on the service.
func WithPublicStream() EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaStream] = "true"
		return nil
	}
}

// WithPublicStreamChunk sets the preferred stream chunk size in bytes.
func WithPublicStreamChunk(bytes int) EndpointOpt {
	return func(e *endpointOpts) error {
		if bytes < 1024 {
			return fmt.Errorf("%w: stream chunk must be >= 1024", ErrConfigValidation)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaStreamChunk] = fmt.Sprintf("%d", bytes)
		return nil
	}
}

// WithPublicPolicyCEL attaches a CEL expression enforced by the gate after auth.
// Expression variables: method, path, params, headers, claims, auth_type, service,
// endpoint, subject, kind, client_cn, client_san, client_serial.
func WithPublicPolicyCEL(expr string) EndpointOpt {
	return func(e *endpointOpts) error {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			return fmt.Errorf("%w: policy CEL expression required", ErrConfigValidation)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaPolicyCEL] = expr
		return nil
	}
}

// ParseByteSize parses sizes like "1024", "32KB", "32KiB", "2MB", "2MiB", "1G".
func ParseByteSize(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty size")
	}
	// plain integer
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if n < 0 {
			return 0, fmt.Errorf("negative size")
		}
		return n, nil
	}
	s = strings.ToUpper(s)
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "KIB"):
		mult = 1024
		s = strings.TrimSuffix(s, "KIB")
	case strings.HasSuffix(s, "MIB"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "MIB")
	case strings.HasSuffix(s, "GIB"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "GIB")
	case strings.HasSuffix(s, "KB"):
		mult = 1000
		s = strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "MB"):
		mult = 1000 * 1000
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "GB"):
		mult = 1000 * 1000 * 1000
		s = strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "K"):
		mult = 1024
		s = strings.TrimSuffix(s, "K")
	case strings.HasSuffix(s, "M"):
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "M")
	case strings.HasSuffix(s, "G"):
		mult = 1024 * 1024 * 1024
		s = strings.TrimSuffix(s, "G")
	case strings.HasSuffix(s, "B"):
		s = strings.TrimSuffix(s, "B")
	}
	s = strings.TrimSpace(s)
	n, err := strconv.ParseFloat(s, 64)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("invalid size %q", s)
	}
	return int64(n * float64(mult)), nil
}

func normalizeAuthMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case AuthRequired, "req", "must":
		return AuthRequired
	case AuthOptional, "opt":
		return AuthOptional
	case AuthNone, "public", "skip", "false", "0":
		return AuthNone
	default:
		return ""
	}
}

func authModeFromMetadata(md map[string]string) string {
	if md == nil {
		return ""
	}
	return normalizeAuthMode(md[MetaAuth])
}

// PublicRoutesFromInfo extracts public HTTP and WebSocket routes from INFO.
func PublicRoutesFromInfo(info Info) []PublicRoute {
	var routes []PublicRoute
	for _, ep := range info.Endpoints {
		if r, ok := publicWSRouteFromEndpoint(info, ep); ok {
			routes = append(routes, r)
			continue
		}
		if r, ok := publicHTTPRouteFromEndpoint(info, ep); ok {
			routes = append(routes, r)
		}
	}
	return routes
}

func publicHTTPRouteFromEndpoint(info Info, ep EndpointInfo) (PublicRoute, bool) {
	if ep.Metadata == nil {
		return PublicRoute{}, false
	}
	if !isTruthy(ep.Metadata[MetaPublic]) {
		return PublicRoute{}, false
	}
	// Skip WS-only advertisements.
	if strings.EqualFold(ep.Metadata[MetaTransport], TransportWS) || ep.Metadata[MetaWSPath] != "" {
		if ep.Metadata[MetaHTTPPath] == "" {
			return PublicRoute{}, false
		}
	}
	method := strings.ToUpper(strings.TrimSpace(ep.Metadata[MetaHTTPMethod]))
	path := strings.TrimSpace(ep.Metadata[MetaHTTPPath])
	if validatePublicHTTP(method, path) != nil {
		return PublicRoute{}, false
	}

	var timeout time.Duration
	if raw := strings.TrimSpace(ep.Metadata[MetaTimeout]); raw != "" {
		if d, err := time.ParseDuration(raw); err == nil && d > 0 {
			timeout = d
		}
	}

	var maxBody int64
	if raw := strings.TrimSpace(ep.Metadata[MetaMaxBody]); raw != "" {
		if n, err := ParseByteSize(raw); err == nil && n > 0 {
			maxBody = n
		}
	}
	var streamChunk int
	if raw := strings.TrimSpace(ep.Metadata[MetaStreamChunk]); raw != "" {
		if n, err := ParseByteSize(raw); err == nil && n >= 1024 {
			streamChunk = int(n)
		}
	}

	pat, err := ParsePathPattern(path)
	if err != nil {
		return PublicRoute{}, false
	}

	return PublicRoute{
		Kind:        TransportHTTP,
		Service:     info.Name,
		ServiceID:   info.ID,
		Version:     info.Version,
		Endpoint:    ep.Name,
		Subject:     ep.Subject,
		Method:      method,
		Path:        pat.Raw,
		Timeout:     timeout,
		Auth:        authModeFromMetadata(ep.Metadata),
		MaxBody:     maxBody,
		Multipart:   isTruthy(ep.Metadata[MetaMultipart]),
		Stream:      isTruthy(ep.Metadata[MetaStream]),
		StreamChunk: streamChunk,
		Metadata:    ep.Metadata,
		pattern:     pat,
	}, true
}

func validatePublicHTTP(method, path string) error {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodHead, http.MethodOptions:
	default:
		return fmt.Errorf("invalid HTTP method %q", method)
	}
	if _, err := ParsePathPattern(path); err != nil {
		return err
	}
	return nil
}

func isTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
