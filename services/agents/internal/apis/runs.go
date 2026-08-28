package apis

import (
	"context"
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/agents/models"
)

func registerRuns(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("runs_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListRuns(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/runs", "agents", "run.list")))

	must(svc.AddEndpoint("runs_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.StartRunRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		run, err := startRun(req.Context(), d, spaceID, in.AgentID, in.IssueID, in.Input)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, run)
	}), mini.Public("POST", "/api/spaces/{spaceId}/runs", "agents", "run.create")))

	must(svc.AddEndpoint("runs_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		runID := strings.TrimSpace(mini.PathParam(req, "runId"))
		if spaceID == "" || runID == "" {
			httpx.Error(req, 400, "spaceId and runId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		run, err := d.Store.GetRunInSpace(req.Context(), spaceID, runID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if run == nil {
			httpx.Error(req, 404, "Run not found.", nil)
			return
		}
		httpx.JSON(req, 200, run)
	}), mini.Public("GET", "/api/spaces/{spaceId}/runs/{runId}", "agents", "run.get")))

	must(svc.AddEndpoint("runs_cancel", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		runID := strings.TrimSpace(mini.PathParam(req, "runId"))
		if spaceID == "" || runID == "" {
			httpx.Error(req, 400, "spaceId and runId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		run, err := cancelRun(req.Context(), d, spaceID, runID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if run == nil {
			httpx.Error(req, 404, "Run not found.", nil)
			return
		}
		httpx.JSON(req, 200, run)
	}), mini.Public("POST", "/api/spaces/{spaceId}/runs/{runId}/cancel", "agents", "run.cancel")))

	must(svc.AddEndpoint("session_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		runID := strings.TrimSpace(mini.PathParam(req, "runId"))
		if spaceID == "" || runID == "" {
			httpx.Error(req, 400, "spaceId and runId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		run, err := d.Store.GetRunInSpace(req.Context(), spaceID, runID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if run == nil {
			httpx.Error(req, 404, "Run not found.", nil)
			return
		}
		sess, err := d.Store.GetSession(req.Context(), runID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if sess == nil {
			httpx.Error(req, 404, "Session not found.", nil)
			return
		}
		httpx.JSON(req, 200, sess)
	}), mini.Public("GET", "/api/spaces/{spaceId}/runs/{runId}/session", "agents", "session.get")))
}

func startRun(ctx context.Context, d *Deps, spaceID, agentID, issueID, input string) (*models.Run, error) {
	if err := ensureAgent(ctx, d, spaceID, strings.TrimSpace(agentID)); err != nil {
		return nil, err
	}
	run, err := d.Store.CreateRun(ctx, spaceID, agentID, issueID, input)
	if err != nil {
		return nil, err
	}
	publishRun(d, subjects.EventAgentsRunStarted, "run.started", run)
	if d.Runner == nil {
		return run, nil
	}
	finished, err := d.Runner.Execute(ctx, run)
	if err != nil {
		return nil, err
	}
	publishRun(d, subjects.EventAgentsRunFinished, "run.finished", finished)
	return finished, nil
}

func cancelRun(ctx context.Context, d *Deps, spaceID, runID string) (*models.Run, error) {
	cur, err := d.Store.GetRunInSpace(ctx, spaceID, runID)
	if err != nil || cur == nil {
		return cur, err
	}
	return d.Store.CancelRun(ctx, cur.ID)
}
