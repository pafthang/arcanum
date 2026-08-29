package docker

import (
	"context"
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
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Id":"abc123"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1.43/containers/abc123/start":
			started = true
			w.WriteHeader(204)
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/v1.43/containers/abc123/stop"):
			stopped = true
			w.WriteHeader(204)
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := New(srv.URL)
	c.HTTP = srv.Client()
	id, err := c.Start(context.Background(), "Dev Box 1", "debian:bookworm-slim")
	if err != nil || id != "abc123" || !started {
		t.Fatalf("start: %v %s", err, id)
	}
	if err := c.Stop(context.Background(), id); err != nil || !stopped {
		t.Fatal(err)
	}
}

func TestNewEmpty(t *testing.T) {
	if New("") != nil {
		t.Fatal("empty host must be nil")
	}
}

func TestExec(t *testing.T) {
	payload := []byte{1, 0, 0, 0, 0, 0, 0, 5, 'h', 'e', 'l', 'l', 'o'}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1.43/containers/abc123/exec":
			_, _ = w.Write([]byte(`{"Id":"ex1"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1.43/exec/ex1/start":
			_, _ = w.Write(payload)
		case r.Method == http.MethodGet && r.URL.Path == "/v1.43/exec/ex1/json":
			_, _ = w.Write([]byte(`{"ExitCode":0}`))
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()
	c := New(srv.URL)
	c.HTTP = srv.Client()
	out, err := c.Exec(context.Background(), "abc123", []string{"echo", "hello"})
	if err != nil || out.Stdout != "hello" || out.ExitCode != 0 {
		t.Fatalf("%v %+v", err, out)
	}
}
