package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/space/models"
)

// Client talks to the space service over NATS.
type Client struct {
	c *mini.Client
}

// New creates a space client.
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

// MustNew creates a space client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// GetSpace fetches a space by id.
func (c *Client) GetSpace(ctx context.Context, id string) (*models.Space, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Space
	if err := c.c.RequestJSON(ctx, subjects.InternalSpaceGet, map[string]any{"id": id}, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListSpacesForUser returns spaces the user belongs to.
func (c *Client) ListSpacesForUser(ctx context.Context, userID string) ([]models.SpaceWithRole, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out struct {
		Items []models.SpaceWithRole `json:"items"`
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalSpaceListForUser, map[string]any{"userId": userID}, &out, nil); err != nil {
		return nil, err
	}
	if out.Items == nil {
		out.Items = []models.SpaceWithRole{}
	}
	return out.Items, nil
}

// GetUser fetches a user by id.
func (c *Client) GetUser(ctx context.Context, id string) (*models.User, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.User
	if err := c.c.RequestJSON(ctx, subjects.InternalSpaceUserGet, map[string]any{"id": id}, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetUserByEmail fetches a user by email.
func (c *Client) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.User
	if err := c.c.RequestJSON(ctx, subjects.InternalSpaceUserGet, map[string]any{"email": email}, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}
