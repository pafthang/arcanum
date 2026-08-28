// Package gate is the HTTP edge for Optima: auth, routing, policy, proxy, WS fan-out.
//
// Layout (edge style — same skeleton as exec/cron):
//
//	service.go   — Service, NewService, Run (process entry)
//	export.go    — public type aliases + config helpers for root package consumers
//	client/      — typed HTTP client for callers outside the process
//	cmd/gate/ — thin main → gate.Run()
//	internal/    — implementation (not importable from other modules)
//	  config/      env + defaults
//	  core/        HTTP server wiring
//	  auth/        authenticators + middleware
//	  middleware/  CORS, rate limit
//	  policy/      CEL / OPA engine
//	  proxy/       reverse proxy + upload
//	  routing/     route table + cache
//	  websocket/   WS ACL + fan-out
//	  ops/         health, metrics, OpenAPI
//	  admin/       config plane / setup
//	  models/      registry + WS ACL types
//
// HTTP edge: discovers public mini routes via $SRV.INFO and proxies to public.* subjects.
// Domain mini services use the agent-style layout (apis/, config/, store/, …).
package gate

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/gate/internal/config"
	"github.com/pafthang/arcanum/services/gate/internal/core"
)

// Service is the gate process (HTTP edge).
type Service struct {
	cfg *config.Config
	gw  *core.Gate
}

// NewService constructs the gate service from cfg.
// If cfg is nil, defaults + env are applied.
func NewService(cfg *config.Config, nc *nats.Conn) (*Service, error) {
	if cfg == nil {
		cfg = config.DefaultConfig()
		// Apply cfgs/comm.json + cfgs/gate.json into env (missing files OK).
		svcutil.LoadConfig("gate")
		config.FromEnv(cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	gw, err := core.New(cfg, nc)
	if err != nil {
		return nil, fmt.Errorf("gate: %w", err)
	}

	return &Service{cfg: cfg, gw: gw}, nil
}

// Run starts the HTTP edge and blocks until ctx is cancelled.
func (s *Service) Run(ctx context.Context) error {
	if s == nil || s.gw == nil {
		return fmt.Errorf("gate service not initialized")
	}
	return s.gw.Start(ctx)
}

// Gate returns the underlying core Gate.
func (s *Service) Gate() *core.Gate {
	return s.gw
}

// Config returns the active configuration.
func (s *Service) Config() *config.Config {
	return s.cfg
}

// Run is the process entrypoint (load config, start service, wait for signal).
func Run() {
	fromEnv := func() *config.Config {
		cfg := config.DefaultConfig()
		config.FromEnv(cfg)
		if err := cfg.Validate(); err != nil {
			log.Fatalf("gate config: %v", err)
		}
		return cfg
	}

	app, cfg := svcutil.BootstrapWithConfig("gate", "0.2.0", "HTTP Edge and Proxy", fromEnv)
	defer app.Shutdown()

	svc, err := NewService(cfg, app.NC)
	if err != nil {
		log.Fatalf("gate: %v", err)
	}

	slog.Info("gate starting", "listen", cfg.Server.ListenAddr)

	app.AddWorker(func(ctx context.Context) error {
		if err := svc.Run(ctx); err != nil && err.Error() != "http: Server closed" {
			return err
		}
		return nil
	})

	app.Wait()
	slog.Info("gate stopped")
}
