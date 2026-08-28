package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/runtime/models"
)

func registerPublic(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("machines_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.List(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/machines", "runtime", "machine.list")))

	must(svc.AddEndpoint("machines_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateMachineRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		status := models.StatusRecorded
		if d.Config.HasDocker() {
			status = models.StatusRunning
		}
		m, err := d.Store.Create(req.Context(), spaceID, in.Name, in.Image, in.AgentID, status)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, m)
	}), mini.Public("POST", "/api/spaces/{spaceId}/machines", "runtime", "machine.create")))

	must(svc.AddEndpoint("machines_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "machineId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and machineId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		m, err := d.Store.GetInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if m == nil {
			httpx.Error(req, 404, "Machine not found.", nil)
			return
		}
		httpx.JSON(req, 200, m)
	}), mini.Public("GET", "/api/spaces/{spaceId}/machines/{machineId}", "runtime", "machine.get")))

	must(svc.AddEndpoint("machines_stop", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "machineId"))
		if spaceID == "" || id == "" {
			httpx.Error(req, 400, "spaceId and machineId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		m, err := d.Store.GetInSpace(req.Context(), spaceID, id)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if m == nil {
			httpx.Error(req, 404, "Machine not found.", nil)
			return
		}
		m, err = d.Store.SetStatus(req.Context(), spaceID, id, models.StatusStopped, "", "")
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, m)
	}), mini.Public("POST", "/api/spaces/{spaceId}/machines/{machineId}/stop", "runtime", "machine.stop")))
}
