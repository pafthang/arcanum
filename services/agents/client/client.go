package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/agents/models"
)

// Client talks to the agents service over NATS.
type Client struct {
	c *mini.Client
}

// New creates an agents client.
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

// MustNew creates an agents client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// GetRun fetches a run by id, optionally scoped to spaceId.
func (c *Client) GetRun(ctx context.Context, id, spaceID string) (*models.Run, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Run
	in := map[string]any{"id": id}
	if spaceID != "" {
		in["spaceId"] = spaceID
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalAgentsRunGet, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
