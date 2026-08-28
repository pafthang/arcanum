package svcutil

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"path/filepath"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/logx"
	"github.com/pafthang/arcanum/pkg/mini"
)

// App is a started mini domain service process (NATS + service + optional outbox).
// Use Bootstrap for the common wiring path; gate/nats/ctrl may stay custom.
type App struct {
	Name    string
	Version string
	DataDir string
	NC      *nats.Conn
	Svc     mini.Service
	Ctx     context.Context
	cancel  context.CancelFunc
	workers []func(context.Context) error
}

// BootstrapOption allows customizing the bootstrap process.
type BootstrapOption func(*bootstrapConfig)

type bootstrapConfig struct {
	disableNATSLogging bool
}

// WithoutNATSLogging disables NATS structured logging (useful for NATS itself).
func WithoutNATSLogging() BootstrapOption {
	return func(c *bootstrapConfig) {
		c.disableNATSLogging = true
	}
}

// Bootstrap loads cfgs, connects NATS, creates a mini service with standard middleware,
// and returns an App. Caller registers endpoints, WireLifecycle,
// then Wait.
//
//	app := svcutil.Bootstrap("task", "0.3.0", "Team projects…")
//	defer app.Shutdown()
//	store, err := OpenStore(app.DataDir)
//	…
//	app.Wait()
func Bootstrap(name, version, description string, opts ...BootstrapOption) *App {
	cfg := &bootstrapConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	nc := ConnectNATS(name)
	ctx, cancel := context.WithCancel(context.Background())
	svc := NewService(nc, name, version, description)
	appDataDir := DataDir(name)

	var logNATS *nats.Conn
	if !cfg.disableNATSLogging {
		logNATS = nc
	}

	_ = logx.Setup(logx.Config{
		Term:    true,
		File:    filepath.Join(appDataDir, "service.log"),
		NATS:    logNATS,
		Subject: "logs." + name,
	})

	return &App{
		Name:    name,
		Version: version,
		DataDir: appDataDir,
		NC:      nc,
		Svc:     svc,
		Ctx:     ctx,
		cancel:  cancel,
	}
}

// BootstrapWithConfig loads env config, applies it, and runs Bootstrap.
// Example: app, cfg := svcutil.BootstrapWithConfig("gate", "0.1.0", "...", config.FromEnv)
func BootstrapWithConfig[T any](name, version, description string, fromEnv func() T, opts ...BootstrapOption) (*App, T) {
	LoadConfig(name)
	cfg := fromEnv()
	app := Bootstrap(name, version, description, opts...)
	return app, cfg
}

// AddWorker registers a background task that should run while the App is running.
// It will be started in a goroutine and stopped when the App is shutting down.
func (a *App) AddWorker(worker func(ctx context.Context) error) {
	a.workers = append(a.workers, worker)
}

// WireLifecycle subscribes to platform reload/restart for this service.
func (a *App) WireLifecycle(reloader lifecycle.Reloader) {
	if a == nil {
		return
	}
	WireLifecycle(a.NC, a.Name, reloader)
}

// Wait starts all workers and blocks until SIGINT/SIGTERM or a worker fails.
// It then cancels context, stops the mini service, and closes NATS.
func (a *App) Wait() {
	if a == nil {
		return
	}

	errCh := make(chan error, len(a.workers))
	for _, w := range a.workers {
		worker := w
		go func() {
			if err := worker(a.Ctx); err != nil {
				errCh <- err
			}
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		slog.Info("shutdown signal received", "app", a.Name)
	case err := <-errCh:
		slog.Error("worker stopped unexpectedly", "err", err, "app", a.Name)
	}

	a.Shutdown()
}

// Shutdown cancels context, stops the service, and closes NATS (idempotent-ish).
func (a *App) Shutdown() {
	if a == nil {
		return
	}
	if a.cancel != nil {
		a.cancel()
		a.cancel = nil
	}
	if a.Svc != nil {
		if err := a.Svc.Stop(); err != nil {
			slog.Debug("service stop", "name", a.Name, "err", err)
		}
	}
	if a.NC != nil {
		a.NC.Close()
		a.NC = nil
	}
}
