package auth

import (
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwtAuth struct {
	issuer   string
	audience string
	keys     *jwksCache
}

type jwksCache struct {
	mu     sync.RWMutex
	keys   map[string]*rsa.PublicKey
	url    string
	ttl    time.Duration
	loaded time.Time
}

func newJWTAuth(cfg any) (Authenticator, error) {
	// cfg is a map with issuer, audience, jwks_url keys
	c, ok := cfg.(map[string]string)
	if !ok {
		c = make(map[string]string)
	}
	return &jwtAuth{
		issuer:   c["issuer"],
		audience: c["audience"],
		keys: &jwksCache{
			url:  c["jwks_url"],
			ttl:  10 * time.Minute,
			keys: make(map[string]*rsa.PublicKey),
		},
	}, nil
}

func (a *jwtAuth) Authenticate(r *http.Request) (any, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, ErrUnauthorized
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	if tokenStr == "" {
		return nil, ErrUnauthorized
	}

	token, err := jwt.Parse(tokenStr, a.keys.keyFunc)
	if err != nil || !token.Valid {
		return nil, ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}

	// Basic issuer / audience check (when set)
	if a.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != a.issuer {
			return nil, ErrUnauthorized
		}
	}
	if a.audience != "" {
		if aud, _ := claims["aud"].(string); aud != a.audience {
			// aud may be an array — simplified
			return nil, ErrUnauthorized
		}
	}

	sub, _ := claims["sub"].(string)
	return &Identity{
		Subject: sub,
		Claims:  claims,
		Source:  "jwt",
	}, nil
}

func (c *jwksCache) keyFunc(t *jwt.Token) (any, error) {
	kid, _ := t.Header["kid"].(string)
	c.mu.RLock()
	key, ok := c.keys[kid]
	c.mu.RUnlock()
	if ok {
		return key, nil
	}

	// TODO: fetch JWKS from c.url and cache with TTL
	// For now return error — key must be preloaded
	return nil, fmt.Errorf("key not found for kid=%s (JWKS fetch not implemented)", kid)
}

// SetKey manually adds a key (tests / static config).
func (c *jwksCache) SetKey(kid string, key *rsa.PublicKey) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.keys[kid] = key
	c.loaded = time.Now()
}
