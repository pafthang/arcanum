package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrUnknownAuthKind = errors.New("unknown auth kind")
)

// Authenticator is the auth interface.
type Authenticator interface {
	Authenticate(r *http.Request) (identity any, err error)
}

// NewAuthenticator builds an Authenticator by type.
func NewAuthenticator(kind string, cfg any) (Authenticator, error) {
	switch kind {
	case "bearer":
		return newBearerAuth(cfg)
	case "jwt":
		// Prefer HS256 shared-secret when secret is present (Optima auth).
		if c, ok := cfg.(HMACJWTConfig); ok && strings.TrimSpace(c.Secret) != "" {
			return newHMACJWTAuth(c)
		}
		if m, ok := cfg.(map[string]string); ok && m["secret"] != "" {
			return newHMACJWTAuth(m)
		}
		return newJWTAuth(cfg)
	case "hmac", "hmac-jwt", "hs256":
		return newHMACJWTAuth(cfg)
	case "mtls":
		return newMTLSAuth(cfg)
	case "apikey":
		return newAPIKeyAuth(cfg)
	default:
		return nil, ErrUnknownAuthKind
	}
}

// Identity is a successful auth result.
type Identity struct {
	Subject string
	Claims  map[string]any
	Source  string // bearer / jwt / mtls / apikey
}

type ctxKey struct{}

// WithIdentity stores Identity in context.
func WithIdentity(ctx context.Context, id *Identity) context.Context {
	return context.WithValue(ctx, ctxKey{}, id)
}

// IdentityFromContext loads Identity from context.
func IdentityFromContext(ctx context.Context) (*Identity, bool) {
	id, ok := ctx.Value(ctxKey{}).(*Identity)
	return id, ok
}
