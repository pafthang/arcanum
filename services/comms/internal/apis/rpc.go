package apis

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/comms/models"
)

func registerInternal(d *Deps) {
	if d.NC == nil {
		return
	}

	_, _ = d.NC.Subscribe(subjects.InternalCommsChannelGet, func(msg *nats.Msg) {
		var in struct {
			ID      string `json:"id"`
			SpaceID string `json:"spaceId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		var (
			ch  *models.Channel
			err error
		)
		if in.SpaceID != "" {
			ch, err = d.Store.GetChannelInSpace(context.Background(), in.SpaceID, in.ID)
		} else {
			ch, err = d.Store.GetChannel(context.Background(), in.ID)
		}
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if ch == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		respondJSON(msg, ch)
	})

	_, _ = d.NC.Subscribe(subjects.InternalCommsChannelList, func(msg *nats.Msg) {
		var in struct {
			SpaceID string `json:"spaceId"`
			TeamID  string `json:"teamId"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		items, err := d.Store.ListChannels(context.Background(), in.SpaceID, in.TeamID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		respondJSON(msg, map[string]any{"items": items})
	})

	_, _ = d.NC.Subscribe(subjects.InternalCommsMessageCreate, func(msg *nats.Msg) {
		var in models.CreateMessageInternal
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		out, err := createInternal(context.Background(), d, in)
		if err != nil {
			respondErr(d.NC, msg, "400", err.Error())
			return
		}
		respondJSON(msg, out)
	})

	_, _ = d.NC.Subscribe(subjects.InternalCommsInbound, func(msg *nats.Msg) {
		var in models.InboundMessage
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		out, err := ingest(context.Background(), d, in)
		if err != nil {
			respondErr(d.NC, msg, "400", err.Error())
			return
		}
		respondJSON(msg, out)
	})

	_, _ = d.NC.Subscribe(subjects.EventIntegMessageInbound, func(msg *nats.Msg) {
		_, in, err := events.Decode[models.InboundMessage](msg.Data)
		if err != nil {
			if json.Unmarshal(msg.Data, &in) != nil {
				return
			}
		}
		_, _ = ingest(context.Background(), d, in)
	})
}

func createInternal(ctx context.Context, d *Deps, in models.CreateMessageInternal) (*models.Message, error) {
	ch, err := resolveChannel(ctx, d, in.SpaceID, in.ChannelID, "", "", "")
	if err != nil {
		return nil, err
	}
	source := strings.TrimSpace(in.Source)
	if source == "" {
		source = models.SourceAgent
	}
	msg, err := d.Store.CreateMessage(ctx, ch.ID, ch.SpaceID, in.ActorID, in.Body, in.ParentID, in.BlobID, source, in.ExternalRef)
	if err != nil {
		return nil, err
	}
	publishMessage(d, msg)
	return msg, nil
}

func ingest(ctx context.Context, d *Deps, in models.InboundMessage) (*models.Message, error) {
	if strings.TrimSpace(in.SpaceID) == "" {
		return nil, errMsg("space_id required")
	}
	if strings.TrimSpace(in.ActorID) == "" {
		return nil, errMsg("actor_id required")
	}
	ch, err := resolveChannel(ctx, d, in.SpaceID, in.ChannelID, in.ChannelName, in.TeamID, in.Kind)
	if err != nil {
		return nil, err
	}
	msg, err := d.Store.CreateMessage(ctx, ch.ID, ch.SpaceID, in.ActorID, in.Body, in.ParentID, in.BlobID, models.SourceInteg, in.ExternalRef)
	if err != nil {
		return nil, err
	}
	publishMessage(d, msg)
	return msg, nil
}

func resolveChannel(ctx context.Context, d *Deps, spaceID, channelID, name, teamID, kind string) (*models.Channel, error) {
	if channelID != "" {
		ch, err := d.Store.GetChannelInSpace(ctx, spaceID, channelID)
		if err != nil {
			return nil, err
		}
		if ch == nil {
			return nil, errMsg("channel not found")
		}
		return ch, nil
	}
	name = firstNonEmpty(name)
	if name == "" {
		return nil, errMsg("channel_id or channel_name required")
	}
	existing, err := d.Store.FindChannelByName(ctx, spaceID, name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil
	}
	ch, err := d.Store.CreateChannel(ctx, spaceID, name, teamID, kind)
	if err != nil {
		return nil, err
	}
	publishChannel(d, ch)
	return ch, nil
}
