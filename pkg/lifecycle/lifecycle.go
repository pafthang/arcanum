// Package lifecycle handles platform reload/restart commands over NATS.
package lifecycle

import (
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/subjects"
)

// Reloader is implemented by services that can soft-reopen resources (SQLite, caches).
type Reloader interface {
	Reload() error
}

// ReloaderFunc adapts a function to Reloader.
type ReloaderFunc func() error

// Reload implements Reloader.
func (f ReloaderFunc) Reload() error { return f() }

// Options configures lifecycle listeners for a service instance.
type Options struct {
	// Name is the service name (auth, exec, agent, …).
	Name string
	// Reloader optional soft reload hook.
	Reloader Reloader
	// OnRestart optional hook before os.Exit (flush outbox, etc.).
	OnRestart func()
	// Logg optional.
	Logg *slog.Logger
}

// Listen subscribes to platform reload/restart commands.
// Safe to call once per process.
func Listen(nc *nats.Conn, opt Options) error {
	if opt.Logg == nil {
		opt.Logg = slog.Default()
	}
	name := opt.Name

	handle := func(kind string, msg *nats.Msg) {
		cmd, ok := decode(msg.Data)
		if !ok {
			return
		}
		if !targets(name, cmd.Services) {
			return
		}
		delay := time.Duration(cmd.DelayMs) * time.Millisecond
		if delay <= 0 {
			if kind == "restart" {
				delay = 300 * time.Millisecond
			} else {
				delay = 50 * time.Millisecond
			}
		}
		opt.Logg.Info("lifecycle: command",
			"kind", kind, "service", name, "reason", cmd.Reason, "delay", delay)

		time.AfterFunc(delay, func() {
			switch kind {
			case "reload":
				if opt.Reloader == nil {
					opt.Logg.Info("lifecycle: reload not supported, no-op", "service", name)
					return
				}
				if err := opt.Reloader.Reload(); err != nil {
					opt.Logg.Error("lifecycle: reload failed", "service", name, "err", err)
					return
				}
				opt.Logg.Info("lifecycle: reloaded", "service", name)
			case "restart":
				if opt.OnRestart != nil {
					opt.OnRestart()
				}
				opt.Logg.Info("lifecycle: exiting for restart", "service", name)
				os.Exit(0)
			}
		})
	}

	if _, err := nc.Subscribe(subjects.CommandPlatformReload, func(msg *nats.Msg) {
		handle("reload", msg)
	}); err != nil {
		return err
	}
	if _, err := nc.Subscribe(subjects.CommandPlatformRestart, func(msg *nats.Msg) {
		handle("restart", msg)
	}); err != nil {
		return err
	}
	return nil
}

// PublishReload broadcasts a soft reload request.
func PublishReload(nc *nats.Conn, services []string, reason string, delayMs int) error {
	return events.PublishData(nc, subjects.CommandPlatformReload, "platform.reload", "ops", events.LifecycleCommand{
		Services: services, Reason: reason, DelayMs: delayMs,
	})
}

// PublishRestart broadcasts a hard restart request (process exit).
func PublishRestart(nc *nats.Conn, services []string, reason string, delayMs int) error {
	return events.PublishData(nc, subjects.CommandPlatformRestart, "platform.restart", "ops", events.LifecycleCommand{
		Services: services, Reason: reason, DelayMs: delayMs,
	})
}

func decode(data []byte) (events.LifecycleCommand, bool) {
	_, cmd, err := events.Decode[events.LifecycleCommand](data)
	if err == nil {
		return cmd, true
	}
	var bare events.LifecycleCommand
	if err := json.Unmarshal(data, &bare); err != nil {
		return events.LifecycleCommand{}, false
	}
	return bare, true
}

func targets(name string, services []string) bool {
	if len(services) == 0 {
		return true
	}
	for _, s := range services {
		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" || s == "*" || s == "all" {
			return true
		}
		if s == strings.ToLower(name) {
			return true
		}
	}
	return false
}
