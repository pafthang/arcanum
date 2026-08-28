package access

import (
	"strconv"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
)

// Middleware publishes events.access.request after each handler (best-effort).
// Skips when service == "logs" to avoid feedback loops.
func Middleware(nc *nats.Conn, service string) mini.Middleware {
	return func(next mini.Handler) mini.Handler {
		return mini.HandlerFunc(func(req mini.Request) {
			if service == "logs" || nc == nil {
				next.Handle(req)
				return
			}
			cap := &capture{Request: req, status: 200}
			next.Handle(cap)
			Publish(nc, FromRequest(req, cap.status, service))
		})
	}
}

// capture wraps mini.Request to observe response status codes.
type capture struct {
	mini.Request
	status int
}

func (c *capture) Respond(data []byte, opts ...mini.RespondOpt) error {
	c.status = statusFromOpts(200, opts)
	return c.Request.Respond(data, opts...)
}

func (c *capture) RespondJSON(v any, opts ...mini.RespondOpt) error {
	c.status = statusFromOpts(200, opts)
	return c.Request.RespondJSON(v, opts...)
}

func (c *capture) Error(code, description string, data []byte, opts ...mini.RespondOpt) error {
	if n, err := strconv.Atoi(code); err == nil && n >= 100 && n <= 599 {
		c.status = n
	} else {
		c.status = 500
	}
	// WithStatus may still override
	if s := statusFromOpts(c.status, opts); s != c.status && s != 200 {
		c.status = s
	}
	return c.Request.Error(code, description, data, opts...)
}

func statusFromOpts(def int, opts []mini.RespondOpt) int {
	msg := &nats.Msg{Header: nats.Header{}}
	for _, opt := range opts {
		if opt != nil {
			opt(msg)
		}
	}
	if raw := msg.Header.Get(mini.StatusHeader); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			return n
		}
	}
	return def
}

// Ensure capture implements mini.Request (compile-time).
var _ mini.Request = (*capture)(nil)
