package routing

import (
	"net/http"
	"strings"
	"sync"
)

// Route describes a route.
type Route struct {
	ID          string            `json:"id"`
	Path        string            `json:"path"`         // /api/v1/* or exact path
	Methods     []string          `json:"methods"`      // empty = any
	Upstream    string            `json:"upstream"`     // http://backend:8080
	StripPrefix string            `json:"strip_prefix"` // prefix stripped before proxying
	Headers     map[string]string `json:"headers"`      // extra headers
	Timeout     int               `json:"timeout_ms"`   // 0 = default
	Priority    int               `json:"priority"`     // higher = earlier
	Enabled     bool              `json:"enabled"`
}

// Table is a concurrency-safe route table.
type Table struct {
	mu     sync.RWMutex
	routes []*Route
}

// NewTable creates an empty table.
func NewTable() *Table {
	return &Table{routes: make([]*Route, 0)}
}

// Add inserts/updates a route by ID.
func (t *Table) Add(r *Route) {
	if r == nil || r.ID == "" {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, existing := range t.routes {
		if existing.ID == r.ID {
			t.routes[i] = r
			return
		}
	}
	t.routes = append(t.routes, r)
}

// Remove deletes a route by ID.
func (t *Table) Remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, r := range t.routes {
		if r.ID == id {
			t.routes = append(t.routes[:i], t.routes[i+1:]...)
			return
		}
	}
}

// List returns a copy of all routes.
func (t *Table) List() []*Route {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]*Route, len(t.routes))
	copy(out, t.routes)
	return out
}

// Match finds a route for the request.
func (t *Table) Match(r *http.Request) *Route {
	t.mu.RLock()
	defer t.mu.RUnlock()

	path := r.URL.Path
	method := r.Method

	var best *Route
	bestPrio := -1

	for _, route := range t.routes {
		if !route.Enabled {
			continue
		}
		if !matchMethod(route.Methods, method) {
			continue
		}
		if !matchPath(route.Path, path) {
			continue
		}
		if route.Priority > bestPrio {
			best = route
			bestPrio = route.Priority
		}
	}
	return best
}

func matchMethod(allowed []string, method string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, m := range allowed {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

func matchPath(pattern, path string) bool {
	if pattern == path {
		return true
	}
	// Simple wildcard: /api/*
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return path == prefix || strings.HasPrefix(path, prefix+"/")
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return false
}
