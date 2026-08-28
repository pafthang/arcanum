package client

import (
	"context"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/ctrl/models"
)

// Client talks to the ctrl service over NATS.
type Client struct {
	c *mini.Client
}

// New creates a new ctrl client.
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

// MustNew creates a new ctrl client or panics.
func MustNew(nc *nats.Conn, opts ...mini.ClientOption) *Client {
	c, err := New(nc, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// adminHeaders returns headers that simulate a platform admin request for internal service-to-service calls.
func adminHeaders() mini.Headers {
	h := mini.Headers{}
	nats.Header(h).Set("X-Mini-Auth-Type", "admin")
	return h
}

// Reload triggers a soft reload for the specified services.
func (c *Client) Reload(ctx context.Context, body models.LifecycleBody) error {
	var out map[string]any
	return c.c.RequestJSON(ctx, "internal.ctrl.reload", body, &out, adminHeaders())
}

// Restart triggers a hard restart for the specified services.
func (c *Client) Restart(ctx context.Context, body models.LifecycleBody) error {
	var out map[string]any
	return c.c.RequestJSON(ctx, "internal.ctrl.restart", body, &out, adminHeaders())
}

// GetStatus returns the live status of all services.
func (c *Client) GetStatus(ctx context.Context) ([]models.ServiceStatus, error) {
	var out struct {
		Services []models.ServiceStatus `json:"services"`
	}
	err := c.c.RequestJSON(ctx, "internal.ctrl.status", nil, &out, adminHeaders())
	return out.Services, err
}

// GetConfigs returns the declared configuration for all services.
func (c *Client) GetConfigs(ctx context.Context) ([]models.CfgRow, error) {
	var out struct {
		Cfgs []models.CfgRow `json:"cfgs"`
	}
	err := c.c.RequestJSON(ctx, "internal.ctrl.cfgs", nil, &out, adminHeaders())
	return out.Cfgs, err
}
