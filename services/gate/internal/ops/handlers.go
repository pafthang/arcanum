package ops

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"
)

// HealthResponse is the /healthz and /readyz body.
type HealthResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

// InfoResponse is extended service info.
type InfoResponse struct {
	Status       string            `json:"status"`
	Version      string            `json:"version"`
	GoVersion    string            `json:"go_version"`
	NumGoroutine int               `json:"num_goroutine"`
	MemStats     map[string]uint64 `json:"mem_stats"`
	Timestamp    time.Time         `json:"timestamp"`
}

// HealthHandler is a simple health check.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	})
}

// ReadyHandler — readiness probe.
func ReadyHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
	})
}

// InfoHandler returns detailed info (admin).
func InfoHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		writeJSON(w, http.StatusOK, InfoResponse{
			Status:       "ok",
			Version:      version,
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			MemStats: map[string]uint64{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"num_gc":      uint64(m.NumGC),
			},
			Timestamp: time.Now().UTC(),
		})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
