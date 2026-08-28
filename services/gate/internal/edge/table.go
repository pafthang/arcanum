package edge

import (
	"strings"
	"sync"

	"github.com/pafthang/arcanum/pkg/mini"
)

// Table is a live set of discovered public HTTP + WebSocket routes.
type Table struct {
	mu     sync.RWMutex
	routes []mini.PublicRoute
	// method → path tree of route indexes into routes slice (HTTP only)
	byMethod map[string]*mini.PathTree[int]
	// WebSocket path tree → indexes into routes
	wsTree *mini.PathTree[int]
}

// NewTable creates an empty route table.
func NewTable() *Table {
	return &Table{byMethod: make(map[string]*mini.PathTree[int])}
}

// Replace atomically swaps the full route set (HTTP + WS).
func (t *Table) Replace(routes []mini.PublicRoute) {
	byMethod := make(map[string]*mini.PathTree[int])
	wsTree := mini.NewPathTree[int]()
	stored := make([]mini.PublicRoute, 0, len(routes))
	for _, r := range routes {
		pat := r.Pattern()
		if pat == nil {
			continue
		}
		idx := len(stored)
		stored = append(stored, r)

		if strings.EqualFold(r.Kind, mini.TransportWS) || strings.EqualFold(r.Method, mini.WSMethod) {
			wsTree.Add(pat, idx)
			continue
		}
		method := strings.ToUpper(strings.TrimSpace(r.Method))
		if method == "" {
			continue
		}
		tree := byMethod[method]
		if tree == nil {
			tree = mini.NewPathTree[int]()
			byMethod[method] = tree
		}
		tree.Add(pat, idx)
	}
	t.mu.Lock()
	t.routes = stored
	t.byMethod = byMethod
	t.wsTree = wsTree
	t.mu.Unlock()
}

// List returns a copy of current routes (HTTP + WS) for catalog.
func (t *Table) List() []mini.PublicRoute {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]mini.PublicRoute, len(t.routes))
	copy(out, t.routes)
	return out
}

// Match finds a route for method + path and returns path params (HTTP only).
func (t *Table) Match(method, path string) (*mini.PublicRoute, map[string]string, bool) {
	method = strings.ToUpper(strings.TrimSpace(method))
	t.mu.RLock()
	defer t.mu.RUnlock()
	tree := t.byMethod[method]
	if tree == nil {
		return nil, nil, false
	}
	idx, params, ok := tree.Match(path)
	if !ok || idx < 0 || idx >= len(t.routes) {
		return nil, nil, false
	}
	r := t.routes[idx]
	return &r, params, true
}

// MatchWS finds a WebSocket route for the given path.
func (t *Table) MatchWS(path string) (*mini.PublicRoute, map[string]string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.wsTree == nil {
		return nil, nil, false
	}
	idx, params, ok := t.wsTree.Match(path)
	if !ok || idx < 0 || idx >= len(t.routes) {
		return nil, nil, false
	}
	r := t.routes[idx]
	return &r, params, true
}

// Len returns total route count (HTTP + WS).
func (t *Table) Len() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.routes)
}

// HTTPLen returns HTTP-only route count.
func (t *Table) HTTPLen() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	n := 0
	for _, r := range t.routes {
		if !strings.EqualFold(r.Kind, mini.TransportWS) && !strings.EqualFold(r.Method, mini.WSMethod) {
			n++
		}
	}
	return n
}
