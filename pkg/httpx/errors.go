package httpx

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/pafthang/arcanum/pkg/mini"
)

// Shared domain sentinels for HTTP mapping via ErrorFrom.
// Services may alias: var ErrNotFound = httpx.ErrNotFound
// or wrap: fmt.Errorf("%w: project", httpx.ErrNotFound).
var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalid      = errors.New("invalid")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
)

// StatusError is an optional interface for errors that carry an HTTP status.
type StatusError interface {
	error
	HTTPStatus() int
}

// ErrorFrom maps err to a JSON API error response.
// Known sentinels and StatusError are mapped; everything else is 500.
// No-op when err is nil.
func ErrorFrom(req mini.Request, err error) {
	if err == nil {
		return
	}
	var se StatusError
	if errors.As(err, &se) {
		status := se.HTTPStatus()
		if status < 100 || status > 599 {
			status = 500
		}
		msg := err.Error()
		if status == 404 {
			msg = ""
		}
		Error(req, status, msg, nil)
		return
	}
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, sql.ErrNoRows):
		Error(req, 404, "", nil)
	case errors.Is(err, ErrConflict):
		Error(req, 409, cleanErrMsg(err), nil)
	case errors.Is(err, ErrInvalid):
		Error(req, 400, cleanErrMsg(err), nil)
	case errors.Is(err, ErrForbidden):
		Error(req, 403, cleanErrMsg(err), nil)
	case errors.Is(err, ErrUnauthorized):
		Error(req, 401, cleanErrMsg(err), nil)
	default:
		// Local service sentinels often use the same text; map by suffix message.
		msg := err.Error()
		switch {
		case msg == "not found" || strings.HasSuffix(msg, ": not found"):
			Error(req, 404, "", nil)
		case msg == "conflict" || strings.HasSuffix(msg, ": conflict"):
			Error(req, 409, cleanErrMsg(err), nil)
		case msg == "forbidden" || strings.HasSuffix(msg, ": forbidden"):
			Error(req, 403, cleanErrMsg(err), nil)
		default:
			Error(req, 500, msg, nil)
		}
	}
}

func cleanErrMsg(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// Avoid leaking bare sentinel strings as the only message when empty-ish.
	switch msg {
	case "not found", "conflict", "invalid", "forbidden", "unauthorized":
		return ""
	default:
		return msg
	}
}
