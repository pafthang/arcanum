package apis

import (
	"context"
	"strings"

	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	loggmodels "github.com/pafthang/arcanum/services/logg/models"
	"github.com/pafthang/arcanum/services/work/models"
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

func ensureAssignee(ctx context.Context, d *Deps, spaceID, assigneeID string) error {
	assigneeID = strings.TrimSpace(assigneeID)
	if assigneeID == "" {
		return nil
	}
	if d.Space == nil {
		return nil
	}
	ok, err := isMember(ctx, d, spaceID, assigneeID)
	if err != nil {
		return err
	}
	if !ok {
		return errMsg("assignee is not a member of this space")
	}
	return nil
}

func applyIssueExtras(ctx context.Context, d *Deps, spaceID, issueID, priority, dueAt, parentID string, extra, labelIDs []string) error {
	if err := d.Store.SetIssueFields(ctx, issueID, priority, dueAt, parentID); err != nil {
		return err
	}
	for _, uid := range extra {
		if err := ensureAssignee(ctx, d, spaceID, uid); err != nil {
			return err
		}
	}
	if err := d.Store.ReplaceAssignees(ctx, issueID, extra); err != nil {
		return err
	}
	if labelIDs == nil {
		return nil
	}
	return d.Store.SetIssueLabels(ctx, spaceID, issueID, labelIDs)
}

func publishIssue(d *Deps, subject, typ string, iss *models.Issue) {
	publishIssueActor(d, subject, typ, "", iss)
}

func publishIssueActor(d *Deps, subject, typ, actorID string, iss *models.Issue) {
	if iss == nil {
		return
	}
	if d.NC != nil {
		_ = events.PublishData(d.NC, subject, typ, "work", iss)
	}
	if d.Logg == nil {
		return
	}
	d.Logg.AppendActivityAsync(&loggmodels.Activity{
		SpaceID:    iss.SpaceID,
		TargetType: "issue",
		TargetID:   iss.ID,
		ActorID:    actorID,
		Type:       typ,
		Summary:    typ + ": " + iss.Title,
		Payload: map[string]any{
			"status":     iss.Status,
			"assigneeId": iss.AssigneeID,
			"priority":   iss.Priority,
		},
	})
}

type simpleError string

func (e simpleError) Error() string { return string(e) }

func errMsg(s string) error { return simpleError(s) }

func must(err error) {
	if err != nil {
		panic(err)
	}
}
