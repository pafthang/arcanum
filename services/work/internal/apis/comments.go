package apis

import (
	"strings"

	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	loggmodels "github.com/pafthang/arcanum/services/logg/models"
	"github.com/pafthang/arcanum/services/work/models"
)

func registerComments(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("comments_list", mini.HandlerFunc(func(req mini.Request) {
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
		items, err := d.Store.ListComments(req.Context(), issueID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, map[string]any{"items": items})
	}), mini.Public("GET", "/api/spaces/{spaceId}/issues/{issueId}/comments", "work", "comment.list")))

	must(svc.AddEndpoint("comments_create", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		issueID := strings.TrimSpace(mini.PathParam(req, "issueId"))
		if spaceID == "" || issueID == "" {
			httpx.Error(req, 400, "spaceId and issueId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		tc := httpx.SpaceContext(req)
		if tc.UserID == "" {
			httpx.Error(req, 401, "The request requires valid authorization token.", nil)
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
		var in models.CreateCommentRequest
		if err := httpx.BindJSON(req, &in); err != nil {
			httpx.Error(req, 400, "Invalid body.", nil)
			return
		}
		if !requireBlobInSpace(req, d, spaceID, in.BlobID) {
			return
		}
		c, err := d.Store.AddComment(req.Context(), issueID, tc.UserID, in.Body, in.BlobID)
		if err != nil {
			httpx.Error(req, 400, err.Error(), nil)
			return
		}
		if d.Logg != nil {
			d.Logg.AppendActivityAsync(&loggmodels.Activity{
				SpaceID:    spaceID,
				TargetType: "issue",
				TargetID:   issueID,
				ActorID:    tc.UserID,
				Type:       "issue.commented",
				Summary:    "comment on " + iss.Title,
			})
		}
		httpx.JSON(req, 201, c)
	}), mini.Public("POST", "/api/spaces/{spaceId}/issues/{issueId}/comments", "work", "comment.create")))

	must(svc.AddEndpoint("overview", mini.HandlerFunc(func(req mini.Request) {
		spaceID := httpx.PathSpaceID(req)
		if spaceID == "" {
			httpx.Error(req, 400, "spaceId path required.", nil)
			return
		}
		if !requireMember(req, d, spaceID) {
			return
		}
		ov, err := d.Store.Overview(req.Context(), spaceID)
		if err != nil {
			httpx.Error(req, 500, err.Error(), nil)
			return
		}
		httpx.JSON(req, 200, ov)
	}), mini.Public("GET", "/api/spaces/{spaceId}/work/overview", "work", "overview")))
}
