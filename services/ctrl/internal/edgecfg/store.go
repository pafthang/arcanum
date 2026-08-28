package edgecfg

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Store is the NATS KV-backed config plane.
type Store struct {
	nc     *nats.Conn
	js     jetstream.JetStream
	kv     jetstream.KeyValue
	bucket string
}

// Open connects to (or creates) the config plane KV bucket.
func Open(ctx context.Context, nc *nats.Conn, bucket string) (*Store, error) {
	if nc == nil {
		return nil, errors.New("config: nil connection")
	}
	if bucket == "" {
		bucket = DefaultBucket
	}
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, err
	}
	kv, err := js.CreateOrUpdateKeyValue(ctx, jetstream.KeyValueConfig{
		Bucket:      bucket,
		Description: "mini config plane",
		History:     5,
		// Replicas left default (1); set in server config for HA.
	})
	if err != nil {
		return nil, fmt.Errorf("config: open bucket %q: %w", bucket, err)
	}
	return &Store{nc: nc, js: js, kv: kv, bucket: bucket}, nil
}

// Bucket returns the KV bucket name.
func (s *Store) Bucket() string {
	if s == nil {
		return ""
	}
	return s.bucket
}

// KV exposes the underlying KeyValue for advanced use.
func (s *Store) KV() jetstream.KeyValue { return s.kv }

// PutRoute writes a route document (source of truth for gateways).
func (s *Store) PutRoute(ctx context.Context, doc RouteDoc) (RouteDoc, error) {
	if err := doc.Normalize(); err != nil {
		return RouteDoc{}, err
	}
	doc.UpdatedAt = time.Now().UTC()
	data, err := marshal(doc)
	if err != nil {
		return RouteDoc{}, err
	}
	if _, err := s.kv.Put(ctx, RouteKey(doc.ID), data); err != nil {
		return RouteDoc{}, err
	}
	_ = s.bumpRevision(ctx)
	return doc, nil
}

// DeleteRoute removes a route by id (or method+path via RouteID).
func (s *Store) DeleteRoute(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("route id required")
	}
	if err := s.kv.Delete(ctx, RouteKey(id)); err != nil {
		return err
	}
	_ = s.bumpRevision(ctx)
	return nil
}

// GetRoute loads one route.
func (s *Store) GetRoute(ctx context.Context, id string) (RouteDoc, error) {
	if id == "" {
		return RouteDoc{}, fmt.Errorf("route id required")
	}
	e, err := s.kv.Get(ctx, RouteKey(id))
	if err != nil {
		return RouteDoc{}, err
	}
	return unmarshalRoute(e.Value())
}

// Seed writes routes and ACLs in one shot (partial failure returns error after
// successful puts; not transactional).
func (s *Store) Seed(ctx context.Context, routes []RouteDoc, acls []WSACLDoc) error {
	for i, r := range routes {
		if _, err := s.PutRoute(ctx, r); err != nil {
			return fmt.Errorf("routes[%d]: %w", i, err)
		}
	}
	for i, a := range acls {
		if _, err := s.PutWSACL(ctx, a); err != nil {
			return fmt.Errorf("wsacl[%d]: %w", i, err)
		}
	}
	return nil
}

// IsNotFound reports whether err is a missing KV key.
func IsNotFound(err error) bool {
	return err != nil && errors.Is(err, jetstream.ErrKeyNotFound)
}

// ListRoutes returns all enabled (and optionally disabled) routes.
func (s *Store) ListRoutes(ctx context.Context, includeDisabled bool) ([]RouteDoc, error) {
	keys, err := s.kv.Keys(ctx, jetstream.IgnoreDeletes())
	if err != nil {
		// empty bucket
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return nil, nil
		}
		return nil, err
	}
	var out []RouteDoc
	for _, k := range keys {
		if len(k) < len(PrefixRoutes) || k[:len(PrefixRoutes)] != PrefixRoutes {
			continue
		}
		e, err := s.kv.Get(ctx, k)
		if err != nil {
			continue
		}
		doc, err := unmarshalRoute(e.Value())
		if err != nil {
			continue
		}
		if !includeDisabled && !doc.IsEnabled() {
			continue
		}
		out = append(out, doc)
	}
	return out, nil
}

// PutWSACL writes an ACL document.
func (s *Store) PutWSACL(ctx context.Context, doc WSACLDoc) (WSACLDoc, error) {
	if err := doc.Normalize(); err != nil {
		return WSACLDoc{}, err
	}
	doc.UpdatedAt = time.Now().UTC()
	data, err := marshal(doc)
	if err != nil {
		return WSACLDoc{}, err
	}
	if _, err := s.kv.Put(ctx, WSACLKey(doc.ID), data); err != nil {
		return WSACLDoc{}, err
	}
	_ = s.bumpRevision(ctx)
	return doc, nil
}

// DeleteWSACL removes an ACL by id.
func (s *Store) DeleteWSACL(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("wsacl id required")
	}
	if err := s.kv.Delete(ctx, WSACLKey(id)); err != nil {
		return err
	}
	_ = s.bumpRevision(ctx)
	return nil
}

// GetWSACL loads one ACL document.
func (s *Store) GetWSACL(ctx context.Context, id string) (WSACLDoc, error) {
	if id == "" {
		return WSACLDoc{}, fmt.Errorf("wsacl id required")
	}
	e, err := s.kv.Get(ctx, WSACLKey(id))
	if err != nil {
		return WSACLDoc{}, err
	}
	return unmarshalACL(e.Value())
}

// ListWSACL returns ACL docs.
func (s *Store) ListWSACL(ctx context.Context, includeDisabled bool) ([]WSACLDoc, error) {
	keys, err := s.kv.Keys(ctx, jetstream.IgnoreDeletes())
	if err != nil {
		if errors.Is(err, jetstream.ErrNoKeysFound) {
			return nil, nil
		}
		return nil, err
	}
	var out []WSACLDoc
	for _, k := range keys {
		if len(k) < len(PrefixWSACL) || k[:len(PrefixWSACL)] != PrefixWSACL {
			continue
		}
		e, err := s.kv.Get(ctx, k)
		if err != nil {
			continue
		}
		doc, err := unmarshalACL(e.Value())
		if err != nil {
			continue
		}
		if !includeDisabled && !doc.IsEnabled() {
			continue
		}
		out = append(out, doc)
	}
	return out, nil
}

func (s *Store) bumpRevision(ctx context.Context) error {
	var rev uint64 = 1
	if e, err := s.kv.Get(ctx, KeyRevision); err == nil {
		if n, err2 := strconv.ParseUint(string(e.Value()), 10, 64); err2 == nil {
			rev = n + 1
		}
	}
	_, err := s.kv.Put(ctx, KeyRevision, []byte(strconv.FormatUint(rev, 10)))
	return err
}

// Revision returns the current meta revision (0 if missing).
func (s *Store) Revision(ctx context.Context) (uint64, error) {
	e, err := s.kv.Get(ctx, KeyRevision)
	if err != nil {
		if errors.Is(err, jetstream.ErrKeyNotFound) {
			return 0, nil
		}
		return 0, err
	}
	return strconv.ParseUint(string(e.Value()), 10, 64)
}
