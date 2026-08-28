package objectstore

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// LocalSign mints HMAC query tokens for FS downloads through gate.
func LocalSign(secret, spaceID, blobID string, exp time.Time) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = fmt.Fprintf(mac, "%s|%s|%d", spaceID, blobID, exp.Unix())
	return hex.EncodeToString(mac.Sum(nil))
}

// LocalOK reports whether sig/exp is valid.
func LocalOK(secret, spaceID, blobID, sig, expRaw string) bool {
	expUnix, err := strconv.ParseInt(expRaw, 10, 64)
	if err != nil || expUnix < time.Now().Unix() {
		return false
	}
	want := LocalSign(secret, spaceID, blobID, time.Unix(expUnix, 0))
	got, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	w, err := hex.DecodeString(want)
	if err != nil || len(got) != len(w) {
		return false
	}
	return hmac.Equal(got, w)
}
