package proxy

import (
	"net/http"
	"strings"
)

// Hop-by-hop headers that must not be forwarded (RFC 7230).
var hopHeaders = map[string]bool{
	"Connection":          true,
	"Keep-Alive":          true,
	"Proxy-Authenticate":  true,
	"Proxy-Authorization": true,
	"Te":                  true,
	"Trailers":            true,
	"Transfer-Encoding":   true,
	"Upgrade":             true,
}

// CopyHeaders copies headers from src to dst excluding hop-by-hop.
func CopyHeaders(dst, src http.Header) {
	for k, vv := range src {
		if hopHeaders[http.CanonicalHeaderKey(k)] {
			continue
		}
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// SetXForwarded sets/appends X-Forwarded-* headers.
func SetXForwarded(req *http.Request, clientIP string) {
	if prior := req.Header.Get("X-Forwarded-For"); prior != "" {
		req.Header.Set("X-Forwarded-For", prior+", "+clientIP)
	} else {
		req.Header.Set("X-Forwarded-For", clientIP)
	}
	req.Header.Set("X-Forwarded-Host", req.Host)
	if req.TLS != nil {
		req.Header.Set("X-Forwarded-Proto", "https")
	} else {
		req.Header.Set("X-Forwarded-Proto", "http")
	}
}

// RemoveHopHeaders strips hop-by-hop headers.
func RemoveHopHeaders(h http.Header) {
	for k := range hopHeaders {
		h.Del(k)
	}
	// Connection may also list hop headers
	if c := h.Get("Connection"); c != "" {
		for _, f := range strings.Split(c, ",") {
			h.Del(strings.TrimSpace(f))
		}
	}
}
