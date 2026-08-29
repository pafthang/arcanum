package apis

import (
	"encoding/json"
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/comms/models"
)

func registerReactions(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("reactions_list", mini.HandlerFunc(func(req mini.Request) {
		messageID := strings.TrimSpace(mini.PathParam(req, "messageId"))
		msg, err := d.Store.GetMessage(req.Context(), messageID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if msg == nil {
			httpx.Error(req, 404, "Message not found.", nil)
			return
		}
		if !requireMember(req, d, msg.SpaceID) {
			return
		}
		items, err := d.Store.ListReactions(req.Context(), messageID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, items)
	}),
		mini.WithPublicHTTP("GET", "/api/spaces/{spaceId}/channels/{channelId}/messages/{messageId}/reactions"),
		mini.WithPublicSubject("comms", "reactions.list"),
	))

	must(svc.AddEndpoint("reactions_add", mini.HandlerFunc(func(req mini.Request) {
		messageID := strings.TrimSpace(mini.PathParam(req, "messageId"))
		msg, err := d.Store.GetMessage(req.Context(), messageID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if msg == nil {
			httpx.Error(req, 404, "Message not found.", nil)
			return
		}
		if !requireMember(req, d, msg.SpaceID) {
			return
		}
		var in models.AddReactionRequest
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		actorID := req.Subject()
		r, err := d.Store.AddReaction(req.Context(), messageID, msg.SpaceID, actorID, in.Emoji)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, r)
	}),
		mini.WithPublicHTTP("POST", "/api/spaces/{spaceId}/channels/{channelId}/messages/{messageId}/reactions"),
		mini.WithPublicSubject("comms", "reactions.add"),
	))
}
