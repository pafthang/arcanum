// Package config holds process-level settings for the embedded NATS server.
//
// cfgs/nats.json may seed the same env keys via ctrl / LoadConfig.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/pafthang/arcanum/pkg/svcutil"
)

// Config is env-driven process configuration for the embedded NATS server.
//
// When ConfigFile is set, ServerOptions loads that conf and only applies
// fields whose env vars were explicitly present (same as pre-layout behavior).
type Config struct {
	// Host is the client bind address (default 127.0.0.1 when no conf file).
	Host string
	// Port is the client port (default 4222 when no conf file).
	Port int
	// HTTPPort is the monitoring HTTP port (0 disables; default 8222 when no conf).
	HTTPPort int
	// ServerName is nats-server server_name (default micro-nats when no conf).
	ServerName string
	// JetStream enables JetStream (default true when no conf).
	JetStream bool
	// StoreDir is JetStream/file store root (default data/nats when no conf).
	StoreDir string
	// ConfigFile is an optional nats-server conf path.
	ConfigFile string
	// Debug enables server debug logs.
	Debug bool
	// Trace enables protocol trace logs.
	Trace bool

	// which env keys were set (for conf-file overlay semantics)
	setHost, setPort, setHTTPPort, setServerName, setStoreDir bool
	setJetStream, setDebug, setTrace                          bool
}

// Defaults returns local-dev defaults (used when NATS_CONFIG is unset).
func Defaults() Config {
	return Config{
		Host:       "127.0.0.1",
		Port:       4222,
		HTTPPort:   8222,
		ServerName: "micro-nats",
		JetStream:  true,
	}
}

// FromEnv loads Config from NATS_* environment variables.
//
//	NATS_HOST, NATS_PORT, NATS_HTTP_PORT
//	NATS_SERVER_NAME
//	NATS_JETSTREAM          — default true without conf; 0/false/off disables
//	NATS_STORE_DIR          — default data/nats without conf
//	NATS_CONFIG             — optional conf file
//	NATS_DEBUG, NATS_TRACE
func FromEnv() Config {
	c := Defaults()
	if v := strings.TrimSpace(os.Getenv("NATS_HOST")); v != "" {
		c.Host = v
		c.setHost = true
	}
	if v := strings.TrimSpace(os.Getenv("NATS_PORT")); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.Port = p
			c.setPort = true
		}
	}
	if v := strings.TrimSpace(os.Getenv("NATS_HTTP_PORT")); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			c.HTTPPort = p
			c.setHTTPPort = true
		}
	}
	if v := strings.TrimSpace(os.Getenv("NATS_SERVER_NAME")); v != "" {
		c.ServerName = v
		c.setServerName = true
	}
	if v := strings.TrimSpace(os.Getenv("NATS_JETSTREAM")); v != "" {
		c.JetStream = svcutil.EnvBool("NATS_JETSTREAM", true)
		c.setJetStream = true
	}
	if v := strings.TrimSpace(os.Getenv("NATS_STORE_DIR")); v != "" {
		c.StoreDir = v
		c.setStoreDir = true
	}
	c.ConfigFile = strings.TrimSpace(os.Getenv("NATS_CONFIG"))
	if v := strings.TrimSpace(os.Getenv("NATS_DEBUG")); v != "" {
		c.Debug = svcutil.EnvBool("NATS_DEBUG", false)
		c.setDebug = true
	}
	if v := strings.TrimSpace(os.Getenv("NATS_TRACE")); v != "" {
		c.Trace = svcutil.EnvBool("NATS_TRACE", false)
		c.setTrace = true
	}
	return c
}

// ServerOptions builds nats-server Options from Config (and optional conf file).
func (c Config) ServerOptions() (*natsserver.Options, error) {
	if c.ConfigFile != "" {
		opts, err := natsserver.ProcessConfigFile(c.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("config %s: %w", c.ConfigFile, err)
		}
		// Embedded process: we own signals and logging setup.
		opts.NoSigs = true
		opts.ConfigFile = c.ConfigFile
		// Env overlays only when the key was explicitly set.
		if c.setHost {
			opts.Host = c.Host
		}
		if c.setPort {
			opts.Port = c.Port
		}
		if c.setHTTPPort {
			opts.HTTPPort = c.HTTPPort
		}
		if c.setServerName {
			opts.ServerName = c.ServerName
		}
		if c.setStoreDir {
			opts.StoreDir = c.StoreDir
		}
		if c.setDebug {
			opts.Debug = c.Debug
		}
		if c.setTrace {
			opts.Trace = c.Trace
		}
		if c.setJetStream {
			opts.JetStream = c.JetStream
		}
		return opts, nil
	}

	storeDir := c.StoreDir
	if storeDir == "" {
		storeDir = svcutil.DataDir("nats")
	}
	if c.JetStream {
		if err := os.MkdirAll(storeDir, 0o755); err != nil {
			return nil, fmt.Errorf("store dir %s: %w", storeDir, err)
		}
	}

	opts := &natsserver.Options{
		ServerName: c.ServerName,
		Host:       c.Host,
		Port:       c.Port,
		HTTPHost:   c.Host,
		HTTPPort:   c.HTTPPort,
		JetStream:  c.JetStream,
		StoreDir:   storeDir,
		NoSigs:     true,
		Debug:      c.Debug,
		Trace:      c.Trace,
	}
	return opts, nil
}
