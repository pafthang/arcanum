package apis

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	loggmodels "github.com/pafthang/arcanum/services/logg/models"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/ctrl/models"
)

func registerInventory(svc mini.Service, d *Deps) {
	// Live inventory via $SRV.PING (mini monitoring).
	svcutil.Must(svc.AddEndpoint("status", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		wait := 400 * time.Millisecond
		services, err := CollectStats(req.Context(), d.NC, wait)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{
			"ok":       true,
			"count":    len(services),
			"services": services,
		})
	}),
		mini.WithPublicSubject("ctrl", "status"),
		mini.WithPublicHTTP("GET", "/api/platform/status"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// Declared stack from cfgs/ (enabled flags, order, env keys — not secret values).
	svcutil.Must(svc.AddEndpoint("cfgs", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}
		list, err := mini.ListServiceFiles()
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		out := make([]models.CfgRow, 0, len(list))
		for _, c := range list {
			keys := make([]string, 0, len(c.Env))
			for k := range c.Env {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			out = append(out, models.CfgRow{
				Name:        c.Name,
				Description: c.Description,
				Enabled:     c.IsEnabled(),
				Order:       c.StartOrder(),
				EnvKeys:     keys,
				Command:     mini.ResolveCommand(c),
			})
		}
		httpx.JSON(req, 200, map[string]any{
			"ok":      true,
			"cfg_dir": mini.ConfigDir(),
			"count":   len(out),
			"cfgs":    out,
		})
	}),
		mini.WithPublicSubject("ctrl", "cfgs"),
		mini.WithPublicHTTP("GET", "/api/platform/cfgs"),
		mini.WithPublicAuth(mini.AuthRequired),
	))

	// Update a specific service config and trigger a reload.
	svcutil.Must(svc.AddEndpoint("update_cfg", mini.HandlerFunc(func(req mini.Request) {
		if !httpx.RequireAdmin(req, true) {
			return
		}

		name := mini.PathParam(req, "name")
		if name == "" {
			httpx.Error(req, 400, "service name required", nil)
			return
		}

		var in mini.ServiceFileConfig
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}

		// Reload current, merge changes, save back.
		// Wait, if we want to support partial updates, we might need a manual merge.
		// For now, we assume full replacement of the JSON structure sent by the client.

		// Clear name from payload as it is dictated by the path.
		in.Name = name

		if err := mini.SaveServiceFile(name, in); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}

		// Automatically publish a reload for this service.
		_ = lifecycle.PublishReload(d.NC, []string{name}, "ctrl.update_cfg", 0)

		d.recordActivity(loggmodels.Activity{
			SpaceID: "system",
			Type:    "ctrl.cfg.update",
			Summary: "Updated config for service: " + name,
			Payload: map[string]any{"service": name, "cfg": in},
		})

		httpx.JSON(req, 200, map[string]any{
			"ok":       true,
			"service":  name,
			"reloaded": true,
		})
	}),
		mini.WithPublicSubject("ctrl", "update_cfg"),
		mini.WithPublicHTTP("PUT", "/api/platform/cfgs/{name}"),
		mini.WithPublicAuth(mini.AuthRequired),
	))
}

// CollectStats collects mini $SRV.STATS replies for wait duration.
func CollectStats(ctx context.Context, nc *nats.Conn, wait time.Duration) ([]models.ServiceStatus, error) {
	subj, err := mini.ControlSubject(mini.StatsVerb, "", "")
	if err != nil {
		return nil, err
	}
	inbox := nats.NewInbox()
	sub, err := nc.SubscribeSync(inbox)
	if err != nil {
		return nil, err
	}
	defer func() { _ = sub.Unsubscribe() }()

	if err := nc.PublishRequest(subj, inbox, nil); err != nil {
		return nil, err
	}

	deadline := time.Now().Add(wait)
	byID := map[string]models.ServiceStatus{}
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		msg, err := sub.NextMsgWithContext(ctx)
		if err != nil {
			break
		}
		var s mini.Stats
		if err := json.Unmarshal(msg.Data, &s); err != nil {
			continue
		}
		if s.ID == "" {
			continue
		}

		var reqs, errs int
		for _, ep := range s.Endpoints {
			reqs += ep.NumRequests
			errs += ep.NumErrors
		}

		byID[s.ID] = models.ServiceStatus{
			Name:        s.Name,
			ID:          s.ID,
			Version:     s.Version,
			Metadata:    s.Metadata,
			Started:     s.Started.Format(time.RFC3339),
			NumRequests: reqs,
			NumErrors:   errs,
		}
	}
	out := make([]models.ServiceStatus, 0, len(byID))
	for _, s := range byID {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}
