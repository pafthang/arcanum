package models

const (
	StatusQueued     = "queued"
	StatusRunning    = "running"
	StatusCancelling = "cancelling"
	StatusCancelled  = "cancelled"
	StatusSucceeded  = "succeeded"
	StatusFailed     = "failed"

	TierWorking  = "working"
	TierEpisodic = "episodic"
	TierSemantic = "semantic"
)

// Run is one agent execution bound to a space (and optionally a work issue).
type Run struct {
	ID         string `json:"id"`
	SpaceID    string `json:"spaceId"`
	IssueID    string `json:"issueId,omitempty"`
	AgentID    string `json:"agentId"`
	Status     string `json:"status"`
	Input      string `json:"input,omitempty"`
	Output     string `json:"output,omitempty"`
	Error      string `json:"error,omitempty"`
	StartedAt  string `json:"startedAt,omitempty"`
	FinishedAt string `json:"finishedAt,omitempty"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

// Session is the durable context of one run (pipeline payload).
type Session struct {
	ID        string `json:"id"`
	RunID     string `json:"runId"`
	SpaceID   string `json:"spaceId"`
	Stage     string `json:"stage"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Memory is long-lived agent memory inside a space.
type Memory struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	AgentID   string `json:"agentId"`
	Tier      string `json:"tier"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	UpdatedAt string `json:"updatedAt"`
}

// Skill is a space-scoped prompt/tool snippet (GoClaw skill surface, stored here).
type Skill struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	Name      string `json:"name"`
	Body      string `json:"body"`
	CreatedAt string `json:"createdAt"`
}

// StartRunRequest is POST /api/spaces/{spaceId}/runs and commands.agents.run.start.
type StartRunRequest struct {
	SpaceID string `json:"spaceId"`
	AgentID string `json:"agentId"`
	IssueID string `json:"issueId"`
	Input   string `json:"input"`
}

// CancelRunRequest is commands.agents.run.cancel.
type CancelRunRequest struct {
	SpaceID string `json:"spaceId"`
	RunID   string `json:"runId"`
}

// PutMemoryRequest is PUT memory body.
type PutMemoryRequest struct {
	Tier  string `json:"tier"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreateSkillRequest is POST skill body.
type CreateSkillRequest struct {
	Name string `json:"name"`
	Body string `json:"body"`
}

// IssueAssigned is the work event payload that may start a run.
type IssueAssigned struct {
	SpaceID    string `json:"spaceId"`
	IssueID    string `json:"issueId"`
	AssigneeID string `json:"assigneeId"`
	Actor      string `json:"actor"`
}
