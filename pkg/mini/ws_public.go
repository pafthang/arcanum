package mini

import (
	"fmt"
	"strings"
)

// Real-time transport is WebSocket-only (no SSE).
const (
	// MetaTransport selects exposure kind: "http" (default) or "ws".
	MetaTransport = "mini.transport"

	// MetaWSPath is the public WebSocket path (e.g. /v1/orders/ws).
	MetaWSPath = "mini.ws.path"

	// MetaWSSubscribe is the NATS subject the gate subscribes to for
	// server→client frames. Supports {param} placeholders from the path.
	MetaWSSubscribe = "mini.ws.subscribe"

	// MetaWSPublish is the NATS subject the gate publishes client→server
	// frames to. Optional; if empty, the socket is read-only (server push).
	// Supports {param} placeholders from the path.
	MetaWSPublish = "mini.ws.publish"

	TransportHTTP = "http"
	TransportWS   = "ws"

	// WSMethod is the synthetic "method" used in routing tables for WebSockets.
	WSMethod = "WS"
)

// WSConfig configures a public WebSocket route advertised via service INFO.
type WSConfig struct {
	// Path is the public WS path pattern (must start with /), e.g. /v1/orders/{id}/ws.
	Path string

	// Subscribe is the NATS subject for downstream (NATS → WS).
	// Example: public.orders.events.{id} or public.orders.events.>
	Subscribe string

	// Publish is the NATS subject for upstream (WS → NATS). Empty = read-only.
	// Example: public.orders.ws.{id}.in
	Publish string
}

// WithPublicWS marks an endpoint as a public WebSocket bridge endpoint.
// The endpoint still appears in $SRV.INFO (for discovery); its NATS handler is
// typically unused for RPC — real-time traffic goes through Subscribe/Publish.
// Prefer Service.AddPublicWS when you only need to advertise a WS route.
func WithPublicWS(cfg WSConfig) EndpointOpt {
	return func(e *endpointOpts) error {
		cfg.Path = strings.TrimSpace(cfg.Path)
		cfg.Subscribe = strings.TrimSpace(cfg.Subscribe)
		cfg.Publish = strings.TrimSpace(cfg.Publish)
		if err := validatePublicWS(cfg); err != nil {
			return fmt.Errorf("%w: %s", ErrConfigValidation, err)
		}
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[MetaPublic] = "true"
		e.metadata[MetaTransport] = TransportWS
		e.metadata[MetaWSPath] = cfg.Path
		e.metadata[MetaWSSubscribe] = cfg.Subscribe
		if cfg.Publish != "" {
			e.metadata[MetaWSPublish] = cfg.Publish
		}
		return nil
	}
}

// AddPublicWS registers a discoverable WebSocket route without business RPC logic.
// Optional EndpointOpt can set subject (WithPublicSubject / WithEndpointSubject).
// If no subject is set, a private advertisement subject under _mini.ws.ad.<name> is used.
func (s *service) AddPublicWS(name string, cfg WSConfig, opts ...EndpointOpt) error {
	opts = append([]EndpointOpt{WithPublicWS(cfg)}, opts...)
	// Ensure a subject exists for the INFO endpoint entry.
	hasSubject := false
	var probe endpointOpts
	for _, opt := range opts {
		if err := opt(&probe); err != nil {
			return err
		}
	}
	if probe.subject != "" {
		hasSubject = true
	}
	if !hasSubject {
		opts = append(opts, WithEndpointSubject("_mini.ws.ad."+name))
	}
	handler := HandlerFunc(func(req Request) {
		_ = req.Error("400", "websocket endpoint: use WS upgrade on the public path", nil)
	})
	return s.AddEndpoint(name, handler, opts...)
}

func validatePublicWS(cfg WSConfig) error {
	if _, err := ParsePathPattern(cfg.Path); err != nil {
		return fmt.Errorf("ws path: %w", err)
	}
	if cfg.Subscribe == "" {
		return fmt.Errorf("ws subscribe subject is required")
	}
	if strings.Contains(cfg.Subscribe, " ") {
		return fmt.Errorf("invalid subscribe subject")
	}
	if cfg.Publish != "" && strings.Contains(cfg.Publish, " ") {
		return fmt.Errorf("invalid publish subject")
	}
	return nil
}

// PublicWSRoutesFromInfo extracts WebSocket routes from a service INFO response.
func PublicWSRoutesFromInfo(info Info) []PublicRoute {
	var routes []PublicRoute
	for _, ep := range info.Endpoints {
		r, ok := publicWSRouteFromEndpoint(info, ep)
		if ok {
			routes = append(routes, r)
		}
	}
	return routes
}

func publicWSRouteFromEndpoint(info Info, ep EndpointInfo) (PublicRoute, bool) {
	if ep.Metadata == nil || !isTruthy(ep.Metadata[MetaPublic]) {
		return PublicRoute{}, false
	}
	// Explicit transport=ws, or presence of mini.ws.path
	transport := strings.ToLower(strings.TrimSpace(ep.Metadata[MetaTransport]))
	path := strings.TrimSpace(ep.Metadata[MetaWSPath])
	if transport != TransportWS && path == "" {
		return PublicRoute{}, false
	}
	if path == "" {
		return PublicRoute{}, false
	}
	sub := strings.TrimSpace(ep.Metadata[MetaWSSubscribe])
	pub := strings.TrimSpace(ep.Metadata[MetaWSPublish])
	if err := validatePublicWS(WSConfig{Path: path, Subscribe: sub, Publish: pub}); err != nil {
		return PublicRoute{}, false
	}
	pat, err := ParsePathPattern(path)
	if err != nil {
		return PublicRoute{}, false
	}
	return PublicRoute{
		Service:   info.Name,
		ServiceID: info.ID,
		Version:   info.Version,
		Endpoint:  ep.Name,
		Subject:   sub, // primary subject = subscribe for WS
		Method:    WSMethod,
		Path:      pat.Raw,
		Auth:      authModeFromMetadata(ep.Metadata),
		Metadata:  ep.Metadata,
		pattern:   pat,
		Kind:      TransportWS,
		Subscribe: sub,
		Publish:   pub,
	}, true
}

// ExpandSubject replaces {param} placeholders in a NATS subject template
// using path parameters captured from the HTTP/WS path.
func ExpandSubject(template string, params map[string]string) string {
	if template == "" || len(params) == 0 {
		return template
	}
	out := template
	for k, v := range params {
		out = strings.ReplaceAll(out, "{"+k+"}", v)
	}
	return out
}
