package nats

import "github.com/pafthang/arcanum/services/nats/internal/config"

// Re-export key types for consumers of the nats package.

type Config = config.Config

// FromEnv loads Config from NATS_* environment variables.
func FromEnv() Config {
	return config.FromEnv()
}
