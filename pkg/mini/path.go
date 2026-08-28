package mini

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

// ParamHeaderPrefix is prepended to path parameter names when the gate
// forwards captured values into NATS headers (e.g. X-Mini-Param-id).
const ParamHeaderPrefix = "X-Mini-Param-"

// WildcardParam is the header/param name used for a single-segment `*` capture.
const WildcardParam = "*"

// PathParam returns a path parameter injected by the gate.
// For patterns with `*`, use name "*". For catch-all `{*rest}`, use "rest".
func PathParam(req Request, name string) string {
	if req == nil {
		return ""
	}
	return req.Headers().Get(ParamHeaderPrefix + name)
}

// segmentKind classifies a path segment.
type segmentKind int

const (
	segLiteral  segmentKind = iota
	segParam                // {name}
	segWildcard             // *
	segCatchAll             // {*name} — must be last
)

// pathSegment is one slash-separated part of an HTTP path pattern.
type pathSegment struct {
	kind  segmentKind
	value string // literal text, param name, or catch-all name
}

// PathPattern is a compiled HTTP path pattern.
//
// Supported segment forms:
//   - static:     "orders"
//   - param:      "{id}"
//   - wildcard:   "*"          (one segment, captured as param "*")
//   - catch-all:  "{*path}"    (remaining path, must be final segment)
//
// Examples:
//
//	/v1/orders/{id}
//	/v1/files/*
//	/v1/assets/{*path}
type PathPattern struct {
	Raw      string
	segments []pathSegment
	// ranking counts for specificity
	staticCount   int
	paramCount    int // {name} only
	wildcardCount int // *
	hasCatchAll   bool
}

// ParsePathPattern compiles a path pattern.
func ParsePathPattern(path string) (*PathPattern, error) {
	if path == "" || !strings.HasPrefix(path, "/") {
		return nil, fmt.Errorf("path must start with /")
	}
	if strings.Contains(path, " ") {
		return nil, fmt.Errorf("path must not contain spaces")
	}
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = strings.TrimSuffix(path, "/")
	}

	parts := strings.Split(path, "/")
	segs := make([]pathSegment, 0, len(parts)-1)
	seenParams := make(map[string]struct{})
	p := &PathPattern{Raw: path}

	for i, part := range parts[1:] {
		if part == "" {
			return nil, fmt.Errorf("empty path segment in %q", path)
		}
		// catch-all {*name}
		if strings.HasPrefix(part, "{*") && strings.HasSuffix(part, "}") {
			name := part[2 : len(part)-1]
			if name == "" {
				return nil, fmt.Errorf("empty catch-all parameter in %q", path)
			}
			if strings.ContainsAny(name, "/{}*") {
				return nil, fmt.Errorf("invalid catch-all parameter name %q", name)
			}
			if i != len(parts)-2 {
				return nil, fmt.Errorf("catch-all segment must be last in %q", path)
			}
			if _, ok := seenParams[name]; ok {
				return nil, fmt.Errorf("duplicate path parameter %q", name)
			}
			seenParams[name] = struct{}{}
			segs = append(segs, pathSegment{kind: segCatchAll, value: name})
			p.hasCatchAll = true
			continue
		}
		// {name}
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			name := part[1 : len(part)-1]
			if name == "" || strings.HasPrefix(name, "*") {
				return nil, fmt.Errorf("invalid path parameter %q", part)
			}
			if strings.ContainsAny(name, "/{}*") {
				return nil, fmt.Errorf("invalid path parameter name %q", name)
			}
			if _, ok := seenParams[name]; ok {
				return nil, fmt.Errorf("duplicate path parameter %q", name)
			}
			seenParams[name] = struct{}{}
			segs = append(segs, pathSegment{kind: segParam, value: name})
			p.paramCount++
			continue
		}
		// single-segment wildcard
		if part == "*" {
			segs = append(segs, pathSegment{kind: segWildcard, value: WildcardParam})
			p.wildcardCount++
			continue
		}
		if strings.ContainsAny(part, "{}*") {
			return nil, fmt.Errorf("invalid path segment %q", part)
		}
		segs = append(segs, pathSegment{kind: segLiteral, value: part})
		p.staticCount++
	}

	p.segments = segs
	return p, nil
}

// HasParams reports whether the pattern has params, wildcards, or catch-all.
func (p *PathPattern) HasParams() bool {
	if p == nil {
		return false
	}
	return p.paramCount > 0 || p.wildcardCount > 0 || p.hasCatchAll
}

// Match matches path against the pattern and returns captured params.
// Pure-static patterns return (nil, true) without allocating a params map
// or a segment slice.
func (p *PathPattern) Match(path string) (map[string]string, bool) {
	if p == nil {
		return nil, false
	}

	// Fast path: literal-only patterns — scan path in place (0 allocs).
	if !p.HasParams() {
		return nil, matchStaticPath(path, p.segments)
	}

	parts := splitPath(path)
	// Pre-size for common 1–2 param routes.
	params := make(map[string]string, p.paramCount+p.wildcardCount+boolToInt(p.hasCatchAll))
	si, pi := 0, 0
	for si < len(p.segments) {
		seg := p.segments[si]
		switch seg.kind {
		case segLiteral:
			if pi >= len(parts) || parts[pi] != seg.value {
				return nil, false
			}
			si++
			pi++
		case segParam:
			if pi >= len(parts) || parts[pi] == "" {
				return nil, false
			}
			params[seg.value] = parts[pi]
			si++
			pi++
		case segWildcard:
			if pi >= len(parts) || parts[pi] == "" {
				return nil, false
			}
			// last * wins if multiple (unusual)
			params[WildcardParam] = parts[pi]
			si++
			pi++
		case segCatchAll:
			if pi >= len(parts) {
				// require at least one segment
				return nil, false
			}
			params[seg.value] = strings.Join(parts[pi:], "/")
			si++
			pi = len(parts)
		}
	}
	if pi != len(parts) {
		return nil, false
	}
	return params, true
}

// matchStaticPath compares path to literal segments without allocating.
func matchStaticPath(path string, segs []pathSegment) bool {
	i := 0
	for _, seg := range segs {
		for i < len(path) && path[i] == '/' {
			i++
		}
		if i >= len(path) {
			return false
		}
		start := i
		for i < len(path) && path[i] != '/' {
			i++
		}
		if path[start:i] != seg.value {
			return false
		}
	}
	for i < len(path) {
		if path[i] != '/' {
			return false
		}
		i++
	}
	// Empty path only matches empty segment list.
	if len(segs) == 0 {
		return path == "" || path == "/"
	}
	return true
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Specificity ranks patterns: more static > fewer dynamic > no catch-all > longer.
// Returns values suitable for sorting (higher static is better).
func (p *PathPattern) Specificity() (static, dynamic, catchAll, length int) {
	if p == nil {
		return 0, 0, 0, 0
	}
	dynamic = p.paramCount + p.wildcardCount
	if p.hasCatchAll {
		catchAll = 1
	}
	return p.staticCount, dynamic, catchAll, len(p.Raw)
}

// ComparePathSpecificity returns true if a should be tried before b.
func ComparePathSpecificity(a, b *PathPattern) bool {
	as, ad, ac, al := a.Specificity()
	bs, bd, bc, bl := b.Specificity()
	if as != bs {
		return as > bs
	}
	if ad != bd {
		return ad < bd
	}
	if ac != bc {
		return ac < bc // no catch-all first
	}
	return al > bl
}

// ApplyPathParams sets X-Mini-Param-* headers on h.
// Uses nats.Header.Set so keys match Get() MIME canonicalization.
func ApplyPathParams(h Headers, params map[string]string) Headers {
	if len(params) == 0 {
		return h
	}
	if h == nil {
		h = Headers{}
	}
	nh := nats.Header(h)
	for k, v := range params {
		nh.Set(ParamHeaderPrefix+k, v)
	}
	return Headers(nh)
}
