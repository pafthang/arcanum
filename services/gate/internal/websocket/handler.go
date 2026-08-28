package websocket

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     nil, // set via OriginChecker
}

// HandlerConfig configures the WebSocket handler.
type HandlerConfig struct {
	ACL            *ACL
	Fanout         *Fanout
	ConnLimiter    *ConnLimiter
	OriginChecker  OriginChecker
	PingInterval   time.Duration
	PongWait       time.Duration
	WriteWait      time.Duration
	MaxMessageSize int64
}

// DefaultHandlerConfig returns sensible defaults.
func DefaultHandlerConfig() HandlerConfig {
	return HandlerConfig{
		PingInterval:   30 * time.Second,
		PongWait:       60 * time.Second,
		WriteWait:      10 * time.Second,
		MaxMessageSize: 512 * 1024, // 512 KiB
	}
}

// Handler is an HTTP handler for WebSocket upgrade.
type Handler struct {
	cfg HandlerConfig
}

// NewHandler creates a handler.
func NewHandler(cfg HandlerConfig) *Handler {
	if cfg.PingInterval == 0 {
		cfg.PingInterval = 30 * time.Second
	}
	if cfg.PongWait == 0 {
		cfg.PongWait = 60 * time.Second
	}
	if cfg.WriteWait == 0 {
		cfg.WriteWait = 10 * time.Second
	}
	if cfg.MaxMessageSize == 0 {
		cfg.MaxMessageSize = 512 * 1024
	}
	return &Handler{cfg: cfg}
}

// ServeHTTP upgrades and serves the connection.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.cfg.OriginChecker != nil && !h.cfg.OriginChecker.Check(r) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return
	}

	// Subject from auth middleware (if any)
	subject := r.Header.Get("X-Subject")
	if subject == "" {
		subject = "anonymous"
	}

	if h.cfg.ConnLimiter != nil && !h.cfg.ConnLimiter.Acquire(subject) {
		http.Error(w, "connection limit exceeded", http.StatusTooManyRequests)
		return
	}
	defer func() {
		if h.cfg.ConnLimiter != nil {
			h.cfg.ConnLimiter.Release(subject)
		}
	}()

	upgrader.CheckOrigin = func(r *http.Request) bool {
		if h.cfg.OriginChecker == nil {
			return true
		}
		return h.cfg.OriginChecker.Check(r)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket: upgrade error: %v", err)
		return
	}
	defer conn.Close()

	conn.SetReadLimit(h.cfg.MaxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(h.cfg.PongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(h.cfg.PongWait))
		return nil
	})

	sub := &Subscriber{
		ID:      r.RemoteAddr + "-" + subject,
		Subject: subject,
		Send:    make(chan []byte, 32),
	}

	// Conn from query or path
	chName := r.URL.Query().Get("conn")
	if chName == "" {
		chName = "default"
	}

	if h.cfg.ACL != nil && !h.cfg.ACL.Allow(subject, chName) {
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.ClosePolicyViolation, "forbidden"))
		return
	}

	if h.cfg.Fanout != nil {
		h.cfg.Fanout.Subscribe(chName, sub)
		defer h.cfg.Fanout.Unsubscribe(chName, sub.ID)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(h.cfg.PingInterval)
		defer ticker.Stop()
		for {
			select {
			case msg, ok := <-sub.Send:
				_ = conn.SetWriteDeadline(time.Now().Add(h.cfg.WriteWait))
				if !ok {
					_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(h.cfg.WriteWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Reader
	go func() {
		defer wg.Done()
		defer close(sub.Send)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// Echo / publish back
			if h.cfg.Fanout != nil {
				h.cfg.Fanout.Publish(chName, message)
			}
		}
	}()

	wg.Wait()
}
