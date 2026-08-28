package mini

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

func (g *group) AddEndpoint(name string, handler Handler, opts ...EndpointOpt) error {
	var options endpointOpts
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return err
		}
	}
	subject := name
	if options.subject != "" {
		subject = options.subject
	}
	endpointSubject := fmt.Sprintf("%s.%s", g.prefix, subject)
	if g.prefix == "" {
		endpointSubject = subject
	}
	queueGroup, noQueue := resolveQueueGroup(options.queueGroup, g.queueGroup, options.qgDisabled, g.queueGroupDisabled)

	return addEndpoint(g.service, name, endpointSubject, handler, options.metadata, queueGroup, noQueue, options.msgLimit, options.bytesLimit, options.middleware)
}

func resolveQueueGroup(customQG, parentQG string, disabled, parentDisabled bool) (string, bool) {
	if disabled {
		return "", true
	}
	if customQG != "" {
		return customQG, false
	}
	if parentDisabled {
		return "", true
	}
	if parentQG != "" {
		return parentQG, false
	}
	return DefaultQueueGroup, false
}

func (g *group) AddGroup(name string, opts ...GroupOpt) Group {
	var o groupOpts
	for _, opt := range opts {
		opt(&o)
	}
	queueGroup, noQueue := resolveQueueGroup(o.queueGroup, g.queueGroup, o.qgDisabled, g.queueGroupDisabled)

	parts := make([]string, 0, 2)
	if g.prefix != "" {
		parts = append(parts, g.prefix)
	}
	if name != "" {
		parts = append(parts, name)
	}
	prefix := strings.Join(parts, ".")

	return &group{
		service:            g.service,
		prefix:             prefix,
		queueGroup:         queueGroup,
		queueGroupDisabled: noQueue,
	}
}

func (e *Endpoint) stop() error {
	// Drain the subscription. If the connection is closed, draining is not possible.
	if e.subscription == nil {
		return nil
	}
	if err := e.subscription.Drain(); err != nil && !errors.Is(err, nats.ErrConnectionClosed) {
		return fmt.Errorf("draining subscription for request handler: %w", err)
	}
	return nil
}

func (e *Endpoint) reset() {
	e.stats = EndpointStats{
		Name:       e.stats.Name,
		Subject:    e.stats.Subject,
		QueueGroup: e.stats.QueueGroup,
	}
	e.numRequests.Store(0)
	e.numErrors.Store(0)
	e.processingNs.Store(0)
	e.lastError.Store("")
}

// ControlSubject returns monitoring subjects used by the Service.
// Providing a verb is mandatory (it should be one of Ping, Info or Stats).
// Depending on whether kind and id are provided, ControlSubject will return one of the following:
//   - verb only: subject used to monitor all available services
//   - verb and kind: subject used to monitor services with the provided name
//   - verb, name and id: subject used to monitor an instance of a service with the provided ID
