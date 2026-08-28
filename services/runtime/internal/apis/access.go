package apis

import (
	"context"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
)

func requireMember(req mini.Request, d *Deps, spaceID string) bool {
	tc := httpx.SpaceContext(req)
	if tc.UserID == "" && !tc.IsPlatform {
		httpx.Error(req, 401, "The request requires valid authorization token.", nil)
		return false
	}
	if tc.IsPlatform || tc.SpaceID == spaceID {
		return true
	}
	if d.Space == nil || tc.UserID == "" {
		httpx.Error(req, 403, "Not a member of this space.", nil)
		return false
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
