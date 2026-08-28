package config

import "github.com/pafthang/arcanum/pkg/svcutil"

// Config is the media service configuration.
type Config struct {
	MaxBytes int
}

// FromEnv loads config from environment.
func FromEnv() Config {
	c := Config{MaxBytes: 1 << 20}
	c.MaxBytes = svcutil.EnvInt("MEDIA_MAX_BYTES", c.MaxBytes)
	if c.MaxBytes < 1024 {
		c.MaxBytes = 1024
	}
	return c
}
