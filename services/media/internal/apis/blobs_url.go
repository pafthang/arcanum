package apis

import (
	"fmt"
	"strings"
	"time"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/media/internal/objectstore"
	"github.com/pafthang/arcanum/services/media/models"
)

func registerURL(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("blobs_url", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		blobID := strings.TrimSpace(mini.PathParam(req, "blobId"))
		if spaceID == "" || blobID == "" {
			httpx.Error(req, 400, "spaceId and blobId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		meta, err := d.Store.GetMeta(req.Context(), spaceID, blobID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if meta == nil {
			httpx.Error(req, 404, "Blob not found.", nil)
			return
		}
		exp := time.Now().UTC().Add(d.Config.SignTTL)
		backend := "fs"
		var url string
		if d.Store.Objects != nil {
			backend = d.Store.Objects.Name()
			url, err = d.Store.Objects.PresignGet(req.Context(), spaceID+"/"+blobID, meta.Filename, meta.ContentType, int64(d.Config.SignTTL.Seconds()))
			if err != nil {
				httpx.Error(req, 500, err.Error(), nil)
				return
			}
		}
		if url == "" {
			sig := objectstore.LocalSign(d.Config.SignSecret, spaceID, blobID, exp)
			path := fmt.Sprintf("/api/spaces/%s/blobs/%s/content?exp=%d&sig=%s", spaceID, blobID, exp.Unix(), sig)
			if d.Config.PublicBase != "" {
				url = d.Config.PublicBase + path
			} else {
				url = path
			}
			backend = "fs"
		}
		httpx.JSON(req, 200, models.SignedURL{
			URL:       url,
			ExpiresAt: exp.Format(time.RFC3339),
			Backend:   backend,
		})
	}), mini.Public("GET", "/api/spaces/{spaceId}/blobs/{blobId}/url", "media", "blob.url")))
}
