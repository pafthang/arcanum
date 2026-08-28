package docker

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStartStop(t *testing.T) {
	var started, stopped bool
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1.43/images/create"):
			w.WriteHeader(200)
		case r.Method == http.MethodPost && r.URL.Path == "/v1.43/containers/create":
			if r.URL.Query().Get("name") == "" {
				t.Error("missing name")
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), `"Image":"debian:bookworm-slim"`) {
				t.Errorf("body %s", body)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Id":"abc123"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1.43/containers/abc123/start":
			started = true
			w.WriteHeader(204)
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1.43/containers/abc123/stop"):
			stopped = true
			w.WriteHeader(204)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	c.HTTP = srv.Client()
	id, err := c.Start(context.Background(), "Dev Box 1", "debian:bookworm-slim")
	if err != nil || id != "abc123" {
		t.Fatalf("start: %v %s", err, id)
	}
	if !started {
		t.Fatal("not started")
	}
	if err := c.Stop(context.Background(), id); err != nil {
		t.Fatal(err)
	}
	if !stopped {
		t.Fatal("not stopped")
	}
}

func TestNewEmpty(t *testing.T) {
	if New("") != nil {
		t.Fatal("empty host must be nil")
	}
}
