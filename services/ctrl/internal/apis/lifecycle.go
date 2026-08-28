package apis

import (
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/ctrl/models"
	loggmodels "github.com/pafthang/arcanum/services/logg/models"
)

func registerLifecycle(svc mini.Service, d *Deps) {
	// Soft reload (reopen DBs / caches) for named services, or all if empty.
	svcutil.Must(svc.AddEndpoint("reload", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		var in models.LifecycleBody
		_ = httpx.BindJSON(req, &in)
		if err := lifecycle.PublishReload(d.NC, in.Services, reasonOr(in.Reason, "ctrl.reload"), in.DelayMs); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		d.recordActivity(loggmodels.Activity{
			SpaceID: "system",
			Type:    "ctrl.reload",
			Summary: "Platform services reloaded: " + reasonOr(in.Reason, "ctrl.reload"),
			Payload: map[string]any{"services": in.Services, "reason": in.Reason},
		})
		httpx.JSON(req, 200, map[string]any{
			"ok":       true,
			"action":   "reload",
			"services": normalizeServices(in.Services),
			"reason":   reasonOr(in.Reason, "ctrl.reload"),
		})
	}),
		mini.WithPublicSubject("ctrl", "reload"),
		mini.WithPublicHTTP("POST", "/api/platform/reload"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// Hard restart: services exit; supervisor (ctrl -up / dev-up / k8s) brings them back.
	svcutil.Must(svc.AddEndpoint("restart", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		var in models.LifecycleBody
		_ = httpx.BindJSON(req, &in)
		if err := lifecycle.PublishRestart(d.NC, in.Services, reasonOr(in.Reason, "ctrl.restart"), in.DelayMs); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		d.recordActivity(loggmodels.Activity{
			SpaceID: "system",
			Type:    "ctrl.restart",
			Summary: "Platform services restarted: " + reasonOr(in.Reason, "ctrl.restart"),
			Payload: map[string]any{"services": in.Services, "reason": in.Reason},
		})
		httpx.JSON(req, 200, map[string]any{
			"ok":       true,
			"action":   "restart",
			"services": normalizeServices(in.Services),
			"reason":   reasonOr(in.Reason, "ctrl.restart"),
			"note":     "Processes exit; supervisor (ctrl -up / dev-up / k8s) must restart them.",
		})
	}),
		mini.WithPublicSubject("ctrl", "restart"),
		mini.WithPublicHTTP("POST", "/api/platform/restart"),
		mini.WithPublicAuth(mini.AuthRequired),
	))
}
