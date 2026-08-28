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
	_, _ = d.NC.Subscribe(subjects.InternalMediaGet, func(msg *nats.Msg) {
		var in struct {
			SpaceID string `json:"spaceId"`
			ID      string `json:"id"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		meta, err := d.Store.GetMeta(context.Background(), in.SpaceID, in.ID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if meta == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, meta)
	})
	_, _ = d.NC.Subscribe(subjects.InternalMediaGetBytes, func(msg *nats.Msg) {
		var in struct {
			SpaceID string `json:"spaceId"`
			ID      string `json:"id"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		meta, data, err := d.Store.ReadBytes(context.Background(), in.SpaceID, in.ID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if meta == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, map[string]any{"blob": meta, "data": data})
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
