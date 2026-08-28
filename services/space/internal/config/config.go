package config

import (
	"os"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/svcutil"
)

// DefaultHMACSecret matches gate's local default so login works without extra env.
const DefaultHMACSecret = "dev-secret-change-me"

// Config is the space service configuration.
type Config struct {
	JWTSecret    string
	JWTTTL       time.Duration
	SeedPassword string
	DefaultSpace string
}

// Defaults returns local-dev defaults.
func Defaults() Config {
	return Config{
		JWTSecret:    DefaultHMACSecret,
		JWTTTL:       24 * time.Hour,
		SeedPassword: "admin",
		DefaultSpace: "default",
	}
}

// FromEnv loads config from environment.
func FromEnv() Config {
	c := Defaults()
	if v := firstEnv("GATE_JWT_HMAC_SECRET", "SPACE_JWT_HMAC_SECRET", "SPACE_JWT_SECRET"); v != "" {
		c.JWTSecret = v
	}
	if v := os.Getenv("SPACE_JWT_TTL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			c.JWTTTL = d
		} else if n := svcutil.EnvInt("SPACE_JWT_TTL", 0); n > 0 {
			c.JWTTTL = time.Duration(n) * time.Second
		}
	}
	if v := strings.TrimSpace(os.Getenv("SPACE_SEED_PASSWORD")); v != "" {
		c.SeedPassword = v
	}
	return c
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}
