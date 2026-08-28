package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/work/models"
)

// Client talks to the work service over NATS.
type Client struct {
	c *mini.Client
}

// New creates a work client.
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

// MustNew creates a work client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// GetIssue fetches an issue by id, optionally scoped to spaceId.
func (c *Client) GetIssue(ctx context.Context, id, spaceID string) (*models.Issue, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Issue
	in := map[string]any{"id": id}
	if spaceID != "" {
		in["spaceId"] = spaceID
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalWorkIssueGet, in, &out, nil); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListIssues returns issues in a space.
func (c *Client) ListIssues(ctx context.Context, spaceID string) ([]models.Issue, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out struct {
		Items []models.Issue `json:"items"`
	}
	if err := c.c.RequestJSON(ctx, subjects.InternalWorkIssueList, map[string]any{"spaceId": spaceID}, &out, nil); err != nil {
		return nil, err
	}
	if out.Items == nil {
		out.Items = []models.Issue{}
	}
	return out.Items, nil
}

// Overview returns space-scoped issue stats.
func (c *Client) Overview(ctx context.Context, spaceID string) (*models.Overview, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out models.Overview
	if err := c.c.RequestJSON(ctx, subjects.InternalWorkOverview, map[string]any{"spaceId": spaceID}, &out, nil); err != nil {
		return nil, err
	}
	if out.ByStatus == nil {
		out.ByStatus = map[string]int{}
	}
	return &out, nil
}
