package edgecfg

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"github.com/pafthang/arcanum/pkg/mini"
)

// Snapshot is a consistent view of config plane routes + ACL for gateways.
type Snapshot struct {
	Routes    []mini.PublicRoute
	WSACL     []WSACLDoc
	Revision  uint64
	UpdatedAt time.Time
}

// Watcher watches KV and produces Snapshots on change.
type Watcher struct {
	store  *Store
	log    *slog.Logger
	mu     sync.RWMutex
	snap   Snapshot
	notify chan struct{}
}

// NewWatcher creates a watcher. Call Run to start.
func NewWatcher(store *Store, log *slog.Logger) *Watcher {
	if log == nil {
		log = slog.Default()
	}
	return &Watcher{
		store:  store,
		log:    log,
		notify: make(chan struct{}, 1),
	}
}

// Snapshot returns the latest loaded snapshot.
func (w *Watcher) Snapshot() Snapshot {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.snap
}

// Changes returns a conn signaled when the snapshot updates (coalesced).
func (w *Watcher) Changes() <-chan struct{} { return w.notify }

// Reload loads full state from KV once.
func (w *Watcher) Reload(ctx context.Context) error {
	routes, err := w.store.ListRoutes(ctx, false)
	if err != nil {
		return err
	}
	acls, err := w.store.ListWSACL(ctx, false)
	if err != nil {
		return err
	}
	rev, _ := w.store.Revision(ctx)

	pr := make([]mini.PublicRoute, 0, len(routes))
	for _, d := range routes {
		r, err := d.ToPublicRoute()
		if err != nil {
			w.log.Warn("config: skip invalid route", "id", d.ID, "err", err)
			continue
		}
		pr = append(pr, r)
	}

	w.mu.Lock()
	w.snap = Snapshot{
		Routes:    pr,
		WSACL:     acls,
		Revision:  rev,
		UpdatedAt: time.Now().UTC(),
	}
	w.mu.Unlock()
	w.signal()
	return nil
}

func (w *Watcher) signal() {
	select {
	case w.notify <- struct{}{}:
	default:
	}
}

// Run watches KV prefixes and reloads on any change until ctx is done.
func (w *Watcher) Run(ctx context.Context) error {
	if err := w.Reload(ctx); err != nil {
		return err
	}

	// Watch all keys; filter prefixes on events.
	watcher, err := w.store.kv.WatchAll(ctx)
	if err != nil {
		return err
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case e, ok := <-watcher.Updates():
			if !ok {
				return nil
			}
			if e == nil {
				// initial done marker
				continue
			}
			key := e.Key()
			if key != KeyRevision &&
				!(len(key) >= len(PrefixRoutes) && key[:len(PrefixRoutes)] == PrefixRoutes) &&
				!(len(key) >= len(PrefixWSACL) && key[:len(PrefixWSACL)] == PrefixWSACL) {
				continue
			}
			// Coalesce bursts: reload full snapshot.
			if err := w.Reload(ctx); err != nil {
				w.log.Warn("config: reload failed", "err", err)
				continue
			}
			w.log.Info("config plane snapshot updated",
				"revision", w.Snapshot().Revision,
				"routes", len(w.Snapshot().Routes),
				"wsacl", len(w.Snapshot().WSACL),
				"op", e.Operation(),
				"key", key,
			)
		}
	}
}

// Ensure jetstream import used if Watch opts needed later.
var _ = jetstream.IgnoreDeletes
