package gate

import (
	"github.com/pafthang/arcanum/services/gate/internal/config"
	"github.com/pafthang/arcanum/services/gate/internal/core"
)

// Re-export key types for consumers of the gate package.

type (
	Gate   = core.Gate
	Config = config.Config
)

// New creates a core Gate (convenience alias for core.New).
// NATS is connected lazily in Gate.Start.
func New(cfg *Config) (*Gate, error) {
	return core.New(cfg, nil)
}

// DefaultConfig returns configuration defaults (no env applied).
func DefaultConfig() *Config {
	return config.DefaultConfig()
}

// FromEnv applies environment overrides onto cfg.
func FromEnv(cfg *Config) {
	config.FromEnv(cfg)
}

// LoadConfig returns defaults with environment overrides applied.
func LoadConfig() *Config {
	cfg := config.DefaultConfig()
	config.FromEnv(cfg)
	return cfg
}
