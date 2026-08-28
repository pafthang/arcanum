// Package config is process env for the ctrl control plane (not edge RouteDoc config).
package config

import (
	"github.com/pafthang/arcanum/pkg/svcutil"

	"os"
	"strings"
	"time"

	"github.com/pafthang/arcanum/services/ctrl/internal/edgecfg"
)

// Config is env-driven process configuration.
type Config struct {
	// Up enables process supervision (cfgs/* children). Flag -up also sets this.
	Up bool
	// SkipNATS skips supervising the nats process when Up is true.
	SkipNATS bool
	// NATSURL is the broker URL.
	NATSURL string
	// NATSConnectTimeout when waiting for supervised nats.
	NATSConnectTimeout time.Duration
	// SuperviseStartupDelay after launching supervisor before connecting.
	SuperviseStartupDelay time.Duration

	// Edge config plane (JetStream KV) — shared with gate/admin.
	ConfigBucket string
	// AdminEnable starts standalone admin HTTP when set (also used by cmd/config serve).
	AdminEnable bool
	AdminAddr   string
	AdminPrefix string
	AdminTokens []string
}

// Defaults returns production-leaning defaults.
func Defaults() Config {
	return Config{
		Up:                    false,
		SkipNATS:              false,
		NATSURL:               "nats://127.0.0.1:4222",
		NATSConnectTimeout:    60 * time.Second,
		SuperviseStartupDelay: 300 * time.Millisecond,
		ConfigBucket:          edgecfg.DefaultBucket,
		AdminEnable:           false,
		AdminAddr:             ":8090",
		AdminPrefix:           "/_admin",
	}
}

// FromEnv loads Config from process environment.
//
//	CTRL_UP / -up flag (caller)
//	CTRL_SKIP_NATS / SKIP_NATS
//	NATS_URL
//	CONFIG_BUCKET
//	ADMIN_ENABLE, ADMIN_ADDR, ADMIN_PREFIX, ADMIN_TOKENS (comma-separated)
func FromEnv() Config {
	c := Defaults()
	c.Up = svcutil.EnvBool("CTRL_UP", false)
	c.SkipNATS = svcutil.EnvBool("CTRL_SKIP_NATS", false) || svcutil.EnvBool("SKIP_NATS", false)
	if v := strings.TrimSpace(os.Getenv("NATS_URL")); v != "" {
		c.NATSURL = v
	}
	if v := strings.TrimSpace(os.Getenv("CONFIG_BUCKET")); v != "" {
		c.ConfigBucket = v
	}
	c.AdminEnable = svcutil.EnvBool("ADMIN_ENABLE", false) || svcutil.EnvBool("CONFIG_ADMIN_ENABLE", false)
	if v := strings.TrimSpace(os.Getenv("ADMIN_ADDR")); v != "" {
		c.AdminAddr = v
	}
	if v := strings.TrimSpace(os.Getenv("ADMIN_PREFIX")); v != "" {
		c.AdminPrefix = v
	}
	if v := strings.TrimSpace(os.Getenv("ADMIN_TOKENS")); v != "" {
		for _, t := range strings.Split(v, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				c.AdminTokens = append(c.AdminTokens, t)
			}
		}
	}
	if n := svcutil.EnvInt("CTRL_NATS_CONNECT_TIMEOUT_SEC", 0); n > 0 {
		c.NATSConnectTimeout = time.Duration(n) * time.Second
	}
	return c
}
