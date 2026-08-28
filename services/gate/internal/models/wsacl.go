package models

import (
	"sync"
)

// WSACLEntry is an ACL entry for WebSocket connections.
type WSACLEntry struct {
	Subject         string            `json:"subject"`
	AllowedChannels []string          `json:"allowed_channels"`
	MaxConnections  int               `json:"max_connections"`
	Metadata        map[string]string `json:"metadata,omitempty"`
}

// WSACL is a concurrency-safe WebSocket ACL.
type WSACL struct {
	mu      sync.RWMutex
	entries map[string]*WSACLEntry
}

// NewWSACL creates an empty ACL.
func NewWSACL() *WSACL {
	return &WSACL{entries: make(map[string]*WSACLEntry)}
}

// Set inserts/updates an entry.
func (a *WSACL) Set(entry *WSACLEntry) {
	if entry == nil || entry.Subject == "" {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	cp := *entry
	if cp.AllowedChannels != nil {
		ch := make([]string, len(cp.AllowedChannels))
		copy(ch, cp.AllowedChannels)
		cp.AllowedChannels = ch
	}
	if cp.Metadata != nil {
		meta := make(map[string]string, len(cp.Metadata))
		for k, v := range cp.Metadata {
			meta[k] = v
		}
		cp.Metadata = meta
	}
	a.entries[entry.Subject] = &cp
}

// Get returns the entry for subject.
func (a *WSACL) Get(subject string) (*WSACLEntry, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.entries[subject]
	if !ok {
		return nil, false
	}
	cp := *e
	return &cp, true
}

// Allow reports whether subject may access conn.
func (a *WSACL) Allow(subject, conn string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.entries[subject]
	if !ok {
		return false
	}
	for _, ch := range e.AllowedChannels {
		if ch == "*" || ch == conn {
			return true
		}
	}
	return false
}

// Remove deletes an entry.
func (a *WSACL) Remove(subject string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.entries, subject)
}

// List returns all entries.
func (a *WSACL) List() []*WSACLEntry {
	a.mu.RLock()
	defer a.mu.RUnlock()
	out := make([]*WSACLEntry, 0, len(a.entries))
	for _, e := range a.entries {
		cp := *e
		out = append(out, &cp)
	}
	return out
}
