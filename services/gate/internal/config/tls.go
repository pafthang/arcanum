package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// TLSConfig configures TLS.
type TLSConfig struct {
	Enabled            bool   `yaml:"enabled" json:"enabled"`
	CertFile           string `yaml:"cert_file" json:"cert_file"`
	KeyFile            string `yaml:"key_file" json:"key_file"`
	ClientCAFile       string `yaml:"client_ca_file" json:"client_ca_file"` // for mTLS
	MinVersion         string `yaml:"min_version" json:"min_version"`       // "1.2" | "1.3"
	ClientAuth         string `yaml:"client_auth" json:"client_auth"`       // "none" | "request" | "require"
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`
}

// BuildTLSConfig builds *tls.Config from TLSConfig.
func (c *TLSConfig) BuildTLSConfig() (*tls.Config, error) {
	if !c.Enabled {
		return nil, nil
	}

	if c.CertFile == "" || c.KeyFile == "" {
		return nil, fmt.Errorf("tls: cert_file and key_file are required when tls.enabled=true")
	}

	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("tls: load key pair: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	switch c.MinVersion {
	case "1.3":
		tlsCfg.MinVersion = tls.VersionTLS13
	case "1.2", "":
		tlsCfg.MinVersion = tls.VersionTLS12
	default:
		return nil, fmt.Errorf("tls: unsupported min_version %q", c.MinVersion)
	}

	// Client auth (mTLS)
	switch c.ClientAuth {
	case "require", "RequireAndVerifyClientCert":
		tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
	case "request", "RequestClientCert":
		tlsCfg.ClientAuth = tls.RequestClientCert
	case "none", "":
		tlsCfg.ClientAuth = tls.NoClientCert
	default:
		return nil, fmt.Errorf("tls: unsupported client_auth %q", c.ClientAuth)
	}

	if c.ClientCAFile != "" {
		caPEM, err := os.ReadFile(c.ClientCAFile)
		if err != nil {
			return nil, fmt.Errorf("tls: read client CA: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("tls: failed to parse client CA")
		}
		tlsCfg.ClientCAs = pool
	}

	if c.InsecureSkipVerify {
		tlsCfg.InsecureSkipVerify = true
	}

	return tlsCfg, nil
}
