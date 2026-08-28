package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
)

func requireBlobInSpace(req mini.Request, d *Deps, spaceID, blobID string) bool {
	blobID = strings.TrimSpace(blobID)
	if blobID == "" {
		return true
	}
	if d.Media == nil {
		httpx.Error(req, 503, "media unavailable", nil)
		return false
	}
	meta, err := d.Media.GetMeta(req.Context(), spaceID, blobID)
	if err != nil || meta == nil || meta.ID == "" || meta.SpaceID != spaceID {
		httpx.Error(req, 400, "blob not in this space", nil)
		return false
	}
	return true
}
