package apis

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/passwd"
	"github.com/pafthang/arcanum/services/space/internal/store"
	"github.com/pafthang/arcanum/services/space/models"
)

func registerKeys(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("keys_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermWorkspaceManage) {
			return
		}
		items, err := d.Store.ListAPIKeysInSpace(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/keys", "space", "keys.list")))

	must(svc.AddEndpoint("keys_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requirePerm(req, d, spaceID, models.PermWorkspaceManage) {
			return
		}
		var in models.CreateAPIKeyRequest
		_ = httpx.BindJSON(req, &in)
		raw, err := randomKey()
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		hash, err := passwd.Hash(raw)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		email := stringsOr(in.Email, fmt.Sprintf("agent+%s@%s.local", raw[3:11], spaceID))
		u, err := d.Store.CreateUser(req.Context(), email, hash, store.ActorAgent, false)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if _, err := d.Store.AddMember(req.Context(), spaceID, u.ID, store.RoleMember); err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		key, err := d.Store.CreateAPIKey(req.Context(), u.ID, hash)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, models.CreateAPIKeyResponse{
			APIKey: *key,
			User:   u.User,
			Secret: raw,
		})
	}), mini.Public("POST", "/api/spaces/{spaceId}/keys", "space", "keys.create")))
}

func randomKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "ak_" + hex.EncodeToString(b), nil
}

func stringsOr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
