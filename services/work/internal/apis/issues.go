package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/work/models"
)

func registerIssues(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("issues_list", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		items, err := d.Store.ListIssues(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/issues", "work", "issue.list")))

	must(svc.AddEndpoint("issues_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.CreateIssueRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		if err := ensureAssignee(req.Context(), d, spaceID, in.AssigneeID); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		iss, err := d.Store.CreateIssue(req.Context(), spaceID, in.Title, in.Body, in.Status, in.AssigneeID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if err := applyIssueExtras(req.Context(), d, spaceID, iss.ID, in.Priority, in.DueAt, in.ParentID, in.AssigneeIDs); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		iss, _ = d.Store.GetIssueInSpace(req.Context(), spaceID, iss.ID)
		publishIssue(d, subjects.EventWorkIssueCreated, "issue.created", iss)
		if iss.AssigneeID != "" {
			publishIssue(d, subjects.EventWorkIssueAssigned, "issue.assigned", iss)
		}
		httpx.JSON(req, 201, iss)
	}), mini.Public("POST", "/api/spaces/{spaceId}/issues", "work", "issue.create")))

	must(svc.AddEndpoint("issues_get", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		issueID := strings.TrimSpace(mini.PathParam(req, "issueId"))
		if spaceID == "" || issueID == "" {
			httpx.Error(req, 400, "spaceId and issueId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		iss, err := d.Store.GetIssueInSpace(req.Context(), spaceID, issueID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if iss == nil {
			httpx.Error(req, 404, "Issue not found.", nil)
			return
		}
		httpx.JSON(req, 200, iss)
	}), mini.Public("GET", "/api/spaces/{spaceId}/issues/{issueId}", "work", "issue.get")))

	must(svc.AddEndpoint("issues_update", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		issueID := strings.TrimSpace(mini.PathParam(req, "issueId"))
		if spaceID == "" || issueID == "" {
			httpx.Error(req, 400, "spaceId and issueId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		var in models.UpdateIssueRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		cur, err := d.Store.GetIssueInSpace(req.Context(), spaceID, issueID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		if cur == nil {
			httpx.Error(req, 404, "Issue not found.", nil)
			return
		}
		if in.AssigneeID != nil {
			if err := ensureAssignee(req.Context(), d, spaceID, *in.AssigneeID); err != nil {
				httpx.Error(req, 400, err.Error(), nil)
				return
			}
		}
		iss, err := d.Store.UpdateIssue(req.Context(), issueID, in.Title, in.Body, in.Status, in.AssigneeID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		priority, dueAt, parent := iss.Priority, iss.DueAt, iss.ParentID
		if in.Priority != nil {
			priority = *in.Priority
		}
		if in.DueAt != nil {
			dueAt = *in.DueAt
		}
		if in.ParentID != nil {
			parent = *in.ParentID
		}
		if err := applyIssueExtras(req.Context(), d, spaceID, iss.ID, priority, dueAt, parent, in.AssigneeIDs); err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		iss, _ = d.Store.GetIssueInSpace(req.Context(), spaceID, issueID)
		publishIssue(d, subjects.EventWorkIssueUpdated, "issue.updated", iss)
		if in.AssigneeID != nil && strings.TrimSpace(*in.AssigneeID) != "" && *in.AssigneeID != cur.AssigneeID {
			publishIssue(d, subjects.EventWorkIssueAssigned, "issue.assigned", iss)
		}
		httpx.JSON(req, 200, iss)
	}), mini.Public("PATCH", "/api/spaces/{spaceId}/issues/{issueId}", "work", "issue.update")))

	must(svc.AddEndpoint("issue_relations_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		issueID := strings.TrimSpace(mini.PathParam(req, "issueId"))
		if spaceID == "" || issueID == "" {
			httpx.Error(req, 400, "spaceId and issueId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		cur, err := d.Store.GetIssueInSpace(req.Context(), spaceID, issueID)
		if err != nil || cur == nil {
			httpx.Error(req, 404, "Issue not found.", nil)
			return
		}
		var in models.CreateRelationRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		other, err := d.Store.GetIssueInSpace(req.Context(), spaceID, in.ToID)
		if err != nil || other == nil {
			httpx.Error(req, 400, "target issue not in this space.", nil)
			return
		}
		rel, err := d.Store.AddRelation(req.Context(), spaceID, issueID, in.ToID, in.Kind)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		httpx.JSON(req, 201, rel)
	}), mini.Public("POST", "/api/spaces/{spaceId}/issues/{issueId}/relations", "work", "issue.relation.create")))
}
