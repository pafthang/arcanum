package models

// Activity is a durable audit-style event stored centrally.
type Activity struct {
	ID         string         `json:"id"`
	SpaceID    string         `json:"spaceId"`
	TargetType string         `json:"targetType,omitempty"`
	TargetID   string         `json:"targetId,omitempty"`
	ActorID    string         `json:"actorId,omitempty"`
	Type       string         `json:"type"`
	Summary    string         `json:"summary"`
	Payload    map[string]any `json:"payload,omitempty"`
	Created    string         `json:"created"`
}

// ActivityListFilter filters team-wide durable activity.
type ActivityListFilter struct {
	SpaceID    string
	TargetType string // exact targetType (e.g., 'project', 'task', 'user')
	TargetID   string // optional exact targetId
	Type       string // exact type, or prefix like "task" / "project" (matches type or type.*)
	ActorID    string
	Q          string // case-insensitive substring on summary
}
