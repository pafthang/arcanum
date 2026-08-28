package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims are the JWT claims issued by space login.
type Claims struct {
	SpaceID       string `json:"space_id,omitempty"`
	SpaceRole     string `json:"space_role,omitempty"`
	Actor         string `json:"actor,omitempty"`
	PlatformAdmin bool   `json:"platform_admin"`
	PlatformRole  string `json:"platform_role,omitempty"`
	Typ           string `json:"typ,omitempty"`
	Email         string `json:"email,omitempty"`
	Role          string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

// Issue signs an HS256 token for a user in an optional space.
func Issue(secret []byte, ttl time.Duration, sub, email, actor string, platformAdmin bool, spaceID, spaceRole string) (string, error) {
	if len(secret) == 0 {
		return "", fmt.Errorf("jwt secret required")
	}
	if sub == "" {
		return "", fmt.Errorf("sub required")
	}
	now := time.Now()
	c := Claims{
		SpaceID:       spaceID,
		SpaceRole:     spaceRole,
		Actor:         actor,
		PlatformAdmin: platformAdmin,
		Email:         email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	if platformAdmin {
		c.Typ = "admin"
		c.PlatformRole = "platform_admin"
		c.Role = "platform_admin"
	} else {
		c.Typ = "user"
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString(secret)
}
