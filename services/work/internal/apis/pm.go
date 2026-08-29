package apis

import (
	"encoding/json"
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/work/models"
)

func registerPM(svc mini.Service, d *Deps) {
	// Cycles
	must(svc.AddEndpoint("cycles_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListCycles(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, items)
	}),
		mini.WithPublicHTTP("GET", "/api/spaces/{spaceId}/work/cycles"),
		mini.WithPublicSubject("work", "cycles.list"),
	))

	must(svc.AddEndpoint("cycles_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateCycleRequest
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		c, err := d.Store.CreateCycle(req.Context(), spaceID, in.Name, in.Description, in.Status, in.StartDate, in.EndDate)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, c)
	}),
		mini.WithPublicHTTP("POST", "/api/spaces/{spaceId}/work/cycles"),
		mini.WithPublicSubject("work", "cycles.create"),
	))

	// Projects
	must(svc.AddEndpoint("projects_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListProjects(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, items)
	}),
		mini.WithPublicHTTP("GET", "/api/spaces/{spaceId}/work/projects"),
		mini.WithPublicSubject("work", "projects.list"),
	))

	must(svc.AddEndpoint("projects_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateProjectRequest
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		p, err := d.Store.CreateProject(req.Context(), spaceID, in.Name, in.Key, in.Description, in.Status, in.LeadID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, p)
	}),
		mini.WithPublicHTTP("POST", "/api/spaces/{spaceId}/work/projects"),
		mini.WithPublicSubject("work", "projects.create"),
	))

	// Saved Views
	must(svc.AddEndpoint("views_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListViews(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, items)
	}),
		mini.WithPublicHTTP("GET", "/api/spaces/{spaceId}/work/views"),
		mini.WithPublicSubject("work", "views.list"),
	))

	must(svc.AddEndpoint("views_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := strings.TrimSpace(mini.PathParam(req, "spaceId"))
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateViewRequest
		if err := json.Unmarshal(req.Data(), &in); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		v, err := d.Store.CreateView(req.Context(), spaceID, in.Name, in.Description, in.Query, in.Icon, req.Subject())
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, v)
	}),
		mini.WithPublicHTTP("POST", "/api/spaces/{spaceId}/work/views"),
		mini.WithPublicSubject("work", "views.create"),
	))
}
