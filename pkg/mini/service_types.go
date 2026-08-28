package mini

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nats-io/nats.go"
)

type (

	// Service exposes methods to operate on a service instance.
	Service interface {
		// AddEndpoint registers endpoint with given name on a specific subject.
		AddEndpoint(string, Handler, ...EndpointOpt) error

		// AddGroup returns a Group interface, allowing for more complex endpoint topologies.
		// A group can be used to register endpoints with given prefix.
		AddGroup(string, ...GroupOpt) Group

		// AddPublicWS advertises a public WebSocket route via $SRV.INFO without
		// requiring a business request/reply handler. A no-op internal endpoint
		// is registered so the metadata is discoverable by the gate.
		AddPublicWS(name string, cfg WSConfig, opts ...EndpointOpt) error

		// Info returns the service info.
		Info() Info

		// Stats returns statistics for the service endpoint and all monitoring endpoints.
		Stats() Stats

		// Reset resets all statistics (for all endpoints) on a service instance.
		Reset()

		// Stop drains endpoint subscriptions and marks the service as stopped.
		// Equivalent to StopContext(context.Background()).
		Stop() error

		// StopContext cancels in-flight request contexts, drains subscriptions,
		// then waits for handlers to finish until ctx is done.
		// If ctx expires while handlers are still running, it returns ctx.Err()
		// after subscriptions have been drained (best-effort graceful shutdown).
		StopContext(ctx context.Context) error

		// Stopped informs whether [Stop] was executed on the service.
		Stopped() bool
	}

	// Group allows for grouping endpoints on a service.
	//
	// Endpoints created using AddEndpoint will be grouped under common prefix (group name)
	// New groups can also be derived from a group using AddGroup.
	Group interface {
		// AddGroup creates a new group, prefixed by this group's prefix.
		AddGroup(string, ...GroupOpt) Group

		// AddEndpoint registers new endpoints on a service.
		// The endpoint's subject will be prefixed with the group prefix.
		AddEndpoint(string, Handler, ...EndpointOpt) error
	}

	EndpointOpt func(*endpointOpts) error
	GroupOpt    func(*groupOpts)

	endpointOpts struct {
		subject    string
		metadata   map[string]string
		queueGroup string
		qgDisabled bool
		msgLimit   int
		bytesLimit int
		middleware []Middleware
	}

	groupOpts struct {
		queueGroup string
		qgDisabled bool
	}

	// ErrHandler is a function used to configure a custom error handler for a service,
	ErrHandler func(Service, *NATSError)

	// DoneHandler is a function used to configure a custom done handler for a service.
	DoneHandler func(Service)

	// StatsHandler is a function used to configure a custom STATS endpoint.
	// It should return a value which can be serialized to JSON.
	StatsHandler func(*Endpoint) any

	// ServiceIdentity contains fields helping to identity a service instance.
	ServiceIdentity struct {
		Name     string            `json:"name"`
		ID       string            `json:"id"`
		Version  string            `json:"version"`
		Metadata map[string]string `json:"metadata"`
	}

	// Stats is the type returned by STATS monitoring endpoint.
	// It contains stats of all registered endpoints.
	Stats struct {
		ServiceIdentity
		Type      string           `json:"type"`
		Started   time.Time        `json:"started"`
		Endpoints []*EndpointStats `json:"endpoints"`
	}

	// EndpointStats contains stats for a specific endpoint.
	EndpointStats struct {
		Name                  string          `json:"name"`
		Subject               string          `json:"subject"`
		QueueGroup            string          `json:"queue_group"`
		NumRequests           int             `json:"num_requests"`
		NumErrors             int             `json:"num_errors"`
		LastError             string          `json:"last_error"`
		ProcessingTime        time.Duration   `json:"processing_time"`
		AverageProcessingTime time.Duration   `json:"average_processing_time"`
		Data                  json.RawMessage `json:"data,omitempty"`
	}

	// Ping is the response type for PING monitoring endpoint.
	Ping struct {
		ServiceIdentity
		Type string `json:"type"`
	}

	// Info is the basic information about a service type.
	Info struct {
		ServiceIdentity
		Type        string         `json:"type"`
		Description string         `json:"description"`
		Endpoints   []EndpointInfo `json:"endpoints"`
	}

	EndpointInfo struct {
		Name       string            `json:"name"`
		Subject    string            `json:"subject"`
		QueueGroup string            `json:"queue_group"`
		Metadata   map[string]string `json:"metadata"`
	}

	// Endpoint manages a service endpoint.
	Endpoint struct {
		EndpointConfig
		Name string

		service *service

		// stats identity fields (name/subject/queue) protected by service.m on read snapshots.
		stats EndpointStats
		// hot-path counters (lock-free)
		numRequests  atomic.Int64
		numErrors    atomic.Int64
		processingNs atomic.Int64
		lastError    atomic.Value // string
		subscription *nats.Subscription
	}

	group struct {
		service            *service
		prefix             string
		queueGroup         string
		queueGroupDisabled bool
	}

	// Verb represents a name of the monitoring service.
	Verb int64

	// Config is a configuration of a service.
	Config struct {
		// Name represents the name of the service.
		Name string `json:"name"`

		// Endpoint is an optional endpoint configuration.
		// More complex, multi-endpoint services can be configured using
		// Service.AddGroup and Service.AddEndpoint methods.
		Endpoint *EndpointConfig `json:"endpoint"`

		// Version is a SemVer compatible version string.
		Version string `json:"version"`

		// Description of the service.
		Description string `json:"description"`

		// Metadata annotates the service
		Metadata map[string]string `json:"metadata,omitempty"`

		// QueueGroup can be used to override the default queue group name.
		QueueGroup string `json:"queue_group"`

		// QueueGroupDisabled disables the queue group for the service.
		QueueGroupDisabled bool `json:"queue_group_disabled"`

		// StatsHandler is a user-defined custom function.
		// used to calculate additional service stats.
		StatsHandler StatsHandler

		// DoneHandler is invoked when all service subscription are stopped.
		DoneHandler DoneHandler

		// ErrorHandler is invoked on any nats-related service error.
		ErrorHandler ErrHandler

		// Middleware is applied to every endpoint handler (outermost first).
		// Built-ins: Recover, Logging, Timeout, Observe, RateLimit.
		Middleware []Middleware

		// Observer receives request lifecycle events (tracing, metrics).
		// Equivalent to appending mini.Observe(Observer) middleware.
		Observer Observer

		// OnAsyncError controls behavior when a NATS async subscription error
		// is delivered for this service. Default is OnAsyncErrorLog (do not stop).
		OnAsyncError OnAsyncError

		// AdvertiseWhenPublic publishes mini.AdvertiseRefresh after registering
		// an endpoint marked mini.public=true so gateways re-discover immediately.
		// Default false (opt-in).
		AdvertiseWhenPublic bool
	}

	// OnAsyncError selects how subscription-level async errors are handled.
	OnAsyncError int

	EndpointConfig struct {
		// Subject on which the endpoint is registered.
		Subject string

		// Handler used by the endpoint.
		Handler Handler

		// Metadata annotates the service
		Metadata map[string]string `json:"metadata,omitempty"`

		// QueueGroup can be used to override the default queue group name.
		QueueGroup string `json:"queue_group"`

		// QueueGroupDisabled disables the queue group for the endpoint.
		QueueGroupDisabled bool `json:"queue_group_disabled"`
	}

	// NATSError represents an error returned by a NATS Subscription.
	// It contains a subject on which the subscription failed, so that
	// it can be linked with a specific service endpoint.
	NATSError struct {
		Subject     string
		Description string
		err         error
	}

	// service represents a configured NATS service.
	// It should be created using [Add] in order to configure the appropriate NATS subscriptions
	// for request handler and monitoring.
	service struct {
		// Config contains a configuration of the service
		Config

		m         sync.Mutex
		id        string
		endpoints []*Endpoint
		verbSubs  map[string]*nats.Subscription
		started   time.Time
		nc        *nats.Conn
		// stopped is atomic so the request hot path can avoid taking s.m.
		stopped atomic.Bool

		asyncDispatcher asyncCallbacksHandler

		// baseCtx is canceled when the service stops.
		baseCtx    context.Context
		baseCancel context.CancelFunc

		// inFlight tracks active request handlers for graceful shutdown.
		inFlight sync.WaitGroup

		// unreg* detach this service from the shared conn hub.
		unregClosed func()
		unregError  func()
	}

	asyncCallbacksHandler struct {
		cbQueue chan func()
		closed  bool
	}
)

// OnAsyncError values.
const (
	// OnAsyncErrorLog invokes ErrorHandler (if any), updates stats, keeps service running.
	OnAsyncErrorLog OnAsyncError = iota
	// OnAsyncErrorStop also stops the whole service (legacy NATS micro behavior).
	OnAsyncErrorStop
)

const (
	// Queue Group name used across all services
	DefaultQueueGroup = "q"

	// APIPrefix is the root of all control subjects
	APIPrefix = "$SRV"
)

// Service Error headers
const (
	ErrorHeader     = "Nats-Service-Error"
	ErrorCodeHeader = "Nats-Service-Error-Code"

	// StatusHeader carries an HTTP status code (100–599) for gate passthrough
	// on successful responses. Set via WithStatus or Respond headers.
	StatusHeader = "Nats-Service-Status"
)

// Verbs being used to set up a specific control subject.
const (
	PingVerb Verb = iota
	StatsVerb
	InfoVerb
)

const (
	InfoResponseType  = "io.nats.micro.v1.info_response"
	PingResponseType  = "io.nats.micro.v1.ping_response"
	StatsResponseType = "io.nats.micro.v1.stats_response"
)

var (
	// this regular expression is suggested regexp for semver validation: https://semver.org/
	semVerRegexp  = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	nameRegexp    = regexp.MustCompile(`^[A-Za-z0-9\-_]+$`)
	subjectRegexp = regexp.MustCompile(`^[^ >]*[>]?$`)
)

// Common errors returned by the Service framework.
var (
	// ErrConfigValidation is returned when service configuration is invalid
	ErrConfigValidation = errors.New("validation")

	// ErrVerbNotSupported is returned when invalid [Verb] is used (PING, INFO, STATS)
	ErrVerbNotSupported = errors.New("unsupported verb")

	// ErrServiceNameRequired is returned when attempting to generate control subject with ID but empty name
	ErrServiceNameRequired = errors.New("service name is required to generate ID control subject")
)

func (s Verb) String() string {
	switch s {
	case PingVerb:
		return "PING"
	case StatsVerb:
		return "STATS"
	case InfoVerb:
		return "INFO"
	default:
		return ""
	}
}
