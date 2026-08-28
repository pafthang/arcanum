package edge

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
)

// Discoverer periodically refreshes the public route table from $SRV.INFO.
type Discoverer struct {
	NC       *nats.Conn
	Table    *Table
	Interval time.Duration
	Wait     time.Duration
	Log      *slog.Logger

	mu   sync.Mutex
	sub  *nats.Subscription
	stop chan struct{}
}

// Start begins discovery loop and advertise subscription.
func (d *Discoverer) Start(ctx context.Context) {
	if d.Interval <= 0 {
		d.Interval = 5 * time.Second
	}
	if d.Wait <= 0 {
		d.Wait = 400 * time.Millisecond
	}
	d.stop = make(chan struct{})

	// Immediate refresh
	d.refresh(ctx)

	if sub, err := mini.SubscribeAdvertise(d.NC, func() {
		// short-lived context for advertise-triggered refresh
		cctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		d.refresh(cctx)
	}); err == nil {
		d.mu.Lock()
		d.sub = sub
		d.mu.Unlock()
	} else if d.Log != nil {
		d.Log.Warn("gate: advertise subscribe failed", "err", err)
	}

	go func() {
		t := time.NewTicker(d.Interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-d.stop:
				return
			case <-t.C:
				d.refresh(ctx)
			}
		}
	}()
}

// Stop unsubscribes advertise.
func (d *Discoverer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stop != nil {
		select {
		case <-d.stop:
		default:
			close(d.stop)
		}
	}
	if d.sub != nil {
		_ = d.sub.Unsubscribe()
		d.sub = nil
	}
}

func (d *Discoverer) refresh(ctx context.Context) {
	if d.NC == nil || d.Table == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, d.Wait+time.Second)
	defer cancel()
	routes, err := mini.DiscoverPublicRoutes(cctx, d.NC, mini.WithDiscoverWait(d.Wait))
	if err != nil {
		if d.Log != nil {
			d.Log.Debug("gate: discover failed", "err", err)
		}
		return
	}
	d.Table.Replace(routes)
	if d.Log != nil {
		d.Log.Info("gate: routes refreshed",
			"count", d.Table.Len(),
			"http", d.Table.HTTPLen(),
		)
	}
}
