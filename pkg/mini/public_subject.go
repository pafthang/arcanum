package mini

import (
	"fmt"
	"strings"
)

// PublicSubjectPrefix is the recommended subject namespace for endpoints that
// the edge gate is allowed to call.
//
// With NATS accounts, export only `public.>` from the internal account to the
// edge/gate account so the gate cannot invoke private subjects.
const PublicSubjectPrefix = "public"

// PublicSubject builds a subject under the public namespace:
//
//	PublicSubject("orders", "create") => "public.orders.create"
//
// Common 1–2 part forms avoid intermediate slices.
func PublicSubject(parts ...string) string {
	switch len(parts) {
	case 0:
		return PublicSubjectPrefix
	case 1:
		a := strings.Trim(parts[0], ".")
		if a == "" {
			return PublicSubjectPrefix
		}
		return PublicSubjectPrefix + "." + a
	case 2:
		a := strings.Trim(parts[0], ".")
		b := strings.Trim(parts[1], ".")
		switch {
		case a == "" && b == "":
			return PublicSubjectPrefix
		case a == "":
			return PublicSubjectPrefix + "." + b
		case b == "":
			return PublicSubjectPrefix + "." + a
		default:
			return PublicSubjectPrefix + "." + a + "." + b
		}
	}

	var b strings.Builder
	b.Grow(len(PublicSubjectPrefix) + 8*len(parts))
	b.WriteString(PublicSubjectPrefix)
	for _, p := range parts {
		p = strings.Trim(p, ".")
		if p == "" {
			continue
		}
		b.WriteByte('.')
		b.WriteString(p)
	}
	return b.String()
}

// IsPublicSubject reports whether subject is under public.>
func IsPublicSubject(subject string) bool {
	return subject == PublicSubjectPrefix ||
		strings.HasPrefix(subject, PublicSubjectPrefix+".")
}

// WithPublicSubject sets the endpoint NATS subject to public.<parts...>.
// Combine with WithPublicHTTP to both expose HTTP and bind a public subject:
//
//	svc.AddEndpoint("get", h,
//	    mini.WithPublicSubject("orders", "get"),
//	    mini.WithPublicHTTP("GET", "/v1/orders/{id}"),
//	)
//
// Prefer [Public] when auth is the default required mode.
func WithPublicSubject(parts ...string) EndpointOpt {
	return func(e *endpointOpts) error {
		subj := PublicSubject(parts...)
		if subj == PublicSubjectPrefix {
			return fmt.Errorf("%w: public subject requires at least one part after %q", ErrConfigValidation, PublicSubjectPrefix)
		}
		e.subject = subj
		return nil
	}
}

// Public combines subject, HTTP method/path, and AuthRequired in one option:
//
//	svc.AddEndpoint("list", h, mini.Public("GET", "/api/projects", "task", "projects.list"))
//
// Equivalent to WithPublicSubject(domain, op) + WithPublicHTTP(method, path) + WithPublicAuth(AuthRequired).
// Append more EndpointOpt after this for timeout, OpenAPI, multipart, etc.
func Public(method, path, domain, op string) EndpointOpt {
	return func(e *endpointOpts) error {
		if err := WithPublicSubject(domain, op)(e); err != nil {
			return err
		}
		if err := WithPublicHTTP(method, path)(e); err != nil {
			return err
		}
		return WithPublicAuth(AuthRequired)(e)
	}
}

// PublicWithAuth is like Public but with an explicit auth mode (required|optional|none).
func PublicWithAuth(method, path, domain, op, auth string) EndpointOpt {
	return func(e *endpointOpts) error {
		if err := WithPublicSubject(domain, op)(e); err != nil {
			return err
		}
		if err := WithPublicHTTP(method, path)(e); err != nil {
			return err
		}
		return WithPublicAuth(auth)(e)
	}
}

// Internal sets the endpoint NATS subject for service-to-service RPC
// (alias of WithEndpointSubject; use with subjects.Internal* constants).
func Internal(subject string) EndpointOpt {
	return WithEndpointSubject(subject)
}

// WithModule sets mini.module metadata (logical module inside a composite process).
func WithModule(name string) EndpointOpt {
	return WithEndpointMetadataKey(MetaModule, name)
}
