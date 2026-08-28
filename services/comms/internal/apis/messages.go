package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/comms/internal/store"
	"github.com/pafthang/arcanum/services/comms/models"
)

func registerMessages(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("messages_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		channelID := strings.TrimSpace(mini.PathParam(req, "channelId"))
		if spaceID == "" || channelID == "" {
			httpx.Error(req, 400, "spaceId and channelId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		ch, err := d.Store.GetChannelInSpace(req.Context(), spaceID, channelID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if ch == nil {
			httpx.Error(req, 404, "Channel not found.", nil)
			return
		}
		page, perPage := httpx.PageParams(req)
		_, perPage = httpx.ClampList(page, perPage, d.Config.ListDefaultPerPage, d.Config.ListMaxPerPage)
		items, err := d.Store.ListMessages(req.Context(), store.MessageListFilter{
			ChannelID: channelID,
			ParentID:  httpx.Query(req, "parentId"),
			Before:    httpx.Query(req, "before"),
			Limit:     perPage,
		})
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/channels/{channelId}/messages", "comms", "message.list")))

	must(svc.AddEndpoint("messages_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		channelID := strings.TrimSpace(mini.PathParam(req, "channelId"))
		if spaceID == "" || channelID == "" {
			httpx.Error(req, 400, "spaceId and channelId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
			return
		}
		ch, err := d.Store.GetChannelInSpace(req.Context(), spaceID, channelID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if ch == nil {
			httpx.Error(req, 404, "Channel not found.", nil)
			return
		}
		var in models.CreateMessageRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		msg, err := d.Store.CreateMessage(req.Context(), channelID, spaceID, tc.UserID, in.Body, in.ParentID, in.BlobID, models.SourceUser, "")
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		publishMessage(d, msg)
		httpx.JSON(req, 201, msg)
	}), mini.Public("POST", "/api/spaces/{spaceId}/channels/{channelId}/messages", "comms", "message.create")))
}
