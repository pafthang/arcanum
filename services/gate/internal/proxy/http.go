package proxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// HTTPProxy reverse-proxies to an HTTP/HTTPS upstream.
type HTTPProxy struct {
	target  *url.URL
	proxy   *httputil.ReverseProxy
	timeout time.Duration
}

// NewHTTPProxy creates a proxy for target.
func NewHTTPProxy(target string, timeout time.Duration) (*HTTPProxy, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parse target: %w", err)
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	p := &HTTPProxy{
		target:  u,
		timeout: timeout,
	}

	p.proxy = &httputil.ReverseProxy{
		Director: p.director,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			ResponseHeaderTimeout: timeout,
		},
		ErrorHandler: p.errorHandler,
		ModifyResponse: func(resp *http.Response) error {
			RemoveHopHeaders(resp.Header)
			return nil
		},
	}

	return p, nil
}

func (p *HTTPProxy) director(req *http.Request) {
	req.URL.Scheme = p.target.Scheme
	req.URL.Host = p.target.Host
	req.Host = p.target.Host

	clientIP, _, _ := net.SplitHostPort(req.RemoteAddr)
	if clientIP == "" {
		clientIP = req.RemoteAddr
	}
	SetXForwarded(req, clientIP)
	RemoveHopHeaders(req.Header)
}

func (p *HTTPProxy) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, "bad gate: "+err.Error(), http.StatusBadGateway)
}

// ServeHTTP implements http.Handler.
func (p *HTTPProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), p.timeout)
	defer cancel()
	p.proxy.ServeHTTP(w, r.WithContext(ctx))
}

// ProxyRequest performs a single upstream request and returns the response.
func ProxyRequest(ctx context.Context, method, targetURL string, body io.Reader, headers http.Header) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return nil, err
	}
	CopyHeaders(req.Header, headers)
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}
