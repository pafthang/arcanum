package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/integ/models"
)

// Client talks to the integ service over NATS.
type Client struct {
	c *mini.Client
}

// New creates an integ client.
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

// MustNew creates an integ client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// GetConnector fetches a connector by id, optionally scoped to spaceId.
func (c *Client) GetConnector(ctx context.Context, id, spaceID string) (*models.Connector, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Connector
	in := map[string]any{"id": id}
	if spaceID != "" {
		in["spaceId"] = spaceID
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalIntegConnectorGet, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListConnectors returns connectors in a space.
func (c *Client) ListConnectors(ctx context.Context, spaceID string) ([]models.Connector, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out struct {
		Items []models.Connector `json:"items"`
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalIntegConnectorList, map[string]any{"spaceId": spaceID}, &out, nil); err != nil {
		return nil, err
	}
	if out.Items == nil {
		out.Items = []models.Connector{}
	}
	return out.Items, nil
}
