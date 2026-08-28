package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/agents/models"
)

func registerMemory(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("memories_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		agentID := strings.TrimSpace(mini.PathParam(req, "agentId"))
		if spaceID == "" || agentID == "" {
			httpx.Error(req, 400, "spaceId and agentId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListMemories(req.Context(), spaceID, agentID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/agents/{agentId}/memories", "agents", "memory.list")))

	must(svc.AddEndpoint("memories_put", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		agentID := strings.TrimSpace(mini.PathParam(req, "agentId"))
		if spaceID == "" || agentID == "" {
			httpx.Error(req, 400, "spaceId and agentId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.PutMemoryRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		m, err := d.Store.PutMemory(req.Context(), spaceID, agentID, in.Tier, in.Key, in.Value)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, m)
	}), mini.Public("PUT", "/api/spaces/{spaceId}/agents/{agentId}/memories", "agents", "memory.put")))

	must(svc.AddEndpoint("skills_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListSkills(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/skills", "agents", "skill.list")))

	must(svc.AddEndpoint("skills_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateSkillRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		sk, err := d.Store.CreateSkill(req.Context(), spaceID, in.Name, in.Body)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, sk)
	}), mini.Public("POST", "/api/spaces/{spaceId}/skills", "agents", "skill.create")))
}
