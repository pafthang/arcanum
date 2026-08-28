package apis

import (
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/ctrl/internal/edgecfg"
)

func registerEdgecfg(svc mini.Service, d *Deps) {
	if d.EdgeStore == nil {
		return
	}

	// GET /api/platform/edgecfg/status
	svcutil.Must(svc.AddEndpoint("edgecfg_status", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		ctx := req.Context()
		routes, err := d.EdgeStore.ListRoutes(ctx, true)
		if err != nil {
			httpx.Error(req, 502, err.Error(), nil)
			return
		}
		acls, err := d.EdgeStore.ListWSACL(ctx, true)
		if err != nil {
			httpx.Error(req, 502, err.Error(), nil)
			return
		}
		rev, _ := d.EdgeStore.Revision(ctx)
		enabledRoutes, enabledACL := 0, 0
		for _, rt := range routes {
			if rt.IsEnabled() {
				enabledRoutes++
			}
		}
		for _, a := range acls {
			if a.IsEnabled() {
				enabledACL++
			}
		}
		httpx.JSON(req, 200, map[string]any{
			"bucket":         d.EdgeStore.Bucket(),
			"revision":       rev,
			"routes":         len(routes),
			"routes_enabled": enabledRoutes,
			"wsacl":          len(acls),
			"wsacl_enabled":  enabledACL,
			"schema_version": edgecfg.SchemaVersion,
		})
	}),
		mini.WithPublicSubject("ctrl", "edgecfg.status"),
		mini.WithPublicHTTP("GET", "/api/platform/edgecfg/status"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// GET /api/platform/edgecfg/routes
	svcutil.Must(svc.AddEndpoint("edgecfg_routes_list", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		// include_disabled handled implicitly or via query param (not supported natively in mini query yet, assume true for admin)
		list, err := d.EdgeStore.ListRoutes(req.Context(), true)
		if err != nil {
			httpx.Error(req, 502, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"routes": list})
	}),
		mini.WithPublicSubject("ctrl", "edgecfg.routes.list"),
		mini.WithPublicHTTP("GET", "/api/platform/edgecfg/routes"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// PUT /api/platform/edgecfg/routes/{id}
	svcutil.Must(svc.AddEndpoint("edgecfg_routes_put", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		id := mini.PathParam(req, "id")
		var doc edgecfg.RouteDoc
		if err := httpx.BindJSON(req, &doc); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if doc.ID == "" {
			doc.ID = id
		} else if doc.ID != id {
			httpx.Error(req, 400, "body id does not match path", nil)
			return
		}
		out, err := d.EdgeStore.PutRoute(req.Context(), doc)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, out)
	}),
		mini.WithPublicSubject("ctrl", "edgecfg.routes.put"),
		mini.WithPublicHTTP("PUT", "/api/platform/edgecfg/routes/{id}"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// DELETE /api/platform/edgecfg/routes/{id}
	svcutil.Must(svc.AddEndpoint("edgecfg_routes_delete", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		id := mini.PathParam(req, "id")
		if err := d.EdgeStore.DeleteRoute(req.Context(), id); err != nil {
			if edgecfg.IsNotFound(err) {
				httpx.Error(req, 404, "route not found", nil)
				return
			}
			httpx.Error(req, 502, err.Error(), nil)
			return
		}
		httpx.JSON(req, 204, nil)
	}),
		mini.WithPublicSubject("ctrl", "edgecfg.routes.delete"),
		mini.WithPublicHTTP("DELETE", "/api/platform/edgecfg/routes/{id}"),
		mini.WithPublicAuth(mini.AuthRequired),
	))
}
