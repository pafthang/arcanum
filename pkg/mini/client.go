package mini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// DefaultRequestTimeout is used when no timeout is set on the client or context.
const DefaultRequestTimeout = 5 * time.Second

// Client is a thin NATS request/reply client for inter-service calls.
type Client struct {
	nc      *nats.Conn
	timeout time.Duration

	// retries is extra attempts after the first failure (0 = no retry).
	retries int
	// retryBase is the initial backoff between attempts.
	retryBase time.Duration
	// breaker optional per-subject circuit breaker.
	breaker *circuitBreaker
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithClientTimeout sets the default request timeout.
func WithClientTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		if d > 0 {
			c.timeout = d
		}
	}
}

// WithClientRetries enables retries on transient transport failures
// (no responders, timeout). Application ServiceError is never retried.
// retries is the number of additional attempts (total attempts = 1+retries).
func WithClientRetries(retries int, baseBackoff time.Duration) ClientOption {
	return func(c *Client) {
		if retries < 0 {
			retries = 0
		}
		c.retries = retries
		if baseBackoff <= 0 {
			baseBackoff = 25 * time.Millisecond
		}
		c.retryBase = baseBackoff
	}
}

// WithClientCircuitBreaker enables a simple consecutive-failure circuit breaker.
// After tripFailures consecutive transport errors on a subject, further calls fail
// fast for cooldown. Success resets the failure count.
func WithClientCircuitBreaker(tripFailures int, cooldown time.Duration) ClientOption {
	return func(c *Client) {
		if tripFailures < 1 {
			tripFailures = 5
		}
		if cooldown <= 0 {
			cooldown = 5 * time.Second
		}
		c.breaker = newCircuitBreaker(tripFailures, cooldown)
	}
}

// NewClient creates a Client bound to the given connection.
func NewClient(nc *nats.Conn, opts ...ClientOption) (*Client, error) {
	if nc == nil {
		return nil, ErrInvalidConnection
	}
	c := &Client{nc: nc, timeout: DefaultRequestTimeout, retryBase: 25 * time.Millisecond}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// CallResult is a successful (or application-error) response from a service.
type CallResult struct {
	Data    []byte
	Headers Headers
	Subject string
}

// ServiceError is an application-level error returned by a service handler
// via Request.Error (Nats-Service-Error headers).
type ServiceError struct {
	Code        string
	Description string
	Data        []byte
	Headers     Headers
}

func (e *ServiceError) Error() string {
	if e.Description != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Description)
	}
	return e.Code
}

// HTTPStatus maps common service error codes to HTTP status codes.
// Numeric codes in the 100–599 range are used as-is; well-known names are mapped;
// everything else defaults to 500.
func (e *ServiceError) HTTPStatus() int {
	if e == nil || e.Code == "" {
		return 500
	}
	if n, err := strconv.Atoi(e.Code); err == nil && n >= 100 && n <= 599 {
		return n
	}
	switch e.Code {
	case "BAD_REQUEST", "INVALID_ARGUMENT", "INVALID":
		return 400
	case "UNAUTHORIZED":
		return 401
	case "FORBIDDEN":
		return 403
	case "NOT_FOUND":
		return 404
	case "CONFLICT":
		return 409
	case "TOO_MANY_REQUESTS":
		return 429
	case "NOT_IMPLEMENTED":
		return 501
	case "UNAVAILABLE", "SERVICE_UNAVAILABLE":
		return 503
	case "TIMEOUT", "DEADLINE_EXCEEDED":
		return 504
	default:
		return 500
	}
}

// ErrCircuitOpen is returned when the client circuit breaker is open for a subject.
var ErrCircuitOpen = errors.New("circuit breaker open")

// Request sends a request to subject and returns the result.
// If the service responds with error headers, a *ServiceError is returned
// (CallResult data/headers are still available on the error).
func (c *Client) Request(ctx context.Context, subject string, data []byte, headers Headers) (*CallResult, error) {
	if c == nil || c.nc == nil {
		return nil, ErrInvalidConnection
	}
	if subject == "" {
		return nil, fmt.Errorf("%w: subject required", ErrConfigValidation)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	// Apply default client timeout if the caller did not set a deadline.
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		timeout := c.timeout
		if timeout <= 0 {
			timeout = DefaultRequestTimeout
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if c.breaker != nil {
		if err := c.breaker.allow(subject); err != nil {
			return nil, err
		}
	}

	attempts := 1 + c.retries
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for i := 0; i < attempts; i++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if i > 0 {
			backoff := c.retryBase * time.Duration(1<<uint(i-1))
			// jitter ±20%
			j := time.Duration(rand.Int64N(int64(backoff/5) + 1))
			sleep := backoff - backoff/10 + j
			t := time.NewTimer(sleep)
			select {
			case <-ctx.Done():
				t.Stop()
				return nil, ctx.Err()
			case <-t.C:
			}
		}

		msg := &nats.Msg{
			Subject: subject,
			Data:    data,
		}
		if len(headers) > 0 {
			msg.Header = nats.Header(cloneHeaders(headers))
		} else {
			msg.Header = nats.Header{}
		}

		resp, err := c.nc.RequestMsgWithContext(ctx, msg)
		if err != nil {
			lastErr = err
			if !isTransientClientErr(err) || i == attempts-1 {
				if c.breaker != nil && isTransientClientErr(err) {
					c.breaker.fail(subject)
				}
				return nil, err
			}
			continue
		}
		res, err := callResultFromMsg(resp)
		if err != nil {
			// application error — do not trip breaker, do not retry
			if c.breaker != nil {
				c.breaker.success(subject)
			}
			return res, err
		}
		if c.breaker != nil {
			c.breaker.success(subject)
		}
		return res, nil
	}
	return nil, lastErr
}

func isTransientClientErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, nats.ErrNoResponders) || errors.Is(err, nats.ErrTimeout) {
		return true
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	return false
}

// RequestJSON marshals in, performs Request, and unmarshals into out on success.
func (c *Client) RequestJSON(ctx context.Context, subject string, in any, out any, headers Headers) error {
	var data []byte
	var err error
	if in != nil {
		data, err = json.Marshal(in)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
	}
	// Ensure Content-Type without always cloning when already set.
	if headers.Get("Content-Type") == "" {
		headers = cloneHeaders(headers)
		nats.Header(headers).Set("Content-Type", "application/json")
	}
	res, err := c.Request(ctx, subject, data, headers)
	if err != nil {
		return err
	}
	if out == nil || len(res.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(res.Data, out); err != nil {
		return fmt.Errorf("unmarshal response: %w", err)
	}
	return nil
}

// CallJSON is a generic convenience over RequestJSON: marshals in, calls subject,
// unmarshals the response body into T. Prefer typed service clients that wrap this.
//
//	out, err := mini.CallJSON[Agent](ctx, c, subjects.InternalAgentGet, map[string]any{"id": id})
func CallJSON[T any](ctx context.Context, c *Client, subject string, in any) (T, error) {
	var out T
	if c == nil {
		return out, ErrInvalidConnection
	}
	err := c.RequestJSON(ctx, subject, in, &out, nil)
	return out, err
}

// CallJSONHeaders is CallJSON with optional request headers (auth forwarding, etc.).
func CallJSONHeaders[T any](ctx context.Context, c *Client, subject string, in any, headers Headers) (T, error) {
	var out T
	if c == nil {
		return out, ErrInvalidConnection
	}
	err := c.RequestJSON(ctx, subject, in, &out, headers)
	return out, err
}

func callResultFromMsg(msg *nats.Msg) (*CallResult, error) {
	res := &CallResult{
		Data:    msg.Data,
		Headers: Headers(msg.Header),
		Subject: msg.Subject,
	}
	code := msg.Header.Get(ErrorCodeHeader)
	desc := msg.Header.Get(ErrorHeader)
	if code != "" || desc != "" {
		return res, &ServiceError{
			Code:        code,
			Description: desc,
			Data:        msg.Data,
			Headers:     Headers(msg.Header),
		}
	}
	return res, nil
}

func cloneHeaders(h Headers) Headers {
	if h == nil {
		return Headers{}
	}
	out := make(Headers, len(h))
	for k, v := range h {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// AsServiceError extracts *ServiceError from err if present.
func AsServiceError(err error) (*ServiceError, bool) {
	var se *ServiceError
	if errors.As(err, &se) {
		return se, true
	}
	return nil, false
}

// --- circuit breaker ---

type circuitBreaker struct {
	trip     int
	cooldown time.Duration
	mu       sync.Mutex
	states   map[string]*cbState
}

type cbState struct {
	failures  int
	openUntil time.Time
}

func newCircuitBreaker(trip int, cooldown time.Duration) *circuitBreaker {
	return &circuitBreaker{
		trip:     trip,
		cooldown: cooldown,
		states:   make(map[string]*cbState),
	}
}

func (b *circuitBreaker) allow(subject string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	st := b.states[subject]
	if st == nil {
		return nil
	}
	if time.Now().Before(st.openUntil) {
		return fmt.Errorf("%w: %s", ErrCircuitOpen, subject)
	}
	return nil
}

func (b *circuitBreaker) fail(subject string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	st := b.states[subject]
	if st == nil {
		st = &cbState{}
		b.states[subject] = st
	}
	st.failures++
	if st.failures >= b.trip {
		st.openUntil = time.Now().Add(b.cooldown)
		st.failures = 0
	}
}

func (b *circuitBreaker) success(subject string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if st := b.states[subject]; st != nil {
		st.failures = 0
		st.openUntil = time.Time{}
	}
}
