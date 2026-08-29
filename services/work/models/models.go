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
	BlobID    string `json:"blobId,omitempty"`
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
	Body   string `json:"body"`
	BlobID string `json:"blobId"`
}

// Overview is a space-scoped read model over issues (not a separate service).
type Overview struct {
	Issues     int            `json:"issues"`
	ByStatus   map[string]int `json:"byStatus"`
	Assigned   int            `json:"assigned"`
	Unassigned int            `json:"unassigned"`
	Comments   int            `json:"comments"`
}

// Cycle represents a sprint / cycle in a space.
type Cycle struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	StartDate   string `json:"startDate,omitempty"`
	EndDate     string `json:"endDate,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateCycleRequest is POST .../work/cycles body.
type CreateCycleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
}

// Project represents a project in a space.
type Project struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Key         string `json:"key,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	LeadID      string `json:"leadId,omitempty"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

// CreateProjectRequest is POST .../work/projects body.
type CreateProjectRequest struct {
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
	Status      string `json:"status"`
	LeadID      string `json:"leadId"`
}

// View represents a saved view filter in a space.
type View struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Query       string `json:"query,omitempty"`
	Icon        string `json:"icon,omitempty"`
	CreatedBy   string `json:"createdBy,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// CreateViewRequest is POST .../work/views body.
type CreateViewRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Query       string `json:"query"`
	Icon        string `json:"icon"`
}
