package websocket

import (
	"sync"
)

// Message is a fan-out message.
type Message struct {
	Conn string
	Data []byte
}

// Subscriber is a conn subscriber.
type Subscriber struct {
	ID      string
	Subject string
	Send    chan []byte
}

// Fanout is a simple pub/sub for WebSocket conn.
type Fanout struct {
	mu   sync.RWMutex
	subs map[string]map[string]*Subscriber // conn → subID → sub
}

// NewFanout creates a Fanout.
func NewFanout() *Fanout {
	return &Fanout{
		subs: make(map[string]map[string]*Subscriber),
	}
}

// Subscribe attaches a subscriber to a conn.
func (f *Fanout) Subscribe(conn string, sub *Subscriber) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.subs[conn] == nil {
		f.subs[conn] = make(map[string]*Subscriber)
	}
	f.subs[conn][sub.ID] = sub
}

// Unsubscribe detaches a subscriber.
func (f *Fanout) Unsubscribe(conn, subID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if m, ok := f.subs[conn]; ok {
		delete(m, subID)
		if len(m) == 0 {
			delete(f.subs, conn)
		}
	}
}

// UnsubscribeAll detaches a subscriber from all conn.
func (f *Fanout) UnsubscribeAll(subID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for ch, m := range f.subs {
		delete(m, subID)
		if len(m) == 0 {
			delete(f.subs, ch)
		}
	}
}

// Publish sends data to all conn subscribers (non-blocking).
func (f *Fanout) Publish(conn string, data []byte) int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	m, ok := f.subs[conn]
	if !ok {
		return 0
	}
	sent := 0
	for _, sub := range m {
		select {
		case sub.Send <- data:
			sent++
		default:
			// buffer full — drop
		}
	}
	return sent
}

// Conn returns active conn names.
func (f *Fanout) Conn() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make([]string, 0, len(f.subs))
	for ch := range f.subs {
		out = append(out, ch)
	}
	return out
}
