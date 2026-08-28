package models

const (
	KindSpace = "space"
	KindTeam  = "team"
	KindDM    = "dm"

	SourceUser  = "user"
	SourceAgent = "agent"
	SourceInteg = "integ"
)

// Channel is a space-scoped conversation (optionally bound to a team).
type Channel struct {
	ID        string `json:"id"`
	SpaceID   string `json:"spaceId"`
	TeamID    string `json:"teamId,omitempty"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// Message is a channel post. Threads use ParentID. Attachments are media blob ids.
type Message struct {
	ID          string `json:"id"`
	ChannelID   string `json:"channelId"`
	SpaceID     string `json:"spaceId"`
	ParentID    string `json:"parentId,omitempty"`
	ActorID     string `json:"actorId"`
	Body        string `json:"body"`
	BlobID      string `json:"blobId,omitempty"`
	Source      string `json:"source"`
	ExternalRef string `json:"externalRef,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// CreateChannelRequest is POST /api/spaces/{spaceId}/channels body.
type CreateChannelRequest struct {
	Name   string `json:"name"`
	TeamID string `json:"teamId"`
	Kind   string `json:"kind"`
}

// CreateMessageRequest is POST .../channels/{channelId}/messages body.
type CreateMessageRequest struct {
	Body     string `json:"body"`
	ParentID string `json:"parentId"`
	BlobID   string `json:"blobId"`
}

// InboundMessage is integ → comms ingest (external messengers are not SoT).
type InboundMessage struct {
	SpaceID     string `json:"spaceId"`
	ChannelID   string `json:"channelId"`
	ChannelName string `json:"channelName"`
	TeamID      string `json:"teamId"`
	Kind        string `json:"kind"`
	ActorID     string `json:"actorId"`
	Body        string `json:"body"`
	BlobID      string `json:"blobId"`
	ParentID    string `json:"parentId"`
	ExternalRef string `json:"externalRef"`
}

// CreateMessageInternal is agents/integ posting without HTTP.
type CreateMessageInternal struct {
	ChannelID   string `json:"channelId"`
	SpaceID     string `json:"spaceId"`
	ActorID     string `json:"actorId"`
	Body        string `json:"body"`
	ParentID    string `json:"parentId"`
	BlobID      string `json:"blobId"`
	Source      string `json:"source"`
	ExternalRef string `json:"externalRef"`
}
