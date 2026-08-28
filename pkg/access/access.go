// Package access publishes HTTP/activity events for the logs service.
package access

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
)

// FromRequest builds an AccessRequest from a mini request and status.
func FromRequest(req mini.Request, status int, service string) events.AccessRequest {
	auth := "guest"
	if t := req.Headers().Get("X-Mini-Auth-Type"); t != "" {
		auth = t
	} else if req.Headers().Get("X-Mini-Subject") != "" {
		auth = "user"
	}
	meta := map[string]any{}
	if tid := httpx.AuthSpaceID(req); tid != "" {
		meta["space_id"] = tid
	}
	if sub := httpx.AuthSubject(req); sub != "" {
		meta["user_id"] = sub
	}
	if len(meta) == 0 {
		meta = nil
	}
	return events.AccessRequest{
		Method:    first(req.Headers().Get("X-Mini-HTTP-Method"), "NATS"),
		URL:       first(req.Headers().Get("X-Mini-HTTP-Path"), req.Subject()),
		Status:    status,
		Auth:      auth,
		IP:        first(req.Headers().Get("X-Forwarded-For"), req.Headers().Get("X-Real-Ip"), ""),
		UserAgent: req.Headers().Get("User-Agent"),
		Referer:   req.Headers().Get("Referer"),
		Service:   service,
		Meta:      meta,
	}
}

// Publish sends the access event (best-effort, fire-and-forget).
func Publish(nc *nats.Conn, a events.AccessRequest) {
	if nc == nil {
		return
	}
	_ = events.PublishData(nc, subjects.EventAccessRequest, "access.request", a.Service, a)
}

func first(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
