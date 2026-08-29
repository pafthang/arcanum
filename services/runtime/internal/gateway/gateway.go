package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// Handler proxies HTTP traffic to a Dev Machine target (IP/Host + Port).
type Handler struct {
	TargetHost string
	Port       string
}

// New creates a machine gateway proxy handler.
func New(targetHost, port string) (*Handler, error) {
	targetHost = strings.TrimSpace(targetHost)
	port = strings.TrimSpace(port)
	if targetHost == "" {
		targetHost = "127.0.0.1"
	}
	if port == "" {
		return nil, fmt.Errorf("port required")
	}
	return &Handler{
		TargetHost: targetHost,
		Port:       port,
	}, nil
}

// ServeHTTP proxies an incoming HTTP request to the target container.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rawURL := fmt.Sprintf("http://%s:%s", h.TargetHost, h.Port)
	target, err := url.Parse(rawURL)
	if err != nil {
		http.Error(w, "invalid proxy target", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Arcanum-DevMachine", h.TargetHost)
	}
	proxy.ServeHTTP(w, r)
}
