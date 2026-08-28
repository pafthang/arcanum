package passwd

import "testing"

func TestHashVerify(t *testing.T) {
	h, err := Hash("admin")
	if err != nil {
		t.Fatal(err)
	}
	if h == "" || h == "admin" {
		t.Fatalf("placeholder hash: %q", h)
	}
	if !Verify("admin", h) {
		t.Fatal("verify failed for matching password")
	}
	if Verify("nope", h) {
		t.Fatal("verify passed for wrong password")
	}
}

func TestKeyHashStable(t *testing.T) {
	a := KeyHash("ak_abc")
	b := KeyHash("ak_abc")
	if a == "" || a != b || a == "ak_abc" {
		t.Fatalf("key hash %q", a)
	}
	if KeyHash("ak_abd") == a {
		t.Fatal("different secrets collided")
	}
}
