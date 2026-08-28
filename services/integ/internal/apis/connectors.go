package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/integ/internal/hmacx"
	"github.com/pafthang/arcanum/services/integ/internal/store"
	"github.com/pafthang/arcanum/services/integ/models"
)

func registerConnectors(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("connectors_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListConnectors(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		out := make([]models.Connector, 0, len(items))
		for i := range items {
			out = append(out, *store.PublicConnector(&items[i]))
		}
		httpx.JSON(req, 200, map[string]any{"items": out})
	}), mini.Public("GET", "/api/spaces/{spaceId}/integ/connectors", "integ", "connector.list")))

	must(svc.AddEndpoint("connectors_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateConnectorRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		secret := strings.TrimSpace(in.Secret)
		if secret == "" {
			secret = hmacx.NewSecret()
		}
		c, err := d.Store.CreateConnector(req.Context(), spaceID, in.Kind, in.Name, in.Status, secret, in.Config)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, store.PublicConnector(c))
	}), mini.Public("POST", "/api/spaces/{spaceId}/integ/connectors", "integ", "connector.create")))

	must(svc.AddEndpoint("connectors_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and connectorId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		c, err := d.Store.GetConnectorInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if c == nil {
			httpx.Error(req, 404, "Connector not found.", nil)
			return
		}
		httpx.JSON(req, 200, store.PublicConnector(c))
	}), mini.Public("GET", "/api/spaces/{spaceId}/integ/connectors/{connectorId}", "integ", "connector.get")))

	must(svc.AddEndpoint("connectors_update", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and connectorId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.UpdateConnectorRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		cur, err := d.Store.GetConnectorInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if cur == nil {
			httpx.Error(req, 404, "Connector not found.", nil)
			return
		}
		c, err := d.Store.UpdateConnector(req.Context(), id, in.Name, in.Status, in.Secret, in.Config)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, store.PublicConnector(c))
	}), mini.Public("PATCH", "/api/spaces/{spaceId}/integ/connectors/{connectorId}", "integ", "connector.update")))

	must(svc.AddEndpoint("repos_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and connectorId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		c, err := d.Store.GetConnectorInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if c == nil {
			httpx.Error(req, 404, "Connector not found.", nil)
			return
		}
		items, err := d.Store.ListRepos(req.Context(), id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/integ/connectors/{connectorId}/repos", "integ", "repo.list")))

	must(svc.AddEndpoint("repos_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and connectorId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		c, err := d.Store.GetConnectorInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if c == nil {
			httpx.Error(req, 404, "Connector not found.", nil)
			return
		}
		if c.Kind != models.KindGitHub {
			httpx.Error(req, 400, "repos require a github connector.", nil)
			return
		}
		var in models.CreateRepoRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		repo, err := d.Store.CreateRepo(req.Context(), id, spaceID, in.Owner, in.Name, in.InstallationID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, repo)
	}), mini.Public("POST", "/api/spaces/{spaceId}/integ/connectors/{connectorId}/repos", "integ", "repo.create")))
}
