package admin

import (
	"net/http"
)

// SetupAdmin registers admin endpoints on mux.
func SetupAdmin(mux *http.ServeMux, handlers map[string]http.HandlerFunc) {
	for path, h := range handlers {
		mux.HandleFunc(path, WrapAdmin(h))
	}
}

// WrapAdmin wraps a handler with extra admin logic
// (auth, audit, rate-limit, etc.).
func WrapAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: admin auth / audit
		next(w, r)
	}
}
