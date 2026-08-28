package apis

import (
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/integ/internal/hmacx"
	"github.com/pafthang/arcanum/services/integ/internal/store"
	"github.com/pafthang/arcanum/services/integ/models"
)

func registerWebhooks(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("webhooks_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListWebhooks(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		out := make([]models.Webhook, 0, len(items))
		for i := range items {
			out = append(out, *store.PublicWebhook(&items[i]))
		}
		httpx.JSON(req, 200, map[string]any{"items": out})
	}), mini.Public("GET", "/api/spaces/{spaceId}/integ/webhooks", "integ", "webhook.list")))

	must(svc.AddEndpoint("webhooks_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateWebhookRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		secret := in.Secret
		if secret == "" {
			secret = hmacx.NewSecret()
		}
		active := true
		if in.Active != nil {
			active = *in.Active
		}
		h, err := d.Store.CreateWebhook(req.Context(), spaceID, in.URL, secret, in.Events, active)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, store.PublicWebhook(h))
	}), mini.Public("POST", "/api/spaces/{spaceId}/integ/webhooks", "integ", "webhook.create")))

	must(svc.AddEndpoint("deliveries_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListDeliveries(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/integ/deliveries", "integ", "delivery.list")))
}
