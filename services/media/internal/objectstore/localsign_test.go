package objectstore

import (
	"strconv"
	"testing"
	"time"
)

func TestLocalSignRoundTrip(t *testing.T) {
	exp := time.Now().Add(5 * time.Minute)
	sig := LocalSign("secret", "sp1", "b1", exp)
	if !LocalOK("secret", "sp1", "b1", sig, strconv.FormatInt(exp.Unix(), 10)) {
		t.Fatal("valid sig rejected")
	}
	if LocalOK("other", "sp1", "b1", sig, strconv.FormatInt(exp.Unix(), 10)) {
		t.Fatal("wrong secret accepted")
	}
	if LocalOK("secret", "sp2", "b1", sig, strconv.FormatInt(exp.Unix(), 10)) {
		t.Fatal("other space accepted")
	}
	past := time.Now().Add(-time.Minute)
	old := LocalSign("secret", "sp1", "b1", past)
	if LocalOK("secret", "sp1", "b1", old, strconv.FormatInt(past.Unix(), 10)) {
		t.Fatal("expired accepted")
	}
}
