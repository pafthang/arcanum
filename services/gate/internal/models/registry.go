package models

import (
	"sync"
)

// ModelInfo describes a model (AI/LLM gate use case).
type ModelInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Provider  string            `json:"provider"`
	Type      string            `json:"type"`
	MaxTokens int               `json:"max_tokens,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Enabled   bool              `json:"enabled"`
}

// Registry is a concurrency-safe model registry.
type Registry struct {
	mu     sync.RWMutex
	models map[string]*ModelInfo
}

// NewRegistry creates an empty registry.
func NewRegistry() *Registry {
	return &Registry{models: make(map[string]*ModelInfo)}
}

// Register inserts/updates a model.
func (r *Registry) Register(m *ModelInfo) {
	if m == nil || m.ID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	cp := *m
	if cp.Metadata != nil {
		meta := make(map[string]string, len(cp.Metadata))
		for k, v := range cp.Metadata {
			meta[k] = v
		}
		cp.Metadata = meta
	}
	r.models[m.ID] = &cp
}

// Get returns a model by ID.
func (r *Registry) Get(id string) (*ModelInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	m, ok := r.models[id]
	if !ok {
		return nil, false
	}
	cp := *m
	return &cp, true
}

// List returns all models.
func (r *Registry) List() []*ModelInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*ModelInfo, 0, len(r.models))
	for _, m := range r.models {
		cp := *m
		out = append(out, &cp)
	}
	return out
}

// Remove deletes a model.
func (r *Registry) Remove(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.models, id)
}

// Len returns the model count.
func (r *Registry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.models)
}
