package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/media/models"
)

// Client talks to the media service over NATS.
type Client struct {
	c *mini.Client
}

// New creates a media client.
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

// GetMeta fetches blob metadata.
func (c *Client) GetMeta(ctx context.Context, spaceID, id string) (*models.Blob, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Blob
	if err := c.c.RequestJSON(ctx, subjects.InternalMediaGet, map[string]any{
		"spaceId": spaceID, "id": id,
	}, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetBytes fetches metadata and file bytes.
func (c *Client) GetBytes(ctx context.Context, spaceID, id string) (*models.Blob, []byte, error) {
	if c == nil || c.c == nil {
		return nil, nil, mini.ErrInvalidConnection
	}
	var out struct {
		Blob *models.Blob `json:"blob"`
		Data []byte       `json:"data"`
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalMediaGetBytes, map[string]any{
		"spaceId": spaceID, "id": id,
	}, &out, nil); err != nil {
		return nil, nil, err
	}
	return out.Blob, out.Data, nil
}
