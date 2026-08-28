package apis

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/agents/models"
)

func registerCommands(d *Deps) {
	if d.NC == nil {
		return
	}
	_, _ = d.NC.Subscribe(subjects.CommandAgentsRunStart, func(msg *nats.Msg) {
		in, err := decodeStart(msg.Data)
		if err != nil {
			return
		}
		_, _ = startRun(context.Background(), d, in.SpaceID, in.AgentID, in.IssueID, in.Input)
	})
	_, _ = d.NC.Subscribe(subjects.CommandAgentsRunCancel, func(msg *nats.Msg) {
		var in models.CancelRunRequest
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			_, in, err = events.Decode[models.CancelRunRequest](msg.Data)
			if err != nil {
				return
			}
		}
		_, _ = cancelRun(context.Background(), d, in.SpaceID, in.RunID)
	})
	_, _ = d.NC.Subscribe(subjects.EventWorkIssueAssigned, func(msg *nats.Msg) {
		_, data, err := events.Decode[map[string]any](msg.Data)
		if err != nil {
			return
		}
		spaceID, _ := data["spaceId"].(string)
		issueID, _ := data["id"].(string)
		if issueID == "" {
			issueID, _ = data["issueId"].(string)
		}
		assignee, _ := data["assigneeId"].(string)
		title, _ := data["title"].(string)
		if spaceID == "" || assignee == "" {
			return
		}
		_, _ = startRun(context.Background(), d, spaceID, assignee, issueID, title)
	})
}

func decodeStart(raw []byte) (models.StartRunRequest, error) {
	var in models.StartRunRequest
	if err := json.Unmarshal(raw, &in); err == nil && in.SpaceID != "" {
		return in, nil
	}
	_, in, err := events.Decode[models.StartRunRequest](raw)
	return in, err
}

func registerInternal(d *Deps) {
	if d.NC == nil {
		return
	}
	_, _ = d.NC.Subscribe(subjects.InternalAgentsRunGet, func(msg *nats.Msg) {
		var in struct {
			ID      string `json:"id"`
			SpaceID string `json:"spaceId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		var (
			run *models.Run
			err error
		)
		if in.SpaceID != "" {
			run, err = d.Store.GetRunInSpace(context.Background(), in.SpaceID, in.ID)
		} else {
			run, err = d.Store.GetRun(context.Background(), in.ID)
		}
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if run == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, run)
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
