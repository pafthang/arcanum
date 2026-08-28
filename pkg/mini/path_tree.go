package mini

import (
	"strings"
	"sync"
)

// PathTree is a segment radix trie for HTTP path patterns.
// Match priority at each segment: literal → {param} → * → {*catch-all}.
//
// Exact static paths can also be stored here; gateways may still keep a
// separate exact map for O(1) static lookups.
type PathTree[T any] struct {
	root *pathNode[T]
	n    int
}

type pathNode[T any] struct {
	children  map[string]*pathNode[T] // literal segments
	param     *pathNode[T]            // {name}
	paramName string
	wildcard  *pathNode[T] // single-segment *
	catchAll  *pathNode[T] // {*name}
	catchName string

	// leaves are terminal values at this node (path fully consumed).
	// Usually one; multiple allowed and ordered by pattern specificity.
	leaves []pathLeaf[T]
}

type pathLeaf[T any] struct {
	pattern *PathPattern
	value   T
}

// NewPathTree creates an empty path trie.
func NewPathTree[T any]() *PathTree[T] {
	return &PathTree[T]{root: &pathNode[T]{}}
}

// Len returns the number of inserted patterns.
func (t *PathTree[T]) Len() int {
	if t == nil {
		return 0
	}
	return t.n
}

// Add inserts pattern → value. pattern must be non-nil (from ParsePathPattern).
// Replaces an existing leaf with the same Raw pattern.
func (t *PathTree[T]) Add(pattern *PathPattern, value T) {
	if t == nil || pattern == nil {
		return
	}
	if t.root == nil {
		t.root = &pathNode[T]{}
	}
	n := t.root
	for _, seg := range pattern.segments {
		switch seg.kind {
		case segLiteral:
			if n.children == nil {
				n.children = make(map[string]*pathNode[T])
			}
			ch := n.children[seg.value]
			if ch == nil {
				ch = &pathNode[T]{}
				n.children[seg.value] = ch
			}
			n = ch
		case segParam:
			if n.param == nil {
				n.param = &pathNode[T]{}
				n.paramName = seg.value
			}
			// keep first param name if re-used edge (patterns share structure)
			n = n.param
		case segWildcard:
			if n.wildcard == nil {
				n.wildcard = &pathNode[T]{}
			}
			n = n.wildcard
		case segCatchAll:
			if n.catchAll == nil {
				n.catchAll = &pathNode[T]{}
				n.catchName = seg.value
			}
			n = n.catchAll
		}
	}
	// Upsert leaf by Raw pattern.
	for i, leaf := range n.leaves {
		if leaf.pattern != nil && leaf.pattern.Raw == pattern.Raw {
			n.leaves[i].value = value
			n.leaves[i].pattern = pattern
			return
		}
	}
	n.leaves = append(n.leaves, pathLeaf[T]{pattern: pattern, value: value})
	// Keep most-specific first for deterministic multi-leaf nodes.
	if len(n.leaves) > 1 {
		sortLeaves(n.leaves)
	}
	t.n++
}

func sortLeaves[T any](leaves []pathLeaf[T]) {
	// insertion sort — leaves are almost always size 1
	for i := 1; i < len(leaves); i++ {
		j := i
		for j > 0 && ComparePathSpecificity(leaves[j].pattern, leaves[j-1].pattern) {
			leaves[j], leaves[j-1] = leaves[j-1], leaves[j]
			j--
		}
	}
}

// pathPartsPool recycles segment slices used by PathTree.Match.
var pathPartsPool = sync.Pool{
	New: func() any {
		b := make([]string, 0, 16)
		return &b
	},
}

// Match finds the best route for path. Returns value, params, ok.
// Static (literal-only) paths avoid allocating a params map.
func (t *PathTree[T]) Match(path string) (value T, params map[string]string, ok bool) {
	var zero T
	if t == nil || t.root == nil {
		return zero, nil, false
	}
	bufp := pathPartsPool.Get().(*[]string)
	parts := splitPathInto(path, (*bufp)[:0])
	v, p, found := matchNode(t.root, parts, 0, nil)
	// Return buffer only if not oversized (avoid retaining huge caps).
	if cap(parts) <= 64 {
		*bufp = parts[:0]
		pathPartsPool.Put(bufp)
	}
	if found {
		return v, p, true
	}
	return zero, nil, false
}

func matchNode[T any](n *pathNode[T], parts []string, i int, params map[string]string) (T, map[string]string, bool) {
	var zero T
	if n == nil {
		return zero, nil, false
	}

	// Path fully consumed → terminal leaf.
	// Ownership of params transfers to the caller (no clone).
	if i == len(parts) {
		if len(n.leaves) == 0 {
			return zero, nil, false
		}
		leaf := n.leaves[0]
		// Shared param edges keep only the first registered name (e.g. teamId vs id).
		// Re-extract captures from the winning leaf's pattern so param keys match the route.
		if leaf.pattern != nil && leaf.pattern.HasParams() {
			path := "/" + strings.Join(parts, "/")
			if pm, ok := leaf.pattern.Match(path); ok {
				return leaf.value, emptyParamsNil(pm), true
			}
		}
		return leaf.value, emptyParamsNil(params), true
	}

	seg := parts[i]

	// 1) literal — most specific
	if n.children != nil {
		if ch := n.children[seg]; ch != nil {
			if v, p, ok := matchNode(ch, parts, i+1, params); ok {
				return v, p, true
			}
		}
	}

	// 2) named param
	if n.param != nil && seg != "" {
		name := n.paramName
		params = ensureParams(params)
		prev, had := params[name]
		params[name] = seg
		if v, p, ok := matchNode(n.param, parts, i+1, params); ok {
			return v, p, true
		}
		if had {
			params[name] = prev
		} else {
			delete(params, name)
		}
	}

	// 3) single-segment wildcard
	if n.wildcard != nil && seg != "" {
		params = ensureParams(params)
		prev, had := params[WildcardParam]
		params[WildcardParam] = seg
		if v, p, ok := matchNode(n.wildcard, parts, i+1, params); ok {
			return v, p, true
		}
		if had {
			params[WildcardParam] = prev
		} else {
			delete(params, WildcardParam)
		}
	}

	// 4) catch-all consumes remaining segments
	if n.catchAll != nil && i < len(parts) {
		name := n.catchName
		rest := strings.Join(parts[i:], "/")
		if rest == "" {
			return zero, nil, false
		}
		params = ensureParams(params)
		prev, had := params[name]
		params[name] = rest
		if len(n.catchAll.leaves) > 0 {
			return n.catchAll.leaves[0].value, emptyParamsNil(params), true
		}
		if had {
			params[name] = prev
		} else {
			delete(params, name)
		}
	}

	return zero, nil, false
}

func ensureParams(params map[string]string) map[string]string {
	if params == nil {
		return make(map[string]string, 2)
	}
	return params
}

// emptyParamsNil normalizes an empty map to nil (stable API for static matches).
func emptyParamsNil(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	return in
}

// splitPath splits an HTTP path into non-empty segments.
// Leading/trailing slashes and "//" are ignored. Allocates once (count then fill).
func splitPath(path string) []string {
	n := countPathSegments(path)
	if n == 0 {
		return nil
	}
	return splitPathInto(path, make([]string, 0, n))
}

// splitPathInto is like splitPath but reuses buf's capacity when possible.
// Prefer a pre-sized or pooled buf (PathTree.Match) to avoid growth allocs.
func splitPathInto(path string, buf []string) []string {
	if path == "" || path == "/" {
		return buf[:0]
	}
	n := countPathSegments(path)
	if n == 0 {
		return buf[:0]
	}
	out := buf[:0]
	if cap(out) < n {
		out = make([]string, 0, n)
	}
	i := 0
	for i < len(path) {
		for i < len(path) && path[i] == '/' {
			i++
		}
		if i >= len(path) {
			break
		}
		start := i
		for i < len(path) && path[i] != '/' {
			i++
		}
		out = append(out, path[start:i])
	}
	return out
}

func countPathSegments(path string) int {
	if path == "" || path == "/" {
		return 0
	}
	n, i := 0, 0
	for i < len(path) {
		for i < len(path) && path[i] == '/' {
			i++
		}
		if i >= len(path) {
			break
		}
		n++
		for i < len(path) && path[i] != '/' {
			i++
		}
	}
	return n
}

// All walks all leaves in unspecified order.
func (t *PathTree[T]) All() []T {
	if t == nil || t.root == nil {
		return nil
	}
	var out []T
	walkNode(t.root, func(leaf pathLeaf[T]) {
		out = append(out, leaf.value)
	})
	return out
}

func walkNode[T any](n *pathNode[T], fn func(pathLeaf[T])) {
	if n == nil {
		return
	}
	for _, leaf := range n.leaves {
		fn(leaf)
	}
	for _, ch := range n.children {
		walkNode(ch, fn)
	}
	walkNode(n.param, fn)
	walkNode(n.wildcard, fn)
	walkNode(n.catchAll, fn)
}
