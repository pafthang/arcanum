package mini

import (
	"fmt"
)

func ControlSubject(verb Verb, name, id string) (string, error) {
	verbStr := verb.String()
	if verbStr == "" {
		return "", fmt.Errorf("%w: %q", ErrVerbNotSupported, verbStr)
	}
	if name == "" && id != "" {
		return "", ErrServiceNameRequired
	}
	if name == "" && id == "" {
		return fmt.Sprintf("%s.%s", APIPrefix, verbStr), nil
	}
	if id == "" {
		return fmt.Sprintf("%s.%s.%s", APIPrefix, verbStr, name), nil
	}
	return fmt.Sprintf("%s.%s.%s.%s", APIPrefix, verbStr, name, id), nil
}

func WithEndpointSubject(subject string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.subject = subject
		return nil
	}
}

func WithEndpointMetadata(metadata map[string]string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.metadata = metadata
		return nil
	}
}

// WithEndpointMetadataKey adds a key-value pair to the endpoints's metadata.
// Prefer using WithEndpointMetadata when you have all the key-value pairs you
// want to add at once or when you want to replace any existing metadata.
func WithEndpointMetadataKey(key, value string) EndpointOpt {
	return func(e *endpointOpts) error {
		if e.metadata == nil {
			e.metadata = map[string]string{}
		}
		e.metadata[key] = value
		return nil
	}
}

func WithEndpointQueueGroup(queueGroup string) EndpointOpt {
	return func(e *endpointOpts) error {
		e.queueGroup = queueGroup
		return nil
	}
}

func WithEndpointQueueGroupDisabled() EndpointOpt {
	return func(e *endpointOpts) error {
		e.qgDisabled = true
		return nil
	}
}

// WithEndpointMiddleware adds middleware that runs inside service-level
// middleware and outside the observer (when configured).
// Order: service Middleware → endpoint Middleware → Observer → handler.
func WithEndpointMiddleware(mws ...Middleware) EndpointOpt {
	return func(e *endpointOpts) error {
		e.middleware = append(e.middleware, mws...)
		return nil
	}
}

// WithEndpointPendingLimits sets the pending limits for the endpoint's
// subscription. These limits how many messages and/or bytes can be buffered in
// memory before the subscription is terminated with nats.ErrSlowConsumer.
// Either limit can be set to -1 to indicate no limit.
func WithEndpointPendingLimits(msgLimit, bytesLimit int) EndpointOpt {
	return func(e *endpointOpts) error {
		if msgLimit == 0 && bytesLimit == 0 {
			return fmt.Errorf("%w: at least one pending limit must be non-zero", ErrConfigValidation)
		}
		e.msgLimit = msgLimit
		e.bytesLimit = bytesLimit
		return nil
	}
}

func WithGroupQueueGroup(queueGroup string) GroupOpt {
	return func(g *groupOpts) {
		g.queueGroup = queueGroup
	}
}

func WithGroupQueueGroupDisabled() GroupOpt {
	return func(g *groupOpts) {
		g.qgDisabled = true
	}
}
