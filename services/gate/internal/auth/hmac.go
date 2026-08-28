package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// HMACJWTConfig configures HS256 JWT validation (shared secret with auth service).
type HMACJWTConfig struct {
	Secret   string
	Issuer   string // optional; empty = skip iss check
	Audience string // optional
}

type hmacJWTAuth struct {
	secret   []byte
	issuer   string
	audience string
}

func newHMACJWTAuth(cfg any) (Authenticator, error) {
	c, ok := cfg.(HMACJWTConfig)
	if !ok {
		if m, ok := cfg.(map[string]string); ok {
			c = HMACJWTConfig{Secret: m["secret"], Issuer: m["issuer"], Audience: m["audience"]}
		}
	}
	if strings.TrimSpace(c.Secret) == "" {
		return nil, fmt.Errorf("%w: hmac jwt secret required", ErrUnknownAuthKind)
	}
	return &hmacJWTAuth{
		secret:   []byte(c.Secret),
		issuer:   c.Issuer,
		audience: c.Audience,
	}, nil
}

func (a *hmacJWTAuth) Authenticate(r *http.Request) (any, error) {
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, "Bearer ") {
		return nil, ErrUnauthorized
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	if tokenStr == "" {
		return nil, ErrUnauthorized
	}

	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if t.Method == nil || t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return a.secret, nil
	})
	if err != nil || token == nil || !token.Valid {
		return nil, ErrUnauthorized
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrUnauthorized
	}
	if a.issuer != "" {
		if iss, _ := claims["iss"].(string); iss != a.issuer {
			return nil, ErrUnauthorized
		}
	}
	if a.audience != "" {
		if !audienceOK(claims, a.audience) {
			return nil, ErrUnauthorized
		}
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return nil, ErrUnauthorized
	}
	// materialize string map for header injection
	flat := make(map[string]any, len(claims))
	for k, v := range claims {
		flat[k] = v
	}
	return &Identity{Subject: sub, Claims: flat, Source: "hmac"}, nil
}

func audienceOK(claims jwt.MapClaims, want string) bool {
	switch aud := claims["aud"].(type) {
	case string:
		return aud == want
	case []any:
		for _, a := range aud {
			if s, ok := a.(string); ok && s == want {
				return true
			}
		}
	case []string:
		for _, s := range aud {
			if s == want {
				return true
			}
		}
	}
	return false
}
