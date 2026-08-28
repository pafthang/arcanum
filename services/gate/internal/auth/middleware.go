package auth

import (
	"net/http"
	"strings"
)

// AuthMiddlewareOpts configures auth middleware.
type AuthMiddlewareOpts struct {
	Authenticators []Authenticator
	Optional       bool // if true, missing auth is allowed
}

// publicPaths skip authentication (k8s probes, OpenAPI).
func isPublicPath(path string) bool {
	switch path {
	case "/healthz", "/readyz", "/openapi.json":
		return true
	default:
		// metrics path is often /metrics
		return path == "/metrics" || strings.HasPrefix(path, "/metrics/")
	}
}

// AuthMiddleware creates HTTP auth middleware.
func AuthMiddleware(opts AuthMiddlewareOpts) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}
			var lastErr error
			for _, a := range opts.Authenticators {
				id, err := a.Authenticate(r)
				if err == nil {
					if ident, ok := id.(*Identity); ok {
						ctx := WithIdentity(r.Context(), ident)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
				lastErr = err
			}
			if opts.Optional {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			_ = lastErr
		})
	}
}
