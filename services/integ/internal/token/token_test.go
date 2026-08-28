package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestIssueHS256Claims(t *testing.T) {
	secret := []byte("dev-secret-change-me")
	raw, err := Issue(secret, time.Hour, "user-1", "admin@kuayle.local", "user", true, "default", "owner")
	if err != nil {
		t.Fatal(err)
	}
	tok, err := jwt.Parse(raw, func(jt *jwt.Token) (any, error) {
		if jt.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			t.Fatalf("alg %s", jt.Method.Alg())
		}
		return secret, nil
	})
	if err != nil || !tok.Valid {
		t.Fatalf("parse: %v", err)
	}
	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("map claims")
	}
	if claims["sub"] != "user-1" {
		t.Fatalf("sub=%v", claims["sub"])
	}
	if claims["space_id"] != "default" {
		t.Fatalf("space_id=%v", claims["space_id"])
	}
	if claims["space_role"] != "owner" {
		t.Fatalf("space_role=%v", claims["space_role"])
	}
	if claims["actor"] != "user" {
		t.Fatalf("actor=%v", claims["actor"])
	}
	if claims["platform_admin"] != true {
		t.Fatalf("platform_admin=%v", claims["platform_admin"])
	}
}
