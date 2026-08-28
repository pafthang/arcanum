package websocket

import (
	"net/http"
	"strings"
)

// OriginChecker validates the Origin header.
type OriginChecker interface {
	Check(r *http.Request) bool
}

// AllowAllOrigins accepts any origin.
type AllowAllOrigins struct{}

func (AllowAllOrigins) Check(r *http.Request) bool { return true }

// AllowedOrigins — whitelist origins.
type AllowedOrigins struct {
	Origins   map[string]bool
	AllowNull bool
}

// NewAllowedOrigins builds a checker from a list.
func NewAllowedOrigins(origins []string) *AllowedOrigins {
	m := make(map[string]bool, len(origins))
	for _, o := range origins {
		m[strings.TrimSpace(o)] = true
	}
	return &AllowedOrigins{Origins: m}
}

// Check implements OriginChecker.
func (a *AllowedOrigins) Check(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// Some clients omit Origin
		return true
	}
	if origin == "null" && a.AllowNull {
		return true
	}
	return a.Origins[origin]
}
