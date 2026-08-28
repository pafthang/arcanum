package apis

import (
	"encoding/json"
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/media/models"
)

func registerPublic(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("blobs_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.List(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/blobs", "media", "blob.list")))

	must(svc.AddEndpoint("blobs_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		filename, contentType, data := readUpload(req)
		if int64(len(data)) > int64(d.Config.MaxBytes) {
			httpx.Error(req, 413, "file too large", nil)
			return
		}
		if len(data) == 0 {
			httpx.Error(req, 400, "empty body", nil)
			return
		}
		actor := httpx.SpaceContext(req).UserID
		b, err := d.Store.Put(req.Context(), spaceID, filename, contentType, actor, data)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, b)
	}), mini.Public("POST", "/api/spaces/{spaceId}/blobs", "media", "blob.create")))

	must(svc.AddEndpoint("blobs_get", mini.HandlerFunc(func(req mini.Request) {
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
		httpx.JSON(req, 200, meta)
	}), mini.Public("GET", "/api/spaces/{spaceId}/blobs/{blobId}", "media", "blob.get")))

	must(svc.AddEndpoint("blobs_download", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		blobID := strings.TrimSpace(mini.PathParam(req, "blobId"))
		if spaceID == "" || blobID == "" {
			httpx.Error(req, 400, "spaceId and blobId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		meta, data, err := d.Store.ReadBytes(req.Context(), spaceID, blobID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if meta == nil {
			httpx.Error(req, 404, "Blob not found.", nil)
			return
		}
		_ = req.Respond(data, mini.WithHeaders(mini.Headers{
			"Content-Type":        []string{meta.ContentType},
			"Content-Disposition": []string{"attachment; filename=\"" + meta.Filename + "\""},
			"X-Blob-Id":           []string{meta.ID},
			"X-Blob-Sha256":       []string{meta.SHA256},
		}))
	}), mini.Public("GET", "/api/spaces/{spaceId}/blobs/{blobId}/content", "media", "blob.content")))
}

func readUpload(req mini.Request) (filename, contentType string, data []byte) {
	filename = mini.Filename(req)
	if filename == "" {
		filename = strings.TrimSpace(httpx.Query(req, "filename"))
	}
	if filename == "" {
		filename = mini.FormValue(req, "filename")
	}
	contentType = req.Headers().Get("Content-Type")
	data = req.Data()
	if strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		var in models.JSONUpload
		if err := json.Unmarshal(data, &in); err == nil && len(in.Data) > 0 {
			if in.Filename != "" {
				filename = in.Filename
			}
			if in.ContentType != "" {
				contentType = in.ContentType
			}
			data = in.Data
		}
	}
	if filename == "" {
		filename = "blob"
	}
	if contentType == "" || strings.HasPrefix(strings.ToLower(contentType), "application/json") {
		contentType = "application/octet-stream"
	}
	return filename, contentType, data
}
