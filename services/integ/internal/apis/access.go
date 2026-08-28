package apis

import (
	"context"

	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
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

func publish(d *Deps, subject, typ string, data any) {
	if d.NC == nil || data == nil {
		return
	}
	_ = events.PublishData(d.NC, subject, typ, "integ", data)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
