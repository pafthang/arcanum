package mini

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
)

// CatalogVersion is the schema version of Catalog JSON.
const CatalogVersion = 1

// MetaCatalogTags is optional comma-separated tags for catalog grouping
// (e.g. "auth,public").
const MetaCatalogTags = "mini.catalog.tags"

// Catalog is the platform API contract: HTTP + WebSocket public routes in one
// document. Source of truth for clients/codegen; OpenAPI is a derived view.
//
// Served by the gate at GET /_catalog (built from the live route table).
// Query filters: ?service=&kind=&auth=&tag=&q=
type Catalog struct {
	// Version is the catalog schema version (not service versions).
	Version int `json:"version"`

	// GeneratedAt is when this snapshot was built (UTC).
	GeneratedAt time.Time `json:"generated_at"`

	// Revision is a stable hash of route ids + subjects (for cache/ETag).
	Revision string `json:"revision"`

	// Count is len(Routes).
	Count int `json:"count"`

	// Services is an index of services present in this snapshot.
	Services []CatalogService `json:"services,omitempty"`

	// Routes is sorted by service then id.
	Routes []CatalogRoute `json:"routes"`
}

// CatalogService summarizes one service in the catalog.
type CatalogService struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
	HTTP    int    `json:"http"`
	WS      int    `json:"ws"`
	Routes  int    `json:"routes"`
}

// CatalogRoute is one public edge operation (HTTP or WS).
type CatalogRoute struct {
	// ID is a stable identifier: "{service}.{endpoint}" (disambiguated if needed).
	ID string `json:"id"`

	// Kind is "http" or "ws".
	Kind string `json:"kind"`

	Service  string `json:"service"`
	Endpoint string `json:"endpoint"`
	Subject  string `json:"subject,omitempty"`

	// Version is the advertising service SemVer (informational).
	Version string `json:"version,omitempty"`

	// Auth is always set: required|optional|none.
	Auth string `json:"auth"`

	// Summary is a short human description.
	Summary string `json:"summary,omitempty"`

	// Tags optional grouping labels.
	Tags []string `json:"tags,omitempty"`

	// TimeoutMs is the suggested gate timeout (HTTP), if advertised.
	TimeoutMs int64 `json:"timeout_ms,omitempty"`

	// HTTP is set for kind=http (and for ws path exposure).
	HTTP *CatalogHTTP `json:"http,omitempty"`

	// WS is set for kind=ws.
	WS *CatalogWS `json:"ws,omitempty"`

	// Request / Response are optional JSON Schema objects (from metadata).
	Request  json.RawMessage `json:"request,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`

	// PolicyCEL is an optional gate CEL expression.
	PolicyCEL string `json:"policy_cel,omitempty"`

	// Multipart / Stream describe body handling (HTTP).
	Multipart   bool  `json:"multipart,omitempty"`
	Stream      bool  `json:"stream,omitempty"`
	StreamChunk int   `json:"stream_chunk,omitempty"`
	MaxBody     int64 `json:"max_body,omitempty"`
}

// CatalogHTTP is the HTTP edge binding.
type CatalogHTTP struct {
	Method string   `json:"method,omitempty"` // empty for pure WS upgrade path
	Path   string   `json:"path"`
	Params []string `json:"params,omitempty"` // path param names in order
}

// CatalogWS is the WebSocket edge binding.
type CatalogWS struct {
	Path      string `json:"path"`
	Subscribe string `json:"subscribe,omitempty"`
	Publish   string `json:"publish,omitempty"`
	// ReadOnly is true when Publish is empty.
	ReadOnly bool `json:"read_only,omitempty"`
}

// CatalogFilter selects a subset of routes (gate query params).
type CatalogFilter struct {
	Service string // exact service name
	Kind    string // http | ws
	Auth    string // required | optional | none
	Tag     string // must include tag
	Q       string // substring match on id, path, summary, subject
}

// BuildCatalog builds a Catalog snapshot from public routes (e.g. gate table).
func BuildCatalog(routes []PublicRoute) Catalog {
	return BuildCatalogFiltered(routes, CatalogFilter{})
}

// BuildCatalogFiltered builds a catalog and applies filter (empty filter = all).
func BuildCatalogFiltered(routes []PublicRoute, f CatalogFilter) Catalog {
	now := time.Now().UTC().Truncate(time.Second)
	out := make([]CatalogRoute, 0, len(routes))
	seenID := map[string]int{}

	for _, r := range routes {
		cr := catalogRouteFromPublic(r)
		// Disambiguate duplicate service.endpoint.
		base := cr.ID
		if n, ok := seenID[base]; ok {
			seenID[base] = n + 1
			cr.ID = fmt.Sprintf("%s#%d", base, n+1)
		} else {
			seenID[base] = 1
		}
		out = append(out, cr)
	}

	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Service != out[j].Service {
			return out[i].Service < out[j].Service
		}
		if out[i].ID != out[j].ID {
			return out[i].ID < out[j].ID
		}
		return out[i].Subject < out[j].Subject
	})

	// Full revision before filter (stable for ETag of complete table).
	fullRev := catalogRevision(out)
	out = applyCatalogFilter(out, f)

	return Catalog{
		Version:     CatalogVersion,
		GeneratedAt: now,
		Revision:    fullRev,
		Count:       len(out),
		Services:    buildServiceIndex(out),
		Routes:      out,
	}
}

// FilterCatalog returns a copy of cat with routes filtered (revision preserved).
func FilterCatalog(cat Catalog, f CatalogFilter) Catalog {
	routes := applyCatalogFilter(cat.Routes, f)
	cat.Routes = routes
	cat.Count = len(routes)
	cat.Services = buildServiceIndex(routes)
	return cat
}

// Find returns a route by id or "service.endpoint".
func (c Catalog) Find(id string) *CatalogRoute {
	for i := range c.Routes {
		r := &c.Routes[i]
		if r.ID == id || r.Service+"."+r.Endpoint == id {
			return r
		}
	}
	return nil
}

// FindHTTP finds the first HTTP route matching method + path pattern.
func (c Catalog) FindHTTP(method, path string) *CatalogRoute {
	method = strings.ToUpper(strings.TrimSpace(method))
	for i := range c.Routes {
		r := &c.Routes[i]
		if r.Kind != TransportHTTP && r.Kind != "" {
			continue
		}
		if r.HTTP == nil {
			continue
		}
		if strings.ToUpper(r.HTTP.Method) == method && r.HTTP.Path == path {
			return r
		}
	}
	return nil
}

func catalogRouteFromPublic(r PublicRoute) CatalogRoute {
	kind := strings.ToLower(strings.TrimSpace(r.Kind))
	if kind == "" {
		if r.Method == WSMethod || r.Subscribe != "" {
			kind = TransportWS
		} else {
			kind = TransportHTTP
		}
	}

	id := CatalogRouteID(r)
	auth := normalizeAuthMode(strings.TrimSpace(r.Auth))
	if auth == "" {
		auth = AuthRequired
	}
	summary := strings.TrimSpace(r.Service + " " + r.Endpoint)
	if r.Metadata != nil {
		if s := strings.TrimSpace(r.Metadata[MetaOpenAPISummary]); s != "" {
			summary = s
		}
	}

	cr := CatalogRoute{
		ID:          id,
		Kind:        kind,
		Service:     r.Service,
		Endpoint:    r.Endpoint,
		Subject:     r.Subject,
		Version:     r.Version,
		Auth:        auth,
		Summary:     summary,
		Tags:        catalogTags(r),
		Multipart:   r.Multipart,
		Stream:      r.Stream,
		StreamChunk: r.StreamChunk,
		MaxBody:     r.MaxBody,
		Request:     rawSchema(r.Metadata, MetaOpenAPIRequest),
		Response:    rawSchema(r.Metadata, MetaOpenAPIResponse),
	}
	if r.Timeout > 0 {
		cr.TimeoutMs = r.Timeout.Milliseconds()
	}
	if r.Metadata != nil {
		cr.PolicyCEL = strings.TrimSpace(r.Metadata[MetaPolicyCEL])
	}

	path := r.Path
	params := pathParamNames(path)

	switch kind {
	case TransportWS:
		cr.WS = &CatalogWS{
			Path:      path,
			Subscribe: r.Subscribe,
			Publish:   r.Publish,
			ReadOnly:  strings.TrimSpace(r.Publish) == "",
		}
		// WS is still reached via HTTP upgrade on the same path.
		cr.HTTP = &CatalogHTTP{
			Method: "",
			Path:   path,
			Params: params,
		}
		if cr.Subject == "" {
			cr.Subject = r.Subscribe
		}
	default:
		method := strings.ToUpper(strings.TrimSpace(r.Method))
		cr.HTTP = &CatalogHTTP{
			Method: method,
			Path:   path,
			Params: params,
		}
	}
	return cr
}

// CatalogRouteID returns a stable id for a public route.
func CatalogRouteID(r PublicRoute) string {
	svc := strings.TrimSpace(r.Service)
	ep := strings.TrimSpace(r.Endpoint)
	if svc != "" && ep != "" {
		return svc + "." + ep
	}
	if svc != "" {
		return svc + "." + strings.ToLower(r.Method) + ":" + r.Path
	}
	return strings.ToLower(r.Method) + ":" + r.Path
}

func catalogTags(r PublicRoute) []string {
	var tags []string
	if r.Service != "" {
		tags = append(tags, r.Service)
	}
	if r.Metadata != nil {
		raw := strings.TrimSpace(r.Metadata[MetaCatalogTags])
		if raw != "" {
			for _, p := range strings.Split(raw, ",") {
				p = strings.TrimSpace(p)
				if p != "" {
					tags = append(tags, p)
				}
			}
		}
	}
	// de-dupe
	if len(tags) <= 1 {
		return tags
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}

func buildServiceIndex(routes []CatalogRoute) []CatalogService {
	type agg struct {
		ver  string
		http int
		ws   int
		n    int
	}
	m := map[string]*agg{}
	var order []string
	for _, r := range routes {
		name := r.Service
		if name == "" {
			name = "_"
		}
		a, ok := m[name]
		if !ok {
			a = &agg{ver: r.Version}
			m[name] = a
			order = append(order, name)
		}
		if a.ver == "" && r.Version != "" {
			a.ver = r.Version
		}
		a.n++
		if r.Kind == TransportWS {
			a.ws++
		} else {
			a.http++
		}
	}
	sort.Strings(order)
	out := make([]CatalogService, 0, len(order))
	for _, name := range order {
		a := m[name]
		out = append(out, CatalogService{
			Name:    name,
			Version: a.ver,
			HTTP:    a.http,
			WS:      a.ws,
			Routes:  a.n,
		})
	}
	return out
}

func applyCatalogFilter(routes []CatalogRoute, f CatalogFilter) []CatalogRoute {
	f.Service = strings.TrimSpace(f.Service)
	f.Kind = strings.ToLower(strings.TrimSpace(f.Kind))
	f.Auth = normalizeAuthMode(strings.TrimSpace(f.Auth))
	f.Tag = strings.TrimSpace(f.Tag)
	f.Q = strings.ToLower(strings.TrimSpace(f.Q))
	if f.Service == "" && f.Kind == "" && f.Auth == "" && f.Tag == "" && f.Q == "" {
		return routes
	}
	out := make([]CatalogRoute, 0, len(routes))
	for _, r := range routes {
		if f.Service != "" && r.Service != f.Service {
			continue
		}
		if f.Kind != "" && strings.ToLower(r.Kind) != f.Kind {
			continue
		}
		if f.Auth != "" && r.Auth != f.Auth {
			continue
		}
		if f.Tag != "" && !hasTag(r.Tags, f.Tag) {
			continue
		}
		if f.Q != "" && !catalogMatchQ(r, f.Q) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func hasTag(tags []string, want string) bool {
	for _, t := range tags {
		if t == want {
			return true
		}
	}
	return false
}

func catalogMatchQ(r CatalogRoute, q string) bool {
	if strings.Contains(strings.ToLower(r.ID), q) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Summary), q) {
		return true
	}
	if strings.Contains(strings.ToLower(r.Subject), q) {
		return true
	}
	if r.HTTP != nil && strings.Contains(strings.ToLower(r.HTTP.Path), q) {
		return true
	}
	if r.WS != nil {
		if strings.Contains(strings.ToLower(r.WS.Path), q) {
			return true
		}
		if strings.Contains(strings.ToLower(r.WS.Subscribe), q) {
			return true
		}
	}
	return false
}

func rawSchema(md map[string]string, key string) json.RawMessage {
	if md == nil {
		return nil
	}
	raw := strings.TrimSpace(md[key])
	if raw == "" {
		return nil
	}
	if !json.Valid([]byte(raw)) {
		return nil
	}
	return json.RawMessage(raw)
}

func pathParamNames(path string) []string {
	p, err := ParsePathPattern(path)
	if err != nil || p == nil {
		return scanPathParamNames(path)
	}
	var names []string
	for _, seg := range p.segments {
		switch seg.kind {
		case segParam, segCatchAll:
			if seg.value != "" {
				names = append(names, seg.value)
			}
		case segWildcard:
			names = append(names, WildcardParam)
		}
	}
	return names
}

func scanPathParamNames(path string) []string {
	var names []string
	for _, part := range strings.Split(path, "/") {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			n := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			n = strings.TrimPrefix(n, "*")
			if n != "" {
				names = append(names, n)
			}
		} else if part == "*" {
			names = append(names, WildcardParam)
		}
	}
	return names
}

func catalogRevision(routes []CatalogRoute) string {
	h := sha256.New()
	for _, r := range routes {
		_, _ = fmt.Fprintf(h, "%s|%s|%s|%s|%s|%d\n", r.ID, r.Kind, r.Subject, r.Auth, httpSig(r), r.TimeoutMs)
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

func httpSig(r CatalogRoute) string {
	if r.HTTP == nil && r.WS == nil {
		return ""
	}
	var b strings.Builder
	if r.HTTP != nil {
		b.WriteString(r.HTTP.Method)
		b.WriteByte(' ')
		b.WriteString(r.HTTP.Path)
	}
	if r.WS != nil {
		b.WriteByte('|')
		b.WriteString(r.WS.Subscribe)
		b.WriteByte('>')
		b.WriteString(r.WS.Publish)
	}
	return b.String()
}
