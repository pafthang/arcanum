package supervise

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pafthang/arcanum/pkg/mini"
)

// SupervisorConfig controls process supervision (ctrl -up).
type SupervisorConfig struct {
	// Root is the module root (cwd for go run ./cmd/...).
	Root string
	// Skip names that should not be started (always includes "ctrl").
	Skip map[string]struct{}
	// RestartDelay between child restarts.
	RestartDelay time.Duration
	// NATSWait how long to wait for nats after starting it.
	NATSWait time.Duration
	// Logg optional.
	Logg *slog.Logger
}

// StartSupervisor launches enabled services from cfgs/ with auto-restart.
// Blocks until ctx is cancelled, then kills children.
func StartSupervisor(ctx context.Context, sc SupervisorConfig) error {
	if sc.Logg == nil {
		sc.Logg = slog.Default()
	}
	if sc.Root == "" {
		root, err := findModuleRoot()
		if err != nil {
			return err
		}
		sc.Root = root
	}
	if sc.RestartDelay <= 0 {
		sc.RestartDelay = time.Second
	}
	if sc.NATSWait <= 0 {
		sc.NATSWait = 45 * time.Second
	}
	if sc.Skip == nil {
		sc.Skip = map[string]struct{}{}
	}
	// Never supervise ourselves.
	sc.Skip["ctrl"] = struct{}{}

	list, err := mini.ListServiceFiles()
	if err != nil {
		return fmt.Errorf("list cfgs: %w", err)
	}
	var services []mini.ServiceFileConfig
	for _, c := range list {
		if !c.IsEnabled() {
			sc.Logg.Info("ctrl: skip disabled service", "name", c.Name)
			continue
		}
		if _, skip := sc.Skip[c.Name]; skip {
			sc.Logg.Info("ctrl: skip service", "name", c.Name)
			continue
		}
		services = append(services, c)
	}
	if len(services) == 0 {
		return fmt.Errorf("no enabled services in %s", mini.ConfigDir())
	}

	sc.Logg.Info("ctrl: supervising stack",
		"root", sc.Root,
		"cfg_dir", mini.ConfigDir(),
		"count", len(services),
	)

	var wg sync.WaitGroup
	// Start in order; wait for nats readiness before later services.
	for _, cfg := range services {
		if ctx.Err() != nil {
			break
		}
		name := cfg.Name
		wg.Add(1)
		go func(cfg mini.ServiceFileConfig) {
			defer wg.Done()
			runServiceLoop(ctx, sc, cfg)
		}(cfg)

		if name == "nats" {
			host := envFromCfg(cfg, "NATS_HOST", "127.0.0.1")
			port := envFromCfg(cfg, "NATS_PORT", "4222")
			if err := waitTCP(ctx, net.JoinHostPort(host, port), sc.NATSWait); err != nil {
				sc.Logg.Error("ctrl: nats not ready", "err", err)
				// still continue — other services will reconnect
			} else {
				sc.Logg.Info("ctrl: nats ready", "addr", net.JoinHostPort(host, port))
			}
		} else {
			// small stagger so go run doesn't thundering-herd the compiler
			select {
			case <-ctx.Done():
			case <-time.After(150 * time.Millisecond):
			}
		}
	}

	<-ctx.Done()
	sc.Logg.Info("ctrl: shutting down supervised children")
	wg.Wait()
	return nil
}

func runServiceLoop(ctx context.Context, sc SupervisorConfig, cfg mini.ServiceFileConfig) {
	name := cfg.Name
	logPath := serviceLogPath(sc.Root, name, cfg)
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		sc.Logg.Error("ctrl: mkdir log dir", "service", name, "err", err)
	}

	delay := sc.RestartDelay
	for {
		if ctx.Err() != nil {
			return
		}
		cmdArgs := mini.ResolveCommand(cfg)
		sc.Logg.Info("ctrl: starting", "service", name, "cmd", strings.Join(cmdArgs, " "), "log", logPath)

		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = sc.Root
		cmd.Env = mini.ChildEnv(cfg)
		cmd.Cancel = func() error {
			sc.Logg.Info("ctrl: sending SIGINT to service", "service", name)
			return cmd.Process.Signal(os.Interrupt)
		}
		cmd.WaitDelay = 5 * time.Second

		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			sc.Logg.Error("ctrl: open log", "service", name, "path", logPath, "err", err)
			sleepCtx(ctx, delay)
			continue
		}
		cmd.Stdout = f
		cmd.Stderr = f

		start := time.Now()
		err = cmd.Run()
		_ = f.Close()
		dur := time.Since(start)

		if ctx.Err() != nil {
			return
		}

		if dur > 15*time.Second {
			// If it ran successfully for a while, reset the backoff
			delay = sc.RestartDelay
		} else {
			// Exponential backoff with a cap of 30 seconds
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}

		code := 0
		if err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else {
				sc.Logg.Error("ctrl: run error", "service", name, "err", err, "dur", dur, "next_delay", delay)
				sleepCtx(ctx, delay)
				continue
			}
		}
		sc.Logg.Info("ctrl: exited, restarting",
			"service", name, "code", code, "dur", dur, "next_delay", delay)
		sleepCtx(ctx, delay)
	}
}

func envFromCfg(cfg mini.ServiceFileConfig, key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if cfg.Env != nil {
		if v, ok := cfg.Env[key]; ok && v != "" {
			return v
		}
	}
	return def
}

// serviceLogPath returns data/<service>/service.log (or DATA_DIR/service.log).
func serviceLogPath(root, name string, cfg mini.ServiceFileConfig) string {
	dir := filepath.Join(root, "data", name)
	if d := envFromCfg(cfg, "DATA_DIR", ""); d != "" {
		if filepath.IsAbs(d) {
			dir = d
		} else {
			dir = filepath.Join(root, d)
		}
	}
	return filepath.Join(dir, "service.log")
}

func waitTCP(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	var last error
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		d := net.Dialer{Timeout: time.Second}
		c, err := d.DialContext(ctx, "tcp", addr)
		if err == nil {
			_ = c.Close()
			return nil
		}
		last = err
		sleepCtx(ctx, 500*time.Millisecond)
	}
	if last == nil {
		last = fmt.Errorf("timeout waiting for %s", addr)
	}
	return last
}

func sleepCtx(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func findModuleRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd, nil // fallback to cwd
		}
		dir = parent
	}
}
