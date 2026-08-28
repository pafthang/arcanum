package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/comms/models"
)

func registerChannels(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("channels_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		teamID := httpx.Query(req, "teamId")
		items, err := d.Store.ListChannels(req.Context(), spaceID, teamID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/channels", "comms", "channel.list")))

	must(svc.AddEndpoint("channels_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateChannelRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		ch, err := d.Store.CreateChannel(req.Context(), spaceID, in.Name, in.TeamID, in.Kind)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		publishChannel(d, ch)
		httpx.JSON(req, 201, ch)
	}), mini.Public("POST", "/api/spaces/{spaceId}/channels", "comms", "channel.create")))

	must(svc.AddEndpoint("channels_get", mini.HandlerFunc(func(req mini.Request) {
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
		httpx.JSON(req, 200, ch)
	}), mini.Public("GET", "/api/spaces/{spaceId}/channels/{channelId}", "comms", "channel.get")))
}
