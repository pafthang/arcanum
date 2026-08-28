package routing

import (
	"sync"
	"time"
)

// routeCacheEntry is a cached route resolve result.
type routeCacheEntry struct {
	route    *Route
	expireAt time.Time
}

// RouteCache is a simple TTL cache for routes.
type RouteCache struct {
	mu      sync.RWMutex
	entries map[string]*routeCacheEntry
	ttl     time.Duration
}

// NewRouteCache creates a cache with the given TTL.
func NewRouteCache(ttl time.Duration) *RouteCache {
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	c := &RouteCache{
		entries: make(map[string]*routeCacheEntry),
		ttl:     ttl,
	}
	go c.cleanupLoop()
	return c
}

// Get returns a route from the cache.
func (c *RouteCache) Get(key string) (*Route, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expireAt) {
		return nil, false
	}
	return e.route, true
}

// Set stores a route in the cache.
func (c *RouteCache) Set(key string, route *Route) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = &routeCacheEntry{
		route:    route,
		expireAt: time.Now().Add(c.ttl),
	}
}

// Invalidate removes a key.
func (c *RouteCache) Invalidate(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear wipes the cache.
func (c *RouteCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*routeCacheEntry)
}

func (c *RouteCache) cleanupLoop() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, e := range c.entries {
			if now.After(e.expireAt) {
				delete(c.entries, k)
			}
		}
		c.mu.Unlock()
	}
}
