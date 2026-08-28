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
