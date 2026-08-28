// Package events defines the shared event envelope and helpers.
package events

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// Envelope is the standard message wrapper for domain events and commands.
type Envelope struct {
	Type      string          `json:"type"`
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"ts"`
	Source    string          `json:"source,omitempty"`
	Data      json.RawMessage `json:"data"`
}

// New builds an envelope with a fresh id and timestamp.
func New(typ string, source string, data any) (Envelope, error) {
	raw, err := json.Marshal(data)
	if err != nil {
		return Envelope{}, err
	}
	return Envelope{
		Type:      typ,
		ID:        uuid.NewString(),
		Timestamp: time.Now().UTC(),
		Source:    source,
		Data:      raw,
	}, nil
}

// Publish marshals and publishes an envelope to subject.
func Publish(nc *nats.Conn, subject string, env Envelope) error {
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return nc.Publish(subject, b)
}

// PublishData builds and publishes an envelope in one step.
func PublishData(nc *nats.Conn, subject, typ, source string, data any) error {
	env, err := New(typ, source, data)
	if err != nil {
		return err
	}
	return Publish(nc, subject, env)
}

// Decode unmarshals raw NATS payload into Envelope and typed data.
func Decode[T any](payload []byte) (Envelope, T, error) {
	var env Envelope
	var zero T
	if err := json.Unmarshal(payload, &env); err != nil {
		return env, zero, err
	}
	var data T
	if len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, &data); err != nil {
			return env, zero, err
		}
	}
	return env, data, nil
}

// AccessRequest is published on HTTP/activity logging.
type AccessRequest struct {
	Method    string         `json:"method"`
	URL       string         `json:"url"`
	Status    int            `json:"status"`
	Auth      string         `json:"auth"` // guest|admin|record
	IP        string         `json:"ip"`
	UserAgent string         `json:"user_agent"`
	Referer   string         `json:"referer"`
	Meta      map[string]any `json:"meta,omitempty"`
	Service   string         `json:"service,omitempty"`
}

// LifecycleCommand is published by ops after restore (or admin tools).
type LifecycleCommand struct {
	// Services lists targets; empty or ["*"] means all.
	Services []string `json:"services"`
	Reason   string   `json:"reason,omitempty"`
	// DelayMs before acting (lets ops return HTTP response first).
	DelayMs int `json:"delay_ms,omitempty"`
}
