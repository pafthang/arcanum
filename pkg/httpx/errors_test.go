package httpx

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorFrom_sentinels(t *testing.T) {
	cases := []struct {
		err  error
		code string
	}{
		{ErrNotFound, "404"},
		{fmt.Errorf("%w: project", ErrNotFound), "404"},
		{ErrConflict, "409"},
		{ErrInvalid, "400"},
		{ErrForbidden, "403"},
		{ErrUnauthorized, "401"},
		{errors.New("boom"), "500"},
		{errors.New("not found"), "404"}, // local service sentinel text
	}
	for _, tc := range cases {
		r := reqWith(nil)
		ErrorFrom(r, tc.err)
		if r.errCode != tc.code {
			t.Fatalf("err=%v: want %s got %s", tc.err, tc.code, r.errCode)
		}
	}
}

func TestErrorFrom_nil(t *testing.T) {
	r := reqWith(nil)
	ErrorFrom(r, nil)
	if r.errCode != "" {
		t.Fatalf("nil should not write error, got %s", r.errCode)
	}
}

type statusErr struct {
	error
	status int
}

func (e statusErr) HTTPStatus() int { return e.status }

func TestErrorFrom_statusError(t *testing.T) {
	r := reqWith(nil)
	ErrorFrom(r, statusErr{error: errors.New("teapot"), status: 418})
	if r.errCode != "418" {
		t.Fatalf("want 418 got %s", r.errCode)
	}
}
