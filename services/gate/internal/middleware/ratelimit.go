package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimitOpts configures the rate limiter.
type RateLimitOpts struct {
	RequestsPerSec  float64
	Burst           int
	CleanupInterval time.Duration
	KeyFunc         func(*http.Request) string
	OnReject        func(*http.Request)
}

// RateLimit returns token-bucket rate limiting middleware.
func RateLimit(opts RateLimitOpts) func(http.Handler) http.Handler {
	if opts.RequestsPerSec <= 0 {
		opts.RequestsPerSec = 100
	}
	if opts.Burst <= 0 {
		opts.Burst = int(opts.RequestsPerSec * 2)
	}
	if opts.KeyFunc == nil {
		opts.KeyFunc = clientIP
	}
	if opts.CleanupInterval <= 0 {
		opts.CleanupInterval = 5 * time.Minute
	}

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		ticker := time.NewTicker(opts.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for k, c := range clients {
				if time.Since(c.lastSeen) > 10*time.Minute {
					delete(clients, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := opts.KeyFunc(r)

			mu.Lock()
			c, exists := clients[key]
			if !exists {
				c = &client{
					limiter: rate.NewLimiter(rate.Limit(opts.RequestsPerSec), opts.Burst),
				}
				clients[key] = c
			}
			c.lastSeen = time.Now()
			mu.Unlock()

			if !c.limiter.Allow() {
				if opts.OnReject != nil {
					opts.OnReject(r)
				}
				http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := splitFirst(xff)
		if parts != "" {
			return parts
		}
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func splitFirst(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return trimSpace(s[:i])
		}
	}
	return trimSpace(s)
}

func trimSpace(s string) string {
	start, end := 0, len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
