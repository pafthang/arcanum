package logx

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/nats-io/nats.go"
)

// Config configures the global logg sinks.
type Config struct {
	Term    bool
	File    string
	NATS    *nats.Conn
	Subject string
}

// Setup replaces the global slog default logg with a multiplexed handler.
func Setup(cfg Config) error {
	var handlers []slog.Handler

	if cfg.Term {
		handlers = append(handlers, slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	if cfg.File != "" {
		f, err := os.OpenFile(cfg.File, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		handlers = append(handlers, slog.NewJSONHandler(f, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}))
	}

	if cfg.NATS != nil && cfg.Subject != "" {
		handlers = append(handlers, &natsHandler{
			nc:   cfg.NATS,
			subj: cfg.Subject,
		})
	}

	if len(handlers) == 0 {
		return nil
	}

	slog.SetDefault(slog.New(&multiHandler{handlers: handlers}))
	return nil
}

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, lvl) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			_ = h.Handle(ctx, r)
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	var clones []slog.Handler
	for _, h := range m.handlers {
		clones = append(clones, h.WithAttrs(attrs))
	}
	return &multiHandler{handlers: clones}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	var clones []slog.Handler
	for _, h := range m.handlers {
		clones = append(clones, h.WithGroup(name))
	}
	return &multiHandler{handlers: clones}
}

// natsHandler publishes log records as JSON to NATS.
type natsHandler struct {
	nc   *nats.Conn
	subj string
}

func (n *natsHandler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return true
}

func (n *natsHandler) Handle(ctx context.Context, r slog.Record) error {
	msg := map[string]any{
		"time":    r.Time.Format("2006-01-02T15:04:05.999Z07:00"),
		"level":   r.Level.String(),
		"message": r.Message,
	}
	r.Attrs(func(a slog.Attr) bool {
		msg[a.Key] = a.Value.Any()
		return true
	})

	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return n.nc.Publish(n.subj, b)
}

func (n *natsHandler) WithAttrs(attrs []slog.Attr) slog.Handler { return n }
func (n *natsHandler) WithGroup(name string) slog.Handler       { return n }
