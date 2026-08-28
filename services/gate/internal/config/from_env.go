package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// FromEnv loads/overrides config from environment variables.
// Supports standard variables:
//
//	GATE_LISTEN_ADDR, GATE_READ_TIMEOUT, AUTH_METHODS,
//	RATE_LIMIT_ENABLED, RATE_LIMIT_RPS, etc.
func FromEnv(cfg *Config) {
	if cfg == nil {
		return
	}

	if v := os.Getenv("GATE_LISTEN_ADDR"); v != "" {
		cfg.Server.ListenAddr = v
	}
	if v := os.Getenv("GATE_READ_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ReadTimeout = d
		}
	}
	if v := os.Getenv("GATE_WRITE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.WriteTimeout = d
		}
	}
	if v := os.Getenv("GATE_IDLE_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.IdleTimeout = d
		}
	}
	if v := os.Getenv("GATE_SHUTDOWN_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			cfg.Server.ShutdownTimeout = d
		}
	}

	// Auth
	if v := os.Getenv("AUTH_METHODS"); v != "" {
		cfg.Auth.Methods = splitAndTrim(v)
	}
	if v := os.Getenv("AUTH_OPTIONAL"); v != "" {
		cfg.Auth.Optional = parseBool(v)
	}
	// Shared HS256 secret (must match auth service)
	if v := firstEnv("GATE_JWT_HMAC_SECRET", "SPACE_JWT_SECRET"); v != "" {
		cfg.Auth.HMACSecret = v
	}
	if v := firstEnv("GATE_JWT_ISSUER", "SPACE_JWT_ISSUER", "JWT_ISSUER"); v != "" {
		cfg.Auth.HMACIssuer = v
		cfg.Auth.JWT.Issuer = v
	}
	if v := os.Getenv("JWT_AUDIENCE"); v != "" {
		cfg.Auth.JWT.Audience = v
	}
	if v := os.Getenv("JWT_JWKS_URL"); v != "" {
		cfg.Auth.JWT.JWKSURL = v
	}
	if v := os.Getenv("CLAIM_HEADERS"); v != "" {
		cfg.ClaimHeaders = v
	}

	// NATS / discovery
	if v := os.Getenv("NATS_URL"); v != "" {
		cfg.NATSURL = v
	}
	if v := os.Getenv("GATE_DISCOVER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.DiscoverInterval = d
		}
	}
	if v := os.Getenv("GATE_DISCOVER_WAIT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.DiscoverWait = d
		}
	}
	if v := os.Getenv("GATE_REQUEST_TIMEOUT"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.RequestTimeout = d
		}
	}
	if v := os.Getenv("GATE_MAX_BODY"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			cfg.MaxBodyBytes = n
		}
	}
	// HTTP_ADDR used by cfgs/gate.json
	if v := os.Getenv("HTTP_ADDR"); v != "" {
		cfg.Server.ListenAddr = v
	}

	// Rate limit (RPS<=0 means disable — common "unlimited" local default)
	if v := os.Getenv("RATE_LIMIT_ENABLED"); v != "" {
		cfg.RateLimit.Enabled = parseBool(v)
	}
	if v := os.Getenv("RATE_LIMIT_RPS"); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			if f <= 0 {
				cfg.RateLimit.Enabled = false
			} else {
				cfg.RateLimit.RequestsPerSec = f
			}
		}
	}
	if v := os.Getenv("RATE_LIMIT_BURST"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			cfg.RateLimit.Burst = i
		}
	}

	// CORS
	if v := os.Getenv("CORS_ALLOWED_ORIGINS"); v != "" {
		cfg.CORS.AllowedOrigins = splitAndTrim(v)
	}
	if v := os.Getenv("CORS_ALLOW_CREDENTIALS"); v != "" {
		cfg.CORS.AllowCredentials = parseBool(v)
	}

	// Admin
	if v := os.Getenv("ADMIN_ENABLED"); v != "" {
		cfg.Admin.Enabled = parseBool(v)
	}
	if v := os.Getenv("ADMIN_LISTEN_ADDR"); v != "" {
		cfg.Admin.ListenAddr = v
	}

	// Metrics
	if v := os.Getenv("METRICS_ENABLED"); v != "" {
		cfg.Metrics.Enabled = parseBool(v)
	}
	if v := os.Getenv("METRICS_PATH"); v != "" {
		cfg.Metrics.Path = v
	}

	// Policy
	if v := os.Getenv("POLICY_ENGINE"); v != "" {
		cfg.Policy.Engine = v
	}
	if v := os.Getenv("POLICY_PATH"); v != "" {
		cfg.Policy.PoliciesPath = v
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes" || s == "on"
}

func firstEnv(keys ...string) string {
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return ""
}
