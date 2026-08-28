package models

import "strings"

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

// ExecMachineRequest is POST /exec body. cmd is a string or string array.
type ExecMachineRequest struct {
	Cmd any `json:"cmd"`
}

// CmdParts normalizes cmd into argv.
func (r ExecMachineRequest) CmdParts() []string {
	switch v := r.Cmd.(type) {
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return nil
		}
		return strings.Fields(v)
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, _ := item.(string)
			s = strings.TrimSpace(s)
			if s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}
