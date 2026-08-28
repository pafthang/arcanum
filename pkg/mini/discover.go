package mini

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// DiscoverOption configures service discovery.
type DiscoverOption func(*discoverOpts)

type discoverOpts struct {
	wait    time.Duration
	name    string // optional: only discover this service name
	maxMsgs int
}

// WithDiscoverWait sets how long to collect INFO replies (default 300ms).
func WithDiscoverWait(d time.Duration) DiscoverOption {
	return func(o *discoverOpts) {
		if d > 0 {
			o.wait = d
		}
	}
}

// WithDiscoverName limits discovery to a single service name
// ($SRV.INFO.<name>).
func WithDiscoverName(name string) DiscoverOption {
	return func(o *discoverOpts) {
		o.name = name
	}
}

// WithDiscoverMaxMsgs caps the number of INFO replies collected (0 = unlimited).
func WithDiscoverMaxMsgs(n int) DiscoverOption {
	return func(o *discoverOpts) {
		o.maxMsgs = n
	}
}

// DiscoverInfos collects INFO responses from NATS micro services.
// It publishes a request on $SRV.INFO (or $SRV.INFO.<name>) and gathers
// replies until the wait window expires or ctx is done.
func DiscoverInfos(ctx context.Context, nc *nats.Conn, opts ...DiscoverOption) ([]Info, error) {
	if nc == nil {
		return nil, ErrInvalidConnection
	}
	if ctx == nil {
		ctx = context.Background()
	}

	o := discoverOpts{wait: 300 * time.Millisecond}
	for _, opt := range opts {
		opt(&o)
	}

	subj, err := ControlSubject(InfoVerb, o.name, "")
	if err != nil {
		return nil, err
	}

	inbox := nats.NewInbox()
	sub, err := nc.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer sub.Unsubscribe()

	if err := nc.PublishRequest(subj, inbox, nil); err != nil {
		return nil, err
	}
	if err := nc.Flush(); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(o.wait)
	if dl, ok := ctx.Deadline(); ok && dl.Before(deadline) {
		deadline = dl
	}

	var infos []Info
	seen := make(map[string]struct{})

	for {
		if o.maxMsgs > 0 && len(infos) >= o.maxMsgs {
			break
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if err := ctx.Err(); err != nil {
			if len(infos) > 0 {
				break
			}
			return nil, err
		}

		msg, err := sub.NextMsg(remaining)
		if err != nil {
			// timeout is expected when no more replies
			if err == nats.ErrTimeout {
				break
			}
			if len(infos) > 0 {
				break
			}
			return nil, err
		}

		var info Info
		if err := json.Unmarshal(msg.Data, &info); err != nil {
			continue
		}
		if info.Type != "" && info.Type != InfoResponseType {
			continue
		}
		key := info.Name + "/" + info.ID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		infos = append(infos, info)
	}

	return infos, nil
}

// DiscoverPublicRoutes discovers services and returns unique public HTTP routes.
// When multiple instances of the same service advertise the same METHOD+PATH,
// the first one wins (subject should be identical for queue-grouped endpoints).
func DiscoverPublicRoutes(ctx context.Context, nc *nats.Conn, opts ...DiscoverOption) ([]PublicRoute, error) {
	infos, err := DiscoverInfos(ctx, nc, opts...)
	if err != nil {
		return nil, err
	}

	byKey := make(map[string]PublicRoute)
	var order []string
	for _, info := range infos {
		for _, r := range PublicRoutesFromInfo(info) {
			key := r.RouteKey()
			if _, exists := byKey[key]; exists {
				continue
			}
			byKey[key] = r
			order = append(order, key)
		}
	}

	routes := make([]PublicRoute, 0, len(order))
	for _, key := range order {
		routes = append(routes, byKey[key])
	}
	return routes, nil
}

// ErrInvalidConnection is returned when a nil connection is passed to helpers.
var ErrInvalidConnection = fmt.Errorf("%w: invalid connection", ErrConfigValidation)

// AdvertiseSubject is published by services (or operators) to nudge gateways
// to refresh their route tables immediately.
const AdvertiseSubject = "$MINI.GATE.REFRESH"

// AdvertiseRefresh publishes a lightweight nudge so gateways can refresh routes
// without waiting for the next poll interval.
func AdvertiseRefresh(nc *nats.Conn) error {
	if nc == nil {
		return ErrInvalidConnection
	}
	return nc.Publish(AdvertiseSubject, []byte("refresh"))
}

// SubscribeAdvertise registers a handler for AdvertiseSubject. Returns unsubscribe.
func SubscribeAdvertise(nc *nats.Conn, onRefresh func()) (*nats.Subscription, error) {
	if nc == nil {
		return nil, ErrInvalidConnection
	}
	if onRefresh == nil {
		onRefresh = func() {}
	}
	return nc.Subscribe(AdvertiseSubject, func(_ *nats.Msg) {
		onRefresh()
	})
}
