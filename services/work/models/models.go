package models

const (
	StatusOpen    = "open"
	StatusStarted = "started"
	StatusDone    = "done"
)

// Issue is a work item living in one space.
type Issue struct {
	ID          string          `json:"id"`
	SpaceID     string          `json:"spaceId"`
	Title       string          `json:"title"`
	Body        string          `json:"body,omitempty"`
	Status      string          `json:"status"`
	AssigneeID  string          `json:"assigneeId,omitempty"`
	AssigneeIDs []string        `json:"assigneeIds,omitempty"`
	Priority    string          `json:"priority,omitempty"`
	DueAt       string          `json:"dueAt,omitempty"`
	ParentID    string          `json:"parentId,omitempty"`
	Relations   []IssueRelation `json:"relations,omitempty"`
	Labels      []Label         `json:"labels,omitempty"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

// IssueRelation links two issues in a space.
type IssueRelation struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	FromID    string `json:"fromId"`
	ToID      string `json:"toId"`
	Kind      string `json:"kind"`
	CreatedAt string `json:"createdAt"`
}

// Label is a space-scoped tag on issues.
type Label struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	Name      string `json:"name"`
	Color     string `json:"color,omitempty"`
	CreatedAt string `json:"createdAt"`
}

// Comment is part of the issue aggregate, not a comms channel.
type Comment struct {
	ID        string `json:"id"`
	IssueID   string `json:"issueId"`
	ActorID   string `json:"actorId"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// CreateIssueRequest is POST /api/spaces/{spaceId}/issues body.
type CreateIssueRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	Status      string   `json:"status"`
	AssigneeID  string   `json:"assigneeId"`
	AssigneeIDs []string `json:"assigneeIds"`
	Priority    string   `json:"priority"`
	DueAt       string   `json:"dueAt"`
	ParentID    string   `json:"parentId"`
	LabelIDs    []string `json:"labelIds"`
}

// UpdateIssueRequest is PATCH /api/spaces/{spaceId}/issues/{issueId} body.
type UpdateIssueRequest struct {
	Title       *string  `json:"title"`
	Body        *string  `json:"body"`
	Status      *string  `json:"status"`
	AssigneeID  *string  `json:"assigneeId"`
	AssigneeIDs []string `json:"assigneeIds"`
	Priority    *string  `json:"priority"`
	DueAt       *string  `json:"dueAt"`
	ParentID    *string  `json:"parentId"`
	LabelIDs    []string `json:"labelIds"`
}

// CreateRelationRequest is POST .../issues/{issueId}/relations body.
type CreateRelationRequest struct {
	ToID string `json:"toId"`
	Kind string `json:"kind"`
}

// CreateLabelRequest is POST /api/spaces/{spaceId}/labels body.
type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// CreateCommentRequest is POST .../issues/{issueId}/comments body.
type CreateCommentRequest struct {
	Body string `json:"body"`
}

// Overview is a space-scoped read model over issues (not a separate service).
type Overview struct {
	Issues     int            `json:"issues"`
	ByStatus   map[string]int `json:"byStatus"`
	Assigned   int            `json:"assigned"`
	Unassigned int            `json:"unassigned"`
	Comments   int            `json:"comments"`
}
