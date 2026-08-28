package models

const (
	KindGitHub   = "github"
	KindHook     = "hook"
	KindTelegram = "telegram"
	KindDiscord  = "discord"
	KindSlack    = "slack"
	KindZalo     = "zalo"
	KindZaloOA   = "zalo_oa"
	KindFeishu   = "feishu"
	KindWhatsApp = "whatsapp"

	StatusDisabled = "disabled"
	StatusPending  = "pending"
	StatusActive   = "active"
	StatusError    = "error"

	DeliveryPending = "pending"
	DeliverySent    = "sent"
	DeliveryFailed  = "failed"
)

// KnownKinds is the Kuayle GitHub/hook surface plus GoClaw channels.
var KnownKinds = []string{
	KindGitHub, KindHook, KindTelegram, KindDiscord, KindSlack,
	KindZalo, KindZaloOA, KindFeishu, KindWhatsApp,
}

// ValidKind reports whether kind is a supported connector.
func ValidKind(kind string) bool {
	switch kind {
	case KindGitHub, KindHook, KindTelegram, KindDiscord, KindSlack,
		KindZalo, KindZaloOA, KindFeishu, KindWhatsApp:
		return true
	}
	return false
}

// ChannelKind reports whether kind is a GoClaw messaging channel.
func ChannelKind(kind string) bool {
	switch kind {
	case KindTelegram, KindDiscord, KindSlack, KindZalo, KindZaloOA, KindFeishu, KindWhatsApp:
		return true
	}
	return false
}

// ValidStatus reports connector status.
func ValidStatus(status string) bool {
	switch status {
	case StatusDisabled, StatusPending, StatusActive, StatusError:
		return true
	}
	return false
}

// Connector is an external integration bound to a space.
type Connector struct {
	ID        string         `json:"id"`
	SpaceID   string         `json:"spaceId"`
	Kind      string         `json:"kind"`
	Name      string         `json:"name"`
	Status    string         `json:"status"`
	Config    map[string]any `json:"config,omitempty"`
	HasSecret bool           `json:"hasSecret"`
	Secret    string         `json:"secret,omitempty"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// CreateConnectorRequest is POST .../integ/connectors body.
type CreateConnectorRequest struct {
	Kind   string         `json:"kind"`
	Name   string         `json:"name"`
	Status string         `json:"status"`
	Config map[string]any `json:"config"`
	Secret string         `json:"secret"`
}

// UpdateConnectorRequest is PATCH .../integ/connectors/{id} body.
type UpdateConnectorRequest struct {
	Name   *string         `json:"name"`
	Status *string         `json:"status"`
	Config *map[string]any `json:"config"`
	Secret *string         `json:"secret"`
}

// Repo is a GitHub repository linked to a github connector (Kuayle).
type Repo struct {
	ID             string `json:"id"`
	ConnectorID    string `json:"connectorId"`
	SpaceID        string `json:"spaceId"`
	Owner          string `json:"owner"`
	Name           string `json:"name"`
	InstallationID string `json:"installationId,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

// CreateRepoRequest is POST .../connectors/{id}/repos body.
type CreateRepoRequest struct {
	Owner          string `json:"owner"`
	Name           string `json:"name"`
	InstallationID string `json:"installationId"`
}

// Webhook is an outbound workspace webhook (Kuayle).
type Webhook struct {
	ID        string   `json:"id"`
	SpaceID   string   `json:"spaceId"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
	Active    bool     `json:"active"`
	HasSecret bool     `json:"hasSecret"`
	Secret    string   `json:"secret,omitempty"`
	CreatedAt string   `json:"createdAt"`
}

// CreateWebhookRequest is POST .../integ/webhooks body.
type CreateWebhookRequest struct {
	URL    string   `json:"url"`
	Events []string `json:"events"`
	Secret string   `json:"secret"`
	Active *bool    `json:"active"`
}

// Delivery is one outbound attempt.
type Delivery struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	TargetType  string `json:"targetType"`
	TargetID    string `json:"targetId"`
	EventType   string `json:"eventType"`
	Status      string `json:"status"`
	Attempts    int    `json:"attempts"`
	LastError   string `json:"lastError,omitempty"`
	CreatedAt   string `json:"createdAt"`
	DeliveredAt string `json:"deliveredAt,omitempty"`
}

// InboundEvent is a stored external payload.
type InboundEvent struct {
	ID          string         `json:"id"`
	SpaceID     string         `json:"spaceId"`
	ConnectorID string         `json:"connectorId"`
	Source      string         `json:"source"`
	ExternalID  string         `json:"externalId,omitempty"`
	Payload     map[string]any `json:"payload,omitempty"`
	IssueKeys   []string       `json:"issueKeys,omitempty"`
	CreatedAt   string         `json:"createdAt"`
}

// GitHubActivity is the normalized Kuayle-style GitHub event.
type GitHubActivity struct {
	ConnectorID string   `json:"connectorId"`
	SpaceID     string   `json:"spaceId"`
	Action      string   `json:"action"`
	Repo        string   `json:"repo,omitempty"`
	Ref         string   `json:"ref,omitempty"`
	Title       string   `json:"title,omitempty"`
	IssueKeys   []string `json:"issueKeys,omitempty"`
	ExternalID  string   `json:"externalId,omitempty"`
}

// ChannelMessage is inbound from a GoClaw messaging channel.
type ChannelMessage struct {
	ConnectorID  string `json:"connectorId"`
	SpaceID      string `json:"spaceId"`
	Kind         string `json:"kind"`
	ChatID       string `json:"chatId,omitempty"`
	ExternalUser string `json:"externalUser,omitempty"`
	Text         string `json:"text,omitempty"`
	ExternalID   string `json:"externalId,omitempty"`
}

// IngestRequest is internal ingest / hook body extras.
type IngestRequest struct {
	ConnectorID string            `json:"connectorId"`
	Source      string            `json:"source"`
	Headers     map[string]string `json:"headers,omitempty"`
	Payload     map[string]any    `json:"payload"`
	Raw         []byte            `json:"-"`
}
