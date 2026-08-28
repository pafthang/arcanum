package models

const (
	StatusRecorded = "recorded"
	StatusRunning  = "running"
	StatusStopped  = "stopped"
	StatusFailed   = "failed"
)

// Machine is a space-scoped compute record. Docker attach is optional.
type Machine struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	Name      string `json:"name"`
	Image     string `json:"image"`
	Status    string `json:"status"`
	DockerID  string `json:"dockerId,omitempty"`
	AgentID   string `json:"agentId,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// CreateMachineRequest is POST body.
type CreateMachineRequest struct {
	Name    string `json:"name"`
	Image   string `json:"image"`
	AgentID string `json:"agentId"`
}
