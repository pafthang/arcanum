package apis

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/comms/models"
)

func requireMember(req mini.Request, d *Deps, spaceID string) bool {
	tc := httpx.SpaceContext(req)
	if tc.UserID == "" && !tc.IsPlatform {
		httpx.Error(req, 401, "The request requires valid authorization token.", nil)
		return false
	}
	if tc.IsPlatform {
		return true
	}
	if tc.SpaceID == spaceID {
		return true
	}
	ok, err := isMember(req.Context(), d, spaceID, tc.UserID)
	if err != nil {
		httpx.Error(req, 500, err.Error(), nil)
		return false
	}
	if !ok {
		httpx.Error(req, 403, "Not a member of this space.", nil)
		return false
	}
	return true
}

func isMember(ctx context.Context, d *Deps, spaceID, userID string) (bool, error) {
	if d.Space == nil || userID == "" {
		return false, nil
	}
	items, err := d.Space.ListSpacesForUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.ID == spaceID {
			return true, nil
		}
	}
	return false, nil
}

func publishMessage(d *Deps, msg *models.Message) {
	if d.NC == nil || msg == nil {
		return
	}
	_ = events.PublishData(d.NC, subjects.EventCommsMessageCreated, "message.created", "comms", msg)
	_ = events.PublishData(d.NC, subjects.EventCommsChannel(msg.SpaceID, msg.ChannelID), "message.created", "comms", msg)
}

func publishChannel(d *Deps, ch *models.Channel) {
	if d.NC == nil || ch == nil {
		return
	}
	_ = events.PublishData(d.NC, subjects.EventCommsChannelCreated, "channel.created", "comms", ch)
}

func respondJSON(msg *nats.Msg, v any) {
	b, _ := json.Marshal(v)
	_ = msg.Respond(b)
}

func respondErr(nc *nats.Conn, msg *nats.Msg, code, text string) {
	reply := nats.NewMsg(msg.Reply)
	reply.Header.Set("Nats-Service-Error", text)
	reply.Header.Set("Nats-Service-Error-Code", code)
	reply.Data = []byte(`{}`)
	_ = nc.PublishMsg(reply)
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errMsg(s string) error { return simpleError(s) }

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
