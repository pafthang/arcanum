package apis

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	loggmodels "github.com/pafthang/arcanum/services/logg/models"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/logg/internal/config"
	"github.com/pafthang/arcanum/services/logg/internal/store"
)

// Deps holds dependencies for the logg service.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Config config.Config
}

// Register registers the HTTP endpoints and NATS subscriptions for the logg service.
func Register(svc mini.Service, d *Deps) {
	registerLogIngestion(d)
	registerActivityInternal(svc, d)
	registerLogListAPI(svc, d)
	registerActivityHTTP(svc, d)
}

// ──────────────────────────────────────────────────────────────────────────────
// Log ingestion via NATS
// ──────────────────────────────────────────────────────────────────────────────

func registerLogIngestion(d *Deps) {
	_, err := d.NC.Subscribe("logs.>", func(msg *nats.Msg) {
		var l map[string]any
		if err := json.Unmarshal(msg.Data, &l); err != nil {
			slog.Warn("logg: failed to parse log message", "error", err)
			return
		}

		timeStr, _ := l["time"].(string)
		level, _ := l["level"].(string)
		message, _ := l["message"].(string)

		serviceName := "unknown"
		if len(msg.Subject) > 5 { // "logs." prefix
			serviceName = msg.Subject[5:]
		}

		if err := d.Store.InsertLog(serviceName, timeStr, level, message, string(msg.Data)); err != nil {
			slog.Warn("logg: failed to insert log", "error", err)
		}
	})
	if err != nil {
		slog.Error("logg: failed to subscribe to logs.>", "error", err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Internal NATS API: activity
// ──────────────────────────────────────────────────────────────────────────────

func registerActivityInternal(svc mini.Service, d *Deps) {
	// Append (fire-and-forget or request/reply)
	_, _ = d.NC.Subscribe(subjects.InternalActivityAppend, func(msg *nats.Msg) {
		var a loggmodels.Activity
		if err := json.Unmarshal(msg.Data, &a); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		if a.ID == "" {
			a.ID = idgen.New()
		}
		if a.Created == "" {
			a.Created = time.Now().UTC().Format(time.RFC3339Nano)
		}
		saved, err := d.Store.AppendActivity(context.Background(), &a)
		if err != nil {
			respondErr(d.NC, msg, "422", err.Error())
			return
		}
		respondJSON(msg, saved)
	})

	// List team activity (paginated + filtered)
	_, _ = d.NC.Subscribe(subjects.InternalActivityList, func(msg *nats.Msg) {
		var in struct {
			SpaceID    string `json:"spaceId"`
			TargetType string `json:"targetType"`
			TargetID   string `json:"targetId"`
			Type       string `json:"type"`
			ActorID    string `json:"actorId"`
			Q          string `json:"q"`
			Page       int    `json:"page"`
			PerPage    int    `json:"perPage"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		items, total, err := d.Store.ListTeamActivity(context.Background(), loggmodels.ActivityListFilter{
			SpaceID:    in.SpaceID,
			TargetType: in.TargetType,
			TargetID:   in.TargetID,
			Type:       in.Type,
			ActorID:    in.ActorID,
			Q:          in.Q,
		}, in.Page, in.PerPage)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		respondJSON(msg, map[string]any{
			"items":      items,
			"totalItems": total,
			"page":       in.Page,
			"perPage":    in.PerPage,
		})
	})

	// List target activity (simple)
	_, _ = d.NC.Subscribe(subjects.InternalActivityListTarget, func(msg *nats.Msg) {
		var in struct {
			SpaceID    string `json:"spaceId"`
			TargetType string `json:"targetType"`
			TargetID   string `json:"targetId"`
			Limit      int    `json:"limit"`
		}
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		items, err := d.Store.ListTargetActivity(context.Background(), in.SpaceID, in.TargetType, in.TargetID, in.Limit)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		respondJSON(msg, map[string]any{
			"items":      items,
			"totalItems": len(items),
		})
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Public HTTP: logs
// ──────────────────────────────────────────────────────────────────────────────

func registerLogListAPI(svc mini.Service, d *Deps) {
	err := svc.AddEndpoint("logs_list", mini.HandlerFunc(func(req mini.Request) {
		page, perPage := httpx.PageParams(req)
		page, perPage = httpx.ClampList(page, perPage, d.Config.ListDefaultPerPage, d.Config.ListMaxPerPage)
		offset := (page - 1) * perPage

		serviceFilter := httpx.Query(req, "service")
		levelFilter := httpx.Query(req, "level")

		logs, err := d.Store.ListLogs(serviceFilter, levelFilter, perPage, offset)
		if err != nil {
			httpx.Error(req, 500, "failed to fetch logs", nil)
			return
		}

		httpx.JSON(req, 200, map[string]any{
			"data":     logs,
			"page":     page,
			"per_page": perPage,
		})
	}),
		mini.WithPublicHTTP("GET", "/api/logs"),
		mini.WithPublicSubject("logg", "logs.list"),
	)
	if err != nil {
		panic(err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Public HTTP: activity (space-scoped)
// ──────────────────────────────────────────────────────────────────────────────

func registerActivityHTTP(svc mini.Service, d *Deps) {
	// GET /api/spaces/{spaceId}/activity
	err := svc.AddEndpoint("activity_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID, ok := httpx.RequireSpacePath(req, false)
		if !ok {
			return
		}
		page, perPage := httpx.PageParams(req)
		page, perPage = httpx.ClampList(page, perPage, d.Config.ListDefaultPerPage, d.Config.ListMaxPerPage)

		items, total, err := d.Store.ListTeamActivity(req.Context(), loggmodels.ActivityListFilter{
			SpaceID:    spaceID,
			TargetType: httpx.Query(req, "targetType"),
			TargetID:   httpx.Query(req, "targetId"),
			Type:       httpx.Query(req, "type"),
			ActorID:    httpx.Query(req, "actorId"),
			Q:          httpx.Query(req, "q"),
		}, page, perPage)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{
			"page": page, "perPage": perPage, "totalItems": total, "items": items,
		})
	}),
		mini.Public("GET", "/api/spaces/{spaceId}/activity", "logg", "activity.list"),
	)
	if err != nil {
		panic(err)
	}

	// POST /api/spaces/{spaceId}/activity
	err = svc.AddEndpoint("activity_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID, ok := httpx.RequireSpacePath(req, true)
		if !ok {
			return
		}
		var in loggmodels.Activity
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		in.SpaceID = spaceID
		if in.ActorID == "" {
			in.ActorID = httpx.SpaceContext(req).UserID
		}
		if in.Type == "" {
			in.Type = "comment"
		}
		if in.ID == "" {
			in.ID = idgen.New()
		}
		if in.Created == "" {
			in.Created = time.Now().UTC().Format(time.RFC3339Nano)
		}

		created, err := d.Store.AppendActivity(req.Context(), &in)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, created)
	}),
		mini.Public("POST", "/api/spaces/{spaceId}/activity", "logg", "activity.create"),
		mini.WithOpenAPISummary("Create a manual activity or comment"),
	)
	if err != nil {
		panic(err)
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Helpers
// ──────────────────────────────────────────────────────────────────────────────

func respondJSON(msg *nats.Msg, v any) {
	b, _ := json.Marshal(v)
	_ = msg.Respond(b)
}

func respondErr(nc *nats.Conn, msg *nats.Msg, code, text string) {
	reply := nats.NewMsg(msg.Reply)
	reply.Header.Set("Nats-Service-Error", text)
	reply.Header.Set("Nats-Service-Error-Code", code)
	reply.Data = []byte(`{}`)
	_ = nc.PublishMsg(reply)
}
