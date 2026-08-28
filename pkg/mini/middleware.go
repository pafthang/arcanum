package mini

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

// Middleware wraps a Handler, forming a chain around endpoint logic.
type Middleware func(Handler) Handler

// Chain applies middlewares around h (last middleware is outermost).
// Chain(h, A, B, C) => A(B(C(h))).
func Chain(h Handler, mws ...Middleware) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

// Recover middleware converts panics into a 500 service error response.
func Recover(logg *slog.Logger) Middleware {
	if logg == nil {
		logg = slog.Default()
	}
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logg.Error("handler panic",
						"subject", req.Subject(),
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					_ = req.Error("500", fmt.Sprintf("internal error: %v", rec), nil)
				}
			}()
			next.Handle(req)
		})
	}
}

// Timeout middleware cancels the request context after d.
// Handlers must observe req.Context(); the underlying NATS subscription
// still runs until the handler returns.
func Timeout(d time.Duration) Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			if d <= 0 {
				next.Handle(req)
				return
			}
			ctx, cancel := context.WithTimeout(req.Context(), d)
			defer cancel()
			if r, ok := req.(contextRequest); ok {
				r.setContext(ctx)
			}
			next.Handle(req)
		})
	}
}

// Logging middleware logs request subject and duration.
func Logging(logg *slog.Logger) Middleware {
	if logg == nil {
		logg = slog.Default()
	}
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			start := time.Now()
			next.Handle(req)
			logg.Info("request",
				"subject", req.Subject(),
				"duration", time.Since(start),
			)
		})
	}
}

// contextRequest is implemented by *request to allow middleware to swap ctx.
type contextRequest interface {
	Request
	setContext(context.Context)
}

// RateLimit returns middleware that allows at most limit concurrent in-flight
// handlers (semaphore). Excess requests get error code 429.
// If limit <= 0, the middleware is a no-op.
func RateLimit(limit int) Middleware {
	if limit <= 0 {
		return func(next Handler) Handler { return next }
	}
	sem := make(chan struct{}, limit)
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
				next.Handle(req)
			default:
				_ = req.Error("429", "too many requests", nil)
			}
		})
	}
}

// TokenBucketRateLimit allows rate events/sec with a burst capacity.
// If rate <= 0, no-op. Burst defaults to rate when <= 0.
func TokenBucketRateLimit(rate float64, burst int) Middleware {
	if rate <= 0 {
		return func(next Handler) Handler { return next }
	}
	if burst <= 0 {
		burst = int(rate)
		if burst < 1 {
			burst = 1
		}
	}
	tb := newTokenBucket(rate, burst)
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			if !tb.allow() {
				_ = req.Error("429", "rate limit exceeded", nil)
				return
			}
			next.Handle(req)
		})
	}
}

type tokenBucket struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	tokens  float64
	last    time.Time
	refills atomic.Int64 // diagnostic
}

func newTokenBucket(rate float64, burst int) *tokenBucket {
	return &tokenBucket{
		rate:   rate,
		burst:  float64(burst),
		tokens: float64(burst),
		last:   time.Now(),
	}
}

func (t *tokenBucket) allow() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(t.last).Seconds()
	t.last = now
	t.tokens += elapsed * t.rate
	if t.tokens > t.burst {
		t.tokens = t.burst
	}
	if t.tokens < 1 {
		return false
	}
	t.tokens--
	t.refills.Add(1)
	return true
}

// RequestID middleware ensures X-Request-Id is present on the request headers
// (generates one if missing) for handler logging/correlation.
func RequestID() Middleware {
	return func(next Handler) Handler {
		return HandlerFunc(func(req Request) {
			if req.Headers().Get("X-Request-Id") == "" {
				if r, ok := req.(*request); ok {
					if r.msg.Header == nil {
						r.msg.Header = nats.Header{}
					}
					r.msg.Header.Set("X-Request-Id", newMiniRequestID())
				}
			}
			next.Handle(req)
		})
	}
}

func newMiniRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
