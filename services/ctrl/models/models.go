// Package models holds shared DTO types for the ctrl control plane.
package models

// LifecycleBody is the JSON body for reload/restart APIs.
type LifecycleBody struct {
	Services []string `json:"services"`
	Reason   string   `json:"reason"`
	DelayMs  int      `json:"delay_ms"`
}

// ServiceStatus is one mini service instance from $SRV.STATS.
type ServiceStatus struct {
	Name     string            `json:"name"`
	ID       string            `json:"id"`
	Version  string            `json:"version,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`

	Started     string `json:"started,omitempty"`
	NumRequests int    `json:"num_requests"`
	NumErrors   int    `json:"num_errors"`
}

// CfgRow is a sanitized view of cfgs/<name>.json for inventory APIs.
type CfgRow struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Enabled     bool     `json:"enabled"`
	Order       int      `json:"order"`
	EnvKeys     []string `json:"env_keys,omitempty"`
	Command     []string `json:"command,omitempty"`
}
