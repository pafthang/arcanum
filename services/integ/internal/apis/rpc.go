package apis

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/subjects"
)

func registerInternal(d *Deps) {
	if d.NC == nil {
		return
	}
	_, _ = d.NC.Subscribe(subjects.InternalIntegConnectorGet, func(msg *nats.Msg) {
		var in struct {
			ID      string `json:"id"`
			SpaceID string `json:"spaceId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		c, err := d.Store.GetConnector(context.Background(), in.ID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if c == nil || (in.SpaceID != "" && c.SpaceID != in.SpaceID) {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, c)
	})
	_, _ = d.NC.Subscribe(subjects.InternalIntegConnectorList, func(msg *nats.Msg) {
		var in struct {
			SpaceID string `json:"spaceId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		items, err := d.Store.ListConnectors(context.Background(), in.SpaceID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		respondJSON(msg, map[string]any{"items": items})
	})
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
