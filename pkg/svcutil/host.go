package svcutil

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/lifecycle"
)

// Module is one unit of domain mounted into a process Host.
// Typical layout: each module registers its own mini.Service (name == Name())
// on the shared NATS connection — same pattern as platform/* and exec/cron today.
//
// Modules must not call WaitSignal or close the host NC; Host owns process lifecycle.
type Module interface {
	Name() string
	// Start mounts endpoints, loops, stores. stop is always called on host shutdown
	// (reverse start order), even if later modules fail to start.
	Start(ctx ModuleContext) (stop func(), err error)
}

// ModuleContext is the host-provided environment for a module start.
type ModuleContext struct {
	// Process is the OS process / cfgs name (e.g. "platform", "exec").
	Process string
	// Name is this module's Name() (e.g. "memo", "cron").
	Name string
	// NC is the shared NATS connection for the process.
	NC *nats.Conn
	// Ctx is cancelled when the host begins shutdown (before stop hooks).
	Ctx context.Context
	// Logg is process logg (defaults to slog.Default).
	Logg *slog.Logger
}

// DataDir returns the conventional data directory for this module (data/<name>).
func (c ModuleContext) DataDir() string {
	if c.Name == "" {
		return DataDir(c.Process)
	}
	return DataDir(c.Name)
}

// Host composes modules into one OS process with a single NATS connection.
//
//	svcutil.NewHost("exec").
//	    Module(cron.Module()).
//	    Module(coreModule()).
//	    Run()
//
// Or attach an existing connection (e.g. tests / custom bootstrap):
//
//	stop, err := svcutil.NewHost("exec").WithNC(nc).Module(cron.Module()).Start()
//	defer stop()
type Host struct {
	name     string
	modules  []Module
	nc       *nats.Conn
	ownNC    bool // close NC on Shutdown when we connected it
	logg     *slog.Logger
	reloader lifecycle.Reloader
	// processLC wires WireLifecycle for the host process name (default true).
	// Disable when the core module already owns process lifecycle (e.g. exec).
	processLC bool
	// disabled modules by name (Skip)
	skip map[string]bool

	mu      sync.Mutex
	started []startedMod
	cancel  context.CancelFunc
	ctx     context.Context
}

type startedMod struct {
	name string
	stop func()
}

// NewHost creates a process host. name is used for NATS client name, cfgs/<name>.json,
// and process-level lifecycle (WireLifecycle).
func NewHost(name string) *Host {
	return &Host{
		name:      name,
		logg:      slog.Default(),
		skip:      map[string]bool{},
		processLC: true,
	}
}

// Module registers a module (start order = registration order).
func (h *Host) Module(m Module) *Host {
	if h == nil || m == nil {
		return h
	}
	h.modules = append(h.modules, m)
	return h
}

// Modules registers several modules in order.
func (h *Host) Modules(ms ...Module) *Host {
	for _, m := range ms {
		h.Module(m)
	}
	return h
}

// Skip disables named modules (no-op if not registered).
func (h *Host) Skip(names ...string) *Host {
	if h == nil {
		return h
	}
	if h.skip == nil {
		h.skip = map[string]bool{}
	}
	for _, n := range names {
		h.skip[n] = true
	}
	return h
}

// WithNC uses an existing connection. Host will not LoadConfig/Connect or Close it.
func (h *Host) WithNC(nc *nats.Conn) *Host {
	if h == nil {
		return h
	}
	h.nc = nc
	h.ownNC = false
	return h
}

// WithLogger overrides the default slog logg.
func (h *Host) WithLogger(l *slog.Logger) *Host {
	if h == nil {
		return h
	}
	if l != nil {
		h.logg = l
	}
	return h
}

// WithReloader sets process-level soft reload (WireLifecycle on host name).
func (h *Host) WithReloader(r lifecycle.Reloader) *Host {
	if h == nil {
		return h
	}
	h.reloader = r
	return h
}

// NoProcessLifecycle skips WireLifecycle for the host process name.
// Use when a core module already registers lifecycle under the same name.
func (h *Host) NoProcessLifecycle() *Host {
	if h == nil {
		return h
	}
	h.processLC = false
	return h
}

// Start connects (unless WithNC), starts enabled modules, wires process lifecycle.
// Returns a stop func that cancels context and stops modules in reverse order.
// Does not WaitSignal — use Run for the full process main.
func (h *Host) Start() (stop func(), err error) {
	if h == nil {
		return func() {}, fmt.Errorf("nil host")
	}
	if h.name == "" {
		return func() {}, fmt.Errorf("host name required")
	}

	h.mu.Lock()
	if h.ctx != nil {
		h.mu.Unlock()
		return func() {}, fmt.Errorf("host %s already started", h.name)
	}
	ctx, cancel := context.WithCancel(context.Background())
	h.ctx = ctx
	h.cancel = cancel
	h.mu.Unlock()

	if h.nc == nil {
		h.nc = ConnectNATS(h.name)
		h.ownNC = true
	}
	if h.logg == nil {
		h.logg = slog.Default()
	}

	if h.processLC {
		WireLifecycle(h.nc, h.name, h.reloader)
	}

	var names []string
	for _, m := range h.modules {
		if m == nil {
			continue
		}
		name := m.Name()
		if name == "" {
			cancel()
			h.stopStarted()
			return func() {}, fmt.Errorf("module with empty name")
		}
		if h.skip[name] {
			h.logg.Info("host: module skipped", "process", h.name, "module", name)
			continue
		}
		mctx := ModuleContext{
			Process: h.name,
			Name:    name,
			NC:      h.nc,
			Ctx:     ctx,
			Logg:    h.logg.With("module", name),
		}
		st, err := m.Start(mctx)
		if err != nil {
			cancel()
			h.stopStarted()
			if h.ownNC && h.nc != nil {
				h.nc.Close()
				h.nc = nil
			}
			return func() {}, fmt.Errorf("module %s: %w", name, err)
		}
		if st == nil {
			st = func() {}
		}
		h.mu.Lock()
		h.started = append(h.started, startedMod{name: name, stop: st})
		h.mu.Unlock()
		names = append(names, name)
	}

	h.logg.Info("host started", "process", h.name, "modules", names)

	var once sync.Once
	return func() {
		once.Do(func() {
			h.Shutdown()
		})
	}, nil
}

// Run starts the host and blocks until SIGINT/SIGTERM, then shuts down.
// Fatals on start errors (process entrypoint).
func (h *Host) Run() {
	stop, err := h.Start()
	if err != nil {
		log.Fatalf("host %s: %v", h.name, err)
	}
	WaitSignal()
	stop()
}

// Shutdown cancels context, stops modules (LIFO), and closes NC if owned.
func (h *Host) Shutdown() {
	if h == nil {
		return
	}
	h.mu.Lock()
	cancel := h.cancel
	h.cancel = nil
	h.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	h.stopStarted()
	h.mu.Lock()
	nc, own := h.nc, h.ownNC
	if own {
		h.nc = nil
	}
	h.ctx = nil
	h.mu.Unlock()
	if own && nc != nil {
		nc.Close()
	}
}

func (h *Host) stopStarted() {
	h.mu.Lock()
	started := h.started
	h.started = nil
	h.mu.Unlock()
	for i := len(started) - 1; i >= 0; i-- {
		s := started[i]
		func() {
			defer func() {
				if r := recover(); r != nil {
					h.logg.Error("host: module stop panic", "module", s.name, "panic", r)
				}
			}()
			if s.stop != nil {
				s.stop()
			}
		}()
	}
}

// NC returns the host connection after Start (nil before).
func (h *Host) NC() *nats.Conn {
	if h == nil {
		return nil
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.nc
}

// --- adapters ---

type funcModule struct {
	name  string
	start func(ModuleContext) (func(), error)
}

func (m funcModule) Name() string { return m.name }

func (m funcModule) Start(ctx ModuleContext) (func(), error) {
	if m.start == nil {
		return func() {}, nil
	}
	return m.start(ctx)
}

// ModuleFunc builds a Module from a name and start function.
func ModuleFunc(name string, start func(ModuleContext) (func(), error)) Module {
	return funcModule{name: name, start: start}
}

// ModuleOf adapts the legacy Start(nc) (stop func()) pattern used by platform/* and cron.
// Panics inside start still propagate (same as today). Prefer ModuleFunc for new code.
func ModuleOf(name string, start func(nc *nats.Conn) (stop func())) Module {
	return ModuleFunc(name, func(ctx ModuleContext) (func(), error) {
		if start == nil {
			return func() {}, nil
		}
		return start(ctx.NC), nil
	})
}
