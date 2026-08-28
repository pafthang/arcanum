// Package edgecfg is the mini config plane stored entirely in NATS KV.
//
// Bucket layout (default bucket "MINI_CP"):
//
//	routes.<id>     — public HTTP/WS route documents
//	wsacl.<id>      — WebSocket ACL rules
//	meta.revision   — monotonic revision hint (optional)
//
// Gateways watch routes.> and wsacl.> and rebuild their routing table.
// Domain services (or CI/operators) Put/Delete documents via Client.
package edgecfg

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/mini"
)

const (
	// DefaultBucket is the JetStream KV bucket for config plane state.
	DefaultBucket = "MINI_CP"

	PrefixRoutes = "routes."
	PrefixWSACL  = "wsacl."
	KeyRevision  = "meta.revision"
)

// SchemaVersion is the config-plane document format version.
const SchemaVersion = 1

// RouteDoc is a config-plane public route (source of truth for the gate).
type RouteDoc struct {
	// ID is stable identity; if empty, derived from method+path.
	ID string `json:"id,omitempty"`

	// Schema is the document schema version (default SchemaVersion).
	Schema int `json:"schema,omitempty"`

	Kind      string        `json:"kind,omitempty"` // http|ws
	Service   string        `json:"service,omitempty"`
	Endpoint  string        `json:"endpoint,omitempty"`
	Method    string        `json:"method,omitempty"`
	Path      string        `json:"path"`
	Subject   string        `json:"subject,omitempty"`
	Subscribe string        `json:"subscribe,omitempty"`
	Publish   string        `json:"publish,omitempty"`
	Timeout   time.Duration `json:"timeout,omitempty"`
	Auth      string        `json:"auth,omitempty"` // required|optional|none

	// Enabled when false is treated as deleted for gateways.
	Enabled *bool `json:"enabled,omitempty"`

	// UpdatedAt is set by the client on write.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// WSACLDoc is a config-plane WebSocket ACL rule.
type WSACLDoc struct {
	ID          string            `json:"id,omitempty"`
	Path        string            `json:"path,omitempty"`
	ClaimEquals map[string]string `json:"claim_equals,omitempty"`
	Roles       []string          `json:"roles,omitempty"`
	RequireAuth bool              `json:"require_auth,omitempty"`
	DenyPublish bool              `json:"deny_publish,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
}

// RouteID returns a stable KV-safe id for a method+path (or explicit ID).
func RouteID(method, path string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = mini.WSMethod
	}
	path = strings.TrimSpace(path)
	sum := sha256.Sum256([]byte(method + " " + path))
	return hex.EncodeToString(sum[:12])
}

// Normalize fills ID and defaults.
func (d *RouteDoc) Normalize() error {
	if d == nil {
		return fmt.Errorf("nil route")
	}
	if d.Schema == 0 {
		d.Schema = SchemaVersion
	}
	if d.Schema != SchemaVersion {
		return fmt.Errorf("unsupported schema version %d (want %d)", d.Schema, SchemaVersion)
	}
	d.Path = strings.TrimSpace(d.Path)
	if d.Path == "" {
		return fmt.Errorf("path required")
	}
	if _, err := mini.ParsePathPattern(d.Path); err != nil {
		return fmt.Errorf("path: %w", err)
	}
	kind := strings.ToLower(strings.TrimSpace(d.Kind))
	if kind == "" {
		if d.Subscribe != "" || strings.EqualFold(d.Method, mini.WSMethod) {
			kind = mini.TransportWS
		} else {
			kind = mini.TransportHTTP
		}
	}
	d.Kind = kind
	if kind == mini.TransportWS {
		d.Method = mini.WSMethod
		if strings.TrimSpace(d.Subscribe) == "" {
			return fmt.Errorf("ws route requires subscribe")
		}
		if !mini.IsPublicSubject(d.Subscribe) {
			return fmt.Errorf("ws subscribe must be under public.>")
		}
		if d.Publish != "" && !mini.IsPublicSubject(d.Publish) {
			return fmt.Errorf("ws publish must be under public.>")
		}
	} else {
		d.Method = strings.ToUpper(strings.TrimSpace(d.Method))
		if d.Method == "" {
			return fmt.Errorf("http route requires method")
		}
		if strings.TrimSpace(d.Subject) == "" {
			return fmt.Errorf("http route requires subject")
		}
		if !mini.IsPublicSubject(d.Subject) {
			return fmt.Errorf("http subject must be under public.>")
		}
	}
	if d.Auth != "" {
		switch strings.ToLower(strings.TrimSpace(d.Auth)) {
		case mini.AuthRequired, mini.AuthOptional, mini.AuthNone:
			d.Auth = strings.ToLower(strings.TrimSpace(d.Auth))
		case "public":
			d.Auth = mini.AuthNone
		default:
			return fmt.Errorf("invalid auth mode %q", d.Auth)
		}
	}
	if d.ID == "" {
		d.ID = RouteID(d.Method, d.Path)
	}
	if d.Enabled == nil {
		t := true
		d.Enabled = &t
	}
	return nil
}

// Validate is an alias for Normalize (strict validation for dry-run tooling).
func (d *RouteDoc) Validate() error {
	return d.Normalize()
}

// IsEnabled reports whether the doc is active.
func (d *RouteDoc) IsEnabled() bool {
	return d != nil && (d.Enabled == nil || *d.Enabled)
}

// ToPublicRoute converts to gate/mini public route.
func (d RouteDoc) ToPublicRoute() (mini.PublicRoute, error) {
	if err := d.Normalize(); err != nil {
		return mini.PublicRoute{}, err
	}
	pr := mini.PublicRoute{
		Kind:      d.Kind,
		Service:   d.Service,
		Endpoint:  d.Endpoint,
		Method:    d.Method,
		Path:      d.Path,
		Subject:   d.Subject,
		Subscribe: d.Subscribe,
		Publish:   d.Publish,
		Timeout:   d.Timeout,
		Auth:      d.Auth,
	}
	if d.Kind == mini.TransportWS && pr.Subject == "" {
		pr.Subject = pr.Subscribe
	}
	return pr, nil
}

// Normalize ACL doc.
func (d *WSACLDoc) Normalize() error {
	if d == nil {
		return fmt.Errorf("nil wsacl")
	}
	if d.ID == "" {
		if d.Path != "" {
			sum := sha256.Sum256([]byte("acl:" + d.Path))
			d.ID = hex.EncodeToString(sum[:12])
		} else {
			sum := sha256.Sum256([]byte(fmt.Sprintf("%v%v", d.ClaimEquals, d.RequireAuth)))
			d.ID = hex.EncodeToString(sum[:12])
		}
	}
	if d.Enabled == nil {
		t := true
		d.Enabled = &t
	}
	return nil
}

func (d *WSACLDoc) IsEnabled() bool {
	return d != nil && (d.Enabled == nil || *d.Enabled)
}

// RouteKey KV key for a route id.
func RouteKey(id string) string {
	return PrefixRoutes + sanitizeKey(id)
}

// WSACLKey KV key for an ACL id.
func WSACLKey(id string) string {
	return PrefixWSACL + sanitizeKey(id)
}

func sanitizeKey(id string) string {
	id = strings.TrimSpace(id)
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	s := b.String()
	if s == "" {
		return "unknown"
	}
	return s
}

func marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func unmarshalRoute(data []byte) (RouteDoc, error) {
	var d RouteDoc
	err := json.Unmarshal(data, &d)
	return d, err
}

func unmarshalACL(data []byte) (WSACLDoc, error) {
	var d WSACLDoc
	err := json.Unmarshal(data, &d)
	return d, err
}
