package config

import (
	"fmt"
	"time"
)

// Config is the root gate configuration.
type Config struct {
	Server    ServerConfig    `yaml:"server" json:"server"`
	Auth      AuthConfig      `yaml:"auth" json:"auth"`
	TLS       TLSConfig       `yaml:"tls" json:"tls"`
	CORS      CORSConfig      `yaml:"cors" json:"cors"`
	RateLimit RateLimitConfig `yaml:"rate_limit" json:"rate_limit"`
	Policy    PolicyConfig    `yaml:"policy" json:"policy"`
	Admin     AdminConfig     `yaml:"admin" json:"admin"`
	Metrics   MetricsConfig   `yaml:"metrics" json:"metrics"`

	// NATS edge
	NATSURL          string        `yaml:"nats_url" json:"nats_url"`
	DiscoverInterval time.Duration `yaml:"discover_interval" json:"discover_interval"`
	DiscoverWait     time.Duration `yaml:"discover_wait" json:"discover_wait"`
	RequestTimeout   time.Duration `yaml:"request_timeout" json:"request_timeout"`
	MaxBodyBytes     int64         `yaml:"max_body_bytes" json:"max_body_bytes"`
	// ClaimHeaders is "claim:Header,..." mapping JWT claims → NATS/HTTP headers.
	ClaimHeaders string `yaml:"claim_headers" json:"claim_headers"`
}

// ServerConfig configures the HTTP server.
type ServerConfig struct {
	ListenAddr      string        `yaml:"listen_addr" json:"listen_addr"`
	ReadTimeout     time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout" json:"write_timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" json:"shutdown_timeout"`
	EnableHTTP2     bool          `yaml:"enable_http2" json:"enable_http2"`
}

// AuthConfig configures authentication.
type AuthConfig struct {
	// Enabled methods: bearer, jwt, hmac, mtls, apikey
	// Default: ["hmac"] (HS256 shared secret with auth service).
	Methods []string `yaml:"methods" json:"methods"`

	Bearer map[string]string `yaml:"bearer" json:"bearer"` // token → subject
	APIKey map[string]string `yaml:"apikey" json:"apikey"` // key → subject

	// HMACSecret is the HS256 key shared with SPACE_JWT_SECRET.
	HMACSecret string `yaml:"hmac_secret" json:"hmac_secret"`
	// HMACIssuer optional iss check (GATE_JWT_ISSUER / SPACE_JWT_ISSUER).
	HMACIssuer string `yaml:"hmac_issuer" json:"hmac_issuer"`

	JWT struct {
		Issuer   string `yaml:"issuer" json:"issuer"`
		Audience string `yaml:"audience" json:"audience"`
		JWKSURL  string `yaml:"jwks_url" json:"jwks_url"`
	} `yaml:"jwt" json:"jwt"`

	MTLS struct {
		AllowedCNs []string `yaml:"allowed_cns" json:"allowed_cns"`
	} `yaml:"mtls" json:"mtls"`

	// If true, auth is optional (legacy global flag; prefer per-route mini.auth)
	Optional bool `yaml:"optional" json:"optional"`
}

// CORSConfig configures CORS.
type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins" json:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" json:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" json:"allowed_headers"`
	ExposedHeaders   []string `yaml:"exposed_headers" json:"exposed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" json:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" json:"max_age"`
}

// RateLimitConfig configures rate limiting.
type RateLimitConfig struct {
	Enabled         bool          `yaml:"enabled" json:"enabled"`
	RequestsPerSec  float64       `yaml:"requests_per_sec" json:"requests_per_sec"`
	Burst           int           `yaml:"burst" json:"burst"`
	CleanupInterval time.Duration `yaml:"cleanup_interval" json:"cleanup_interval"`
}

// PolicyConfig configures the policy engine.
type PolicyConfig struct {
	Engine string `yaml:"engine" json:"engine"` // "cel" | "opa" | "none"
	// Path to policies or inline
	PoliciesPath string `yaml:"policies_path" json:"policies_path"`
	// For CEL — expression list
	CELExpressions []string `yaml:"cel_expressions" json:"cel_expressions"`
	// For OPA
	OPABundleURL string `yaml:"opa_bundle_url" json:"opa_bundle_url"`
}

// AdminConfig configures the admin plane.
type AdminConfig struct {
	Enabled       bool          `yaml:"enabled" json:"enabled"`
	ListenAddr    string        `yaml:"listen_addr" json:"listen_addr"`
	ConfigWatch   bool          `yaml:"config_watch" json:"config_watch"`
	WatchInterval time.Duration `yaml:"watch_interval" json:"watch_interval"`
}

// MetricsConfig configures metrics.
type MetricsConfig struct {
	Enabled bool   `yaml:"enabled" json:"enabled"`
	Path    string `yaml:"path" json:"path"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenAddr:      ":8080",
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    30 * time.Second,
			IdleTimeout:     60 * time.Second,
			ShutdownTimeout: 10 * time.Second,
			EnableHTTP2:     true,
		},
		Auth: AuthConfig{
			Methods:    []string{"hmac"},
			HMACSecret: "dev-secret-change-me",
			// Empty issuer = skip iss check (auth may use optima/micro depending on env).
			HMACIssuer: "",
			Optional:   false,
		},
		CORS: CORSConfig{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Authorization", "Content-Type", "X-API-Key", "X-Request-ID"},
			AllowCredentials: false,
			MaxAge:           86400,
		},
		RateLimit: RateLimitConfig{
			Enabled:         true,
			RequestsPerSec:  100,
			Burst:           200,
			CleanupInterval: 5 * time.Minute,
		},
		Policy: PolicyConfig{
			Engine: "none",
		},
		Admin: AdminConfig{
			Enabled:       false,
			ListenAddr:    ":8081",
			ConfigWatch:   true,
			WatchInterval: 5 * time.Second,
		},
		Metrics: MetricsConfig{
			Enabled: true,
			Path:    "/metrics",
		},
		NATSURL:          "nats://127.0.0.1:4222",
		DiscoverInterval: 5 * time.Second,
		DiscoverWait:     400 * time.Millisecond,
		RequestTimeout:   30 * time.Second,
		MaxBodyBytes:     32 << 20,
		ClaimHeaders:     "typ:X-Mini-Auth-Type,platform_role:X-Mini-Platform-Role,space_id:X-Mini-Space-Id,space_role:X-Mini-Space-Role,email:X-Mini-Email,role:X-Mini-Role,tv:X-Mini-Token-Version",
	}
}

// Validate checks configuration validity.
func (c *Config) Validate() error {
	if c.Server.ListenAddr == "" {
		return fmt.Errorf("server.listen_addr is required")
	}
	if c.RateLimit.Enabled && c.RateLimit.RequestsPerSec <= 0 {
		return fmt.Errorf("rate_limit.requests_per_sec must be > 0")
	}
	return nil
}
