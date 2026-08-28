// Package svcutil provides common bootstrap helpers for mini services.
package svcutil

import (
	"log"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/access"
	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/mini"
)

// Env returns env value or default.
func Env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// DataDir returns DATA_DIR or ./data/<service>.
func DataDir(service string) string {
	if v := os.Getenv("DATA_DIR"); v != "" {
		return v
	}
	return filepath.Join("data", service)
}

// LoadConfig applies cfgs/comm.json + cfgs/<name>.json into process env
// (only keys not already set). See mini.ApplyServiceConfig.
// Fatal on parse errors; missing files are fine.
func LoadConfig(name string) mini.ServiceFileConfig {
	cfg, err := mini.ApplyServiceConfig(name)
	if err != nil {
		log.Fatalf("service config %s: %v", name, err)
	}
	return cfg
}

// ConnectNATS connects to NATS with a named client.
// Requires LoadConfig(name) to have been called by the service entrypoint.
// It will wait for NATS to become available for up to 30 seconds.
func ConnectNATS(name string) *nats.Conn {
	url := Env("NATS_URL", nats.DefaultURL)
	timeout := EnvDuration("NATS_CONNECT_TIMEOUT", 30*time.Second)
	deadline := time.Now().Add(timeout)
	logged := false

	var nc *nats.Conn
	var err error
	for {
		nc, err = nats.Connect(url,
			nats.Name(name),
			nats.MaxReconnects(-1),
			nats.ReconnectWait(time.Second),
		)
		if err == nil {
			return nc
		}
		if !logged {
			slog.Info("waiting for nats", "url", url, "err", err)
			logged = true
		}
		if time.Now().After(deadline) {
			log.Fatalf("nats connect timed out: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// NewService creates a mini service with recover, request-id, logging, and
// access-log middleware. AdvertiseWhenPublic is enabled so gateways refresh
// routes immediately when public endpoints are registered.
func NewService(nc *nats.Conn, name, version, description string) mini.Service {
	logg := slog.Default()
	mws := []mini.Middleware{
		mini.Recover(logg),
		mini.RequestID(),
		mini.Logging(logg),
		access.Middleware(nc, name),
	}
	svc, err := mini.AddService(nc, mini.Config{
		Name:                name,
		Version:             version,
		Description:         description,
		Middleware:          mws,
		AdvertiseWhenPublic: true,
	})
	if err != nil {
		log.Fatalf("add service %s: %v", name, err)
	}
	return svc
}

// WaitSignal blocks until SIGINT/SIGTERM.
func WaitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}

// Must registers endpoint or fatals.
func Must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// WireLifecycle subscribes to platform reload/restart commands for this service.
// After restore, ops publishes restart; pair with scripts/dev-up auto-restart loop.
func WireLifecycle(nc *nats.Conn, name string, reloader lifecycle.Reloader) {
	if err := lifecycle.Listen(nc, lifecycle.Options{
		Name:     name,
		Reloader: reloader,
		Logg:     slog.Default(),
	}); err != nil {
		log.Fatalf("lifecycle listen %s: %v", name, err)
	}
}
