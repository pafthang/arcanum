package auth

import (
	"net/http"
	"strings"
)

type bearerAuth struct {
	tokens map[string]string // token → subject
}

func newBearerAuth(cfg any) (Authenticator, error) {
	// cfg expected as map[string]string or a struct
	m, ok := cfg.(map[string]string)
	if !ok || m == nil {
		m = make(map[string]string)
	}
	return &bearerAuth{tokens: m}, nil
}

func (a *bearerAuth) Authenticate(r *http.Request) (any, error) {
	h := r.Header.Get("Authorization")
	if h == "" || !strings.HasPrefix(h, "Bearer ") {
		return nil, ErrUnauthorized
	}
	token := strings.TrimPrefix(h, "Bearer ")
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, ErrUnauthorized
	}
	sub, ok := a.tokens[token]
	if !ok {
		return nil, ErrUnauthorized
	}
	return &Identity{Subject: sub, Source: "bearer"}, nil
}

type apiKeyAuth struct {
	keys map[string]string // key → subject
}

func newAPIKeyAuth(cfg any) (Authenticator, error) {
	m, ok := cfg.(map[string]string)
	if !ok || m == nil {
		m = make(map[string]string)
	}
	return &apiKeyAuth{keys: m}, nil
}

func (a *apiKeyAuth) Authenticate(r *http.Request) (any, error) {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		key = r.URL.Query().Get("api_key")
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrUnauthorized
	}
	sub, ok := a.keys[key]
	if !ok {
		return nil, ErrUnauthorized
	}
	return &Identity{Subject: sub, Source: "apikey"}, nil
}
