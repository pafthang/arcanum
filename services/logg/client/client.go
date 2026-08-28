// Package client provides a typed NATS client for the logg service's activity API.
package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/idgen"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/logg/models"
)

// Client talks to the logg service over NATS.
type Client struct {
	c  *mini.Client
	nc *nats.Conn
}

// New creates a new logg client.
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
	return &Client{c: c, nc: nc}, nil
}

// MustNew creates a new logg client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// AppendActivity records an activity event. Best-effort: returns error only on transport failure.
func (c *Client) AppendActivity(ctx context.Context, a *models.Activity) (*models.Activity, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	if a.ID == "" {
		a.ID = idgen.New()
	}
	if a.Created == "" {
		a.Created = time.Now().UTC().Format(time.RFC3339Nano)
	}
	var out models.Activity
	err := c.c.RequestJSON(ctx, subjects.InternalActivityAppend, a, &out, nil)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

// AppendActivityAsync publishes an activity event as fire-and-forget (no response awaited).
func (c *Client) AppendActivityAsync(a *models.Activity) {
	if c == nil || c.nc == nil {
		return
	}
	if a.ID == "" {
		a.ID = idgen.New()
	}
	if a.Created == "" {
		a.Created = time.Now().UTC().Format(time.RFC3339Nano)
	}
	data, err := json.Marshal(a)
	if err != nil {
		return
	}
	_ = c.nc.Publish(subjects.InternalActivityAppend, data)
}

// ListTeamActivity returns paginated, filtered activity for a space.
func (c *Client) ListTeamActivity(ctx context.Context, spaceID string, f models.ActivityListFilter, page, perPage int) ([]models.Activity, int, error) {
	if c == nil || c.c == nil {
		return nil, 0, mini.ErrInvalidConnection
	}
	var out struct {
		Items      []models.Activity `json:"items"`
		TotalItems int               `json:"totalItems"`
	}
	err := c.c.RequestJSON(ctx, subjects.InternalActivityList, map[string]any{
		"spaceId":    spaceID,
		"targetType": f.TargetType,
		"targetId":   f.TargetID,
		"type":       f.Type,
		"actorId":    f.ActorID,
		"q":          f.Q,
		"page":       page,
		"perPage":    perPage,
	}, &out, nil)
	if err != nil {
		return nil, 0, err
	}
	if out.Items == nil {
		out.Items = []models.Activity{}
	}
	return out.Items, out.TotalItems, nil
}

// ListTargetActivity returns recent activity for a specific target.
func (c *Client) ListTargetActivity(ctx context.Context, spaceID, targetType, targetID string, limit int) ([]models.Activity, error) {
	if c == nil || c.c == nil {
		return nil, mini.ErrInvalidConnection
	}
	var out struct {
		Items []models.Activity `json:"items"`
	}
	err := c.c.RequestJSON(ctx, subjects.InternalActivityListTarget, map[string]any{
		"spaceId":    spaceID,
		"targetType": targetType,
		"targetId":   targetID,
		"limit":      limit,
	}, &out, nil)
	if err != nil {
		return nil, err
	}
	if out.Items == nil {
		out.Items = []models.Activity{}
	}
	return out.Items, nil
}
