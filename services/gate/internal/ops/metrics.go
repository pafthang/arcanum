package ops

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

// Metrics is simple in-memory metrics.
type Metrics struct {
	requestsTotal   atomic.Int64
	requestsErrors  atomic.Int64
	requestDuration atomic.Int64
	activeRequests  atomic.Int64
}

// NewMetrics creates Metrics.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Middleware counts requests.
// ResponseWriter preserves Hijacker/Flusher so WebSocket upgrades work.
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.activeRequests.Add(1)
		defer m.activeRequests.Add(-1)

		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		m.requestsTotal.Add(1)
		if rw.status >= 400 {
			m.requestsErrors.Add(1)
		}
		m.requestDuration.Add(time.Since(start).Microseconds())
	})
}

// Handler serves metrics as plain text.
func (m *Metrics) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		total := m.requestsTotal.Load()
		errors := m.requestsErrors.Load()
		duration := m.requestDuration.Load()
		active := m.activeRequests.Load()

		avg := float64(0)
		if total > 0 {
			avg = float64(duration) / float64(total) / 1000.0
		}

		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(
			"# HELP gateway_requests_total Total number of requests\n" +
				"# TYPE gateway_requests_total counter\n" +
				"gateway_requests_total " + strconv.FormatInt(total, 10) + "\n" +
				"# HELP gateway_requests_errors_total Total number of error responses\n" +
				"# TYPE gateway_requests_errors_total counter\n" +
				"gateway_requests_errors_total " + strconv.FormatInt(errors, 10) + "\n" +
				"# HELP gateway_request_duration_ms_avg Average request duration in ms\n" +
				"# TYPE gateway_request_duration_ms_avg gauge\n" +
				"gateway_request_duration_ms_avg " + strconv.FormatFloat(avg, 'f', 3, 64) + "\n" +
				"# HELP gateway_active_requests Current active requests\n" +
				"# TYPE gateway_active_requests gauge\n" +
				"gateway_active_requests " + strconv.FormatInt(active, 10) + "\n",
		))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	// First write implies 200 if WriteHeader was never called.
	return r.ResponseWriter.Write(b)
}

// Hijack enables WebSocket upgrades through the metrics wrapper.
func (r *statusRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("response does not implement http.Hijacker")
	}
	// 101 Switching Protocols once hijacked
	r.status = http.StatusSwitchingProtocols
	return h.Hijack()
}

// Flush delegates if the underlying writer supports it.
func (r *statusRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Unwrap exposes the original ResponseWriter (Go 1.20+ ResponseController).
func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}
