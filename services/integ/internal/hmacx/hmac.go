package hmacx

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
)

// NewSecret returns 32 random bytes as hex.
func NewSecret() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte("integ-fallback-secret-not-random"))
	}
	return hex.EncodeToString(b[:])
}

// SignHex returns hex HMAC-SHA256(secret, body).
func SignHex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// Equal reports constant-time hex/string equality.
func Equal(got, want string) bool {
	got = strings.TrimSpace(got)
	want = strings.TrimSpace(want)
	if got == "" || want == "" {
		return false
	}
	if len(got) != len(want) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

// VerifyHex checks hex HMAC.
func VerifyHex(secret, signature string, body []byte) bool {
	return Equal(normalizeSig(signature), SignHex(secret, body))
}

// VerifyGitHub checks X-Hub-Signature-256: sha256=<hex>.
func VerifyGitHub(secret, header string, body []byte) bool {
	header = strings.TrimSpace(header)
	header = strings.TrimPrefix(header, "sha256=")
	return VerifyHex(secret, header, body)
}

// VerifyBearer accepts Authorization: Bearer <secret> or raw secret.
func VerifyBearer(secret, header string) bool {
	header = strings.TrimSpace(header)
	header = strings.TrimPrefix(header, "Bearer ")
	header = strings.TrimPrefix(header, "bearer ")
	return Equal(header, secret)
}

func normalizeSig(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "sha256=")
	s = strings.TrimPrefix(s, "v1=")
	return s
}
