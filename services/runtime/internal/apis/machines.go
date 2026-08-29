package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/runtime/internal/gateway"
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
		m, err := d.Store.Create(req.Context(), spaceID, in.Name, in.Image, in.AgentID, models.StatusRecorded)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if d.Docker != nil {
			cname := "arc-" + m.ID
			if len(cname) > 16 {
				cname = cname[:16]
			}
			id, err := d.Docker.Start(req.Context(), cname, m.Image)
			if err != nil {
				m, _ = d.Store.SetStatus(req.Context(), spaceID, m.ID, models.StatusFailed, "", err.Error())
				httpx.JSON(req, 201, m)
				return
			}
			m, err = d.Store.SetStatus(req.Context(), spaceID, m.ID, models.StatusRunning, id, "")
			if err != nil {
				httpx.Error(req, 500, err.Error(), nil)
				return
			}
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
		if d.Docker != nil && m.DockerID != "" {
			if err := d.Docker.Stop(req.Context(), m.DockerID); err != nil {
				httpx.Error(req, 502, err.Error(), nil)
				return
			}
		}
		m, err = d.Store.SetStatus(req.Context(), spaceID, id, models.StatusStopped, "", "")
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, m)
	}), mini.Public("POST", "/api/spaces/{spaceId}/machines/{machineId}/stop", "runtime", "machine.stop")))

	must(svc.AddEndpoint("machines_exec", mini.HandlerFunc(func(req mini.Request) {
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
		if d.Docker == nil {
			httpx.Error(req, 409, "Docker host is not configured.", nil)
			return
		}
		if m.DockerID == "" {
			httpx.Error(req, 409, "Machine has no container.", nil)
			return
		}
		var in models.ExecMachineRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		cmd := in.CmdParts()
		if len(cmd) == 0 {
			httpx.Error(req, 400, "cmd required.", nil)
			return
		}
		out, err := d.Docker.Exec(req.Context(), m.DockerID, cmd)
		if err != nil {
			httpx.Error(req, 502, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, out)
	}), mini.Public("POST", "/api/spaces/{spaceId}/machines/{machineId}/exec", "runtime", "machine.exec")))

	must(svc.AddEndpoint("machines_proxy", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		id := strings.TrimSpace(mini.PathParam(req, "machineId"))
		port := strings.TrimSpace(mini.PathParam(req, "port"))
		if spaceID == "" || id == "" || port == "" {
			httpx.Error(req, 400, "spaceId, machineId and port required.", nil)
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
		if m.Status != models.StatusRunning {
			httpx.Error(req, 409, "Machine is not running.", nil)
			return
		}
		gw, err := gateway.New("127.0.0.1", port)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{
			"status":    "proxy_ready",
			"machineId": m.ID,
			"port":      port,
			"target":    gw.TargetHost + ":" + gw.Port,
		})
	}), mini.Public("GET", "/api/spaces/{spaceId}/machines/{machineId}/proxy/{port}", "runtime", "machine.proxy")))
}
