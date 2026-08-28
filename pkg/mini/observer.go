package mini

import (
	"context"
	"time"
)

// RequestInfo describes a service request for observers.
type RequestInfo struct {
	Service  string
	Endpoint string
	Subject  string
}

// Observer receives lifecycle events for service requests.
// Implementations must be safe for concurrent use.
// Keep work cheap; heavy export should be async.
type Observer interface {
	// RequestStart is called before the handler. Returned ctx is passed downstream.
	RequestStart(ctx context.Context, info RequestInfo) context.Context
	// RequestEnd is called after the handler returns (or panics are recovered).
	RequestEnd(ctx context.Context, info RequestInfo, err error, d time.Duration)
}

// MultiObserver fans out to multiple observers.
type MultiObserver []Observer

func (m MultiObserver) RequestStart(ctx context.Context, info RequestInfo) context.Context {
	for _, o := range m {
		if o != nil {
			ctx = o.RequestStart(ctx, info)
		}
	}
	return ctx
}

func (m MultiObserver) RequestEnd(ctx context.Context, info RequestInfo, err error, d time.Duration) {
	for _, o := range m {
		if o != nil {
			o.RequestEnd(ctx, info, err, d)
		}
	}
}

// Observe returns middleware that reports handler timing to obs.
func Observe(obs Observer) Middleware {
	return ObserveEndpoint(obs, "", "", "")
}

// ObserveEndpoint is like Observe but fills service/endpoint metadata.
func ObserveEndpoint(obs Observer, service, endpoint, subject string) Middleware {
	if obs == nil {
		return func(next Handler) Handler { return next }
	}
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			info := RequestInfo{
				Service:  service,
				Endpoint: endpoint,
				Subject:  subject,
			}
			if info.Subject == "" {
				info.Subject = req.Subject()
			}
			ctx := obs.RequestStart(req.Context(), info)
			if r, ok := req.(contextRequest); ok {
				r.setContext(ctx)
			}
			start := time.Now()
			next.Handle(req)
			var err error
			if r, ok := req.(*request); ok {
				err = r.respondError
			}
			obs.RequestEnd(ctx, info, err, time.Since(start))
		})
	}
}

// FuncObserver adapts functions to Observer.
type FuncObserver struct {
	Start func(ctx context.Context, info RequestInfo) context.Context
	End   func(ctx context.Context, info RequestInfo, err error, d time.Duration)
}

func (f FuncObserver) RequestStart(ctx context.Context, info RequestInfo) context.Context {
	if f.Start != nil {
		return f.Start(ctx, info)
	}
	return ctx
}

func (f FuncObserver) RequestEnd(ctx context.Context, info RequestInfo, err error, d time.Duration) {
	if f.End != nil {
		f.End(ctx, info, err, d)
	}
}
