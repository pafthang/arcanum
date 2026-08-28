package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/comms/models"
)

// Client talks to the comms service over NATS.
type Client struct {
	c *mini.Client
}

// New creates a comms client.
func New(nc *nats.Conn, opts ...mini.ClientOption) (*Client, error) {
	if nc == nil {
		return nil, mini.ErrInvalidConnection
	}
	if len(opts) == 0 {
		opts = []mini.ClientOption{mini.WithClientTimeout(5 * time.Second)}
	}
	c, err := mini.NewClient(nc, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{c: c}, nil
}

// MustNew creates a comms client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// GetChannel fetches a channel by id, optionally scoped to spaceId.
func (c *Client) GetChannel(ctx context.Context, id, spaceID string) (*models.Channel, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Channel
	in := map[string]any{"id": id}
	if spaceID != "" {
		in["spaceId"] = spaceID
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalCommsChannelGet, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListChannels returns channels in a space.
func (c *Client) ListChannels(ctx context.Context, spaceID, teamID string) ([]models.Channel, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out struct {
		Items []models.Channel `json:"items"`
	}
	in := map[string]any{"spaceId": spaceID}
	if teamID != "" {
		in["teamId"] = teamID
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalCommsChannelList, in, &out, nil); err != nil {
		return nil, err
	}
	if out.Items == nil {
		out.Items = []models.Channel{}
	}
	return out.Items, nil
}

// CreateMessage posts as an agent (or explicit source) without HTTP.
func (c *Client) CreateMessage(ctx context.Context, in models.CreateMessageInternal) (*models.Message, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Message
	if err := c.c.RequestJSON(ctx, subjects.InternalCommsMessageCreate, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// Ingest records an inbound message from integ.
func (c *Client) Ingest(ctx context.Context, in models.InboundMessage) (*models.Message, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Message
	if err := c.c.RequestJSON(ctx, subjects.InternalCommsInbound, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
