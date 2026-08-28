package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
)

func registerDelete(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("blobs_delete", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		blobID := strings.TrimSpace(mini.PathParam(req, "blobId"))
		if spaceID == "" || blobID == "" {
			httpx.Error(req, 400, "spaceId and blobId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		if err := d.Store.Delete(req.Context(), spaceID, blobID); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 204, map[string]any{"ok": true})
	}), mini.Public("DELETE", "/api/spaces/{spaceId}/blobs/{blobId}", "media", "blob.delete")))
}
