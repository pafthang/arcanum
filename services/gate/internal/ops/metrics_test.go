package ops

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type hijackRW struct {
	http.ResponseWriter
	hijacked bool
}

func (h *hijackRW) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h.hijacked = true
	c1, c2 := net.Pipe()
	go c2.Close()
	return c1, bufio.NewReadWriter(bufio.NewReader(c1), bufio.NewWriter(c1)), nil
}

func TestStatusRecorderHijack(t *testing.T) {
	base := httptest.NewRecorder()
	hrw := &hijackRW{ResponseWriter: base}
	rec := &statusRecorder{ResponseWriter: hrw, status: 200}
	if _, ok := any(rec).(http.Hijacker); !ok {
		t.Fatal("statusRecorder must implement Hijacker")
	}
	_, _, err := rec.Hijack()
	if err != nil {
		t.Fatal(err)
	}
	if !hrw.hijacked {
		t.Fatal("underlying not hijacked")
	}
	if rec.status != http.StatusSwitchingProtocols {
		t.Fatalf("status=%d", rec.status)
	}
}
