package apis

import (
	"context"

	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/agents/models"
)

func requireMember(req mini.Request, d *Deps, spaceID string) bool {
	tc := httpx.SpaceContext(req)
	if tc.UserID == "" && !tc.IsPlatform {
		httpx.Error(req, 401, "The request requires valid authorization token.", nil)
		return false
	}
	if tc.IsPlatform {
		return true
	}
	if tc.SpaceID == spaceID {
		return true
	}
	ok, err := isMember(req.Context(), d, spaceID, tc.UserID)
	if err != nil {
		httpx.Error(req, 500, err.Error(), nil)
		return false
	}
	if !ok {
		httpx.Error(req, 403, "Not a member of this space.", nil)
		return false
	}
	return true
}

func isMember(ctx context.Context, d *Deps, spaceID, userID string) (bool, error) {
	if d.Space == nil || userID == "" {
		return false, nil
	}
	items, err := d.Space.ListSpacesForUser(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if item.ID == spaceID {
			return true, nil
		}
	}
	return false, nil
}

func ensureAgent(ctx context.Context, d *Deps, spaceID, agentID string) error {
	if agentID == "" {
		return errMsg("agentId required")
	}
	if d.Space == nil {
		return nil
	}
	u, err := d.Space.GetUser(ctx, agentID)
	if err != nil {
		return err
	}
	if u == nil || u.Actor != "agent" {
		return errMsg("agent principal required")
	}
	ok, err := isMember(ctx, d, spaceID, agentID)
	if err != nil {
		return err
	}
	if !ok {
		return errMsg("agent is not a member of this space")
	}
	return nil
}

func publishRun(d *Deps, subject, typ string, run *models.Run) {
	if d.NC == nil || run == nil {
		return
	}
	_ = events.PublishData(d.NC, subject, typ, "agents", run)
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errMsg(s string) error { return simpleError(s) }

func must(err error) {
	if err != nil {
		panic(err)
	}
}
