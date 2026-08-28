package mini

import (
	"sync"

	"github.com/nats-io/nats.go"
)

// connHub multiplexes Closed/Error handlers so multiple services can share one Conn.
type connHub struct {
	mu sync.Mutex

	// original handlers installed on the connection before the hub took over
	// (plus any that were set by non-mini code after — we chain the first snapshot).
	userClosed nats.ConnHandler
	userError  nats.ErrHandler

	closed map[uint64]nats.ConnHandler
	errors map[uint64]nats.ErrHandler
	nextID uint64
}

var hubs sync.Map // *nats.Conn -> *connHub

func hubFor(nc *nats.Conn) *connHub {
	if nc == nil {
		return nil
	}
	if v, ok := hubs.Load(nc); ok {
		return v.(*connHub)
	}
	h := &connHub{
		closed: make(map[uint64]nats.ConnHandler),
		errors: make(map[uint64]nats.ErrHandler),
	}
	actual, loaded := hubs.LoadOrStore(nc, h)
	if loaded {
		return actual.(*connHub)
	}
	h.install(nc)
	return h
}

func (h *connHub) install(nc *nats.Conn) {
	h.mu.Lock()
	h.userClosed = nc.ClosedHandler()
	h.userError = nc.ErrorHandler()
	h.mu.Unlock()

	nc.SetClosedHandler(func(c *nats.Conn) {
		h.mu.Lock()
		user := h.userClosed
		fns := make([]nats.ConnHandler, 0, len(h.closed))
		for _, fn := range h.closed {
			fns = append(fns, fn)
		}
		h.mu.Unlock()
		for _, fn := range fns {
			if fn != nil {
				fn(c)
			}
		}
		if user != nil {
			user(c)
		}
		hubs.Delete(c)
	})

	nc.SetErrorHandler(func(c *nats.Conn, sub *nats.Subscription, err error) {
		h.mu.Lock()
		user := h.userError
		fns := make([]nats.ErrHandler, 0, len(h.errors))
		for _, fn := range h.errors {
			fns = append(fns, fn)
		}
		h.mu.Unlock()
		for _, fn := range fns {
			if fn != nil {
				fn(c, sub, err)
			}
		}
		if user != nil {
			user(c, sub, err)
		}
	})
}

// onConnClosed registers a closed handler. Returns unregister func.
func onConnClosed(nc *nats.Conn, fn nats.ConnHandler) (unregister func()) {
	if nc == nil || fn == nil {
		return func() {}
	}
	h := hubFor(nc)
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	h.closed[id] = fn
	h.mu.Unlock()
	return func() {
		h.mu.Lock()
		delete(h.closed, id)
		h.mu.Unlock()
		// Hub stays installed until the connection closes (handlers are no-ops when empty).
	}
}

// onConnError registers an async error handler. Returns unregister func.
func onConnError(nc *nats.Conn, fn nats.ErrHandler) (unregister func()) {
	if nc == nil || fn == nil {
		return func() {}
	}
	h := hubFor(nc)
	h.mu.Lock()
	h.nextID++
	id := h.nextID
	h.errors[id] = fn
	h.mu.Unlock()
	return func() {
		h.mu.Lock()
		delete(h.errors, id)
		h.mu.Unlock()
	}
}
