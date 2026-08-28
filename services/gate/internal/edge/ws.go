package edge

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/gate/internal/auth"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  16 * 1024,
	WriteBufferSize: 16 * 1024,
	CheckOrigin: func(r *http.Request) bool {
		// CORS middleware already filters; WS origins are permissive in dev.
		return true
	},
}

// isWebSocketUpgrade reports a browser WS handshake.
func isWebSocketUpgrade(r *http.Request) bool {
	if r == nil || !strings.EqualFold(r.Method, http.MethodGet) {
		return false
	}
	if !strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade") {
		return false
	}
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

// serveWS upgrades the connection and bridges NATS ↔ WebSocket per catalog route.
//
//	Subscribe subject → WS (server push)
//	Publish subject   ← WS (client frames); empty = read-only
func (h *Handler) serveWS(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.Table == nil || h.NC == nil {
		http.Error(w, "gate not ready", http.StatusServiceUnavailable)
		return
	}

	route, params, ok := h.Table.MatchWS(r.URL.Path)
	if !ok || route == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	// Auth (browser cannot set Authorization on WS — accept ?access_token=)
	mode := strings.ToLower(strings.TrimSpace(route.Auth))
	if mode == "" || mode == "public" {
		mode = mini.AuthRequired
	}
	if mode == "public" {
		mode = mini.AuthNone
	}

	var ident *auth.Identity
	switch mode {
	case mini.AuthNone:
		// skip
	case mini.AuthOptional:
		ident, _ = h.tryAuth(withAccessTokenAsBearer(r))
	default:
		var err error
		ident, err = h.requireAuth(withAccessTokenAsBearer(r))
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}

	// Path spaceId must match JWT active workspace (platform_admin may cross).
	pathSpace := strings.TrimSpace(params["spaceId"])
	if pathSpace != "" {
		if ident != nil && !isPlatformAdmin(ident) {
			jwtSpace := claimStr(ident, "space_id")
			if jwtSpace != "" && jwtSpace != pathSpace {
				http.Error(w, "path space does not match active workspace", http.StatusForbidden)
				return
			}
		}
	}

	subTpl := strings.TrimSpace(route.Subscribe)
	if subTpl == "" {
		subTpl = strings.TrimSpace(route.Subject)
	}
	pubTpl := strings.TrimSpace(route.Publish)
	subSubject := mini.ExpandSubject(subTpl, params)
	pubSubject := mini.ExpandSubject(pubTpl, params)
	if subSubject == "" {
		http.Error(w, "ws route misconfigured", http.StatusBadGateway)
		return
	}

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		if h.Log != nil {
			h.Log.Debug("ws upgrade failed", "path", r.URL.Path, "err", err)
		}
		return
	}
	defer conn.Close()

	const (
		pongWait  = 60 * time.Second
		pingEvery = 30 * time.Second
		writeWait = 10 * time.Second
	)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	conn.SetReadLimit(512 * 1024)

	// NATS → WS
	outCh := make(chan []byte, 64)
	natsSub, err := h.NC.Subscribe(subSubject, func(msg *nats.Msg) {
		if msg == nil {
			return
		}
		// non-blocking to avoid stalling NATS callback
		select {
		case outCh <- append([]byte(nil), msg.Data...):
		default:
			// drop if client is slow
		}
	})
	if err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "nats subscribe failed"))
		return
	}
	defer func() { _ = natsSub.Unsubscribe() }()

	var wg sync.WaitGroup
	done := make(chan struct{})

	// Writer + ping
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(pingEvery)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case msg, ok := <-outCh:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if !ok {
					_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					return
				}
			case <-ticker.C:
				_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// Reader: WS → NATS publish (when configured)
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if pubSubject == "" {
			continue // read-only stream
		}
		if err := h.NC.Publish(pubSubject, message); err != nil {
			if h.Log != nil {
				h.Log.Debug("ws publish failed", "subject", pubSubject, "err", err)
			}
		}
	}

	close(done)
	wg.Wait()
	if h.Log != nil {
		h.Log.Debug("ws closed",
			"path", r.URL.Path,
			"sub", subSubject,
			"pub", pubSubject,
			"user", identitySubject(ident),
		)
	}
}

// withAccessTokenAsBearer copies ?access_token= / ?token= into Authorization
// so existing authenticators work for browser WebSockets.
func withAccessTokenAsBearer(r *http.Request) *http.Request {
	if r == nil {
		return r
	}
	if r.Header.Get("Authorization") != "" {
		return r
	}
	tok := strings.TrimSpace(r.URL.Query().Get("access_token"))
	if tok == "" {
		tok = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if tok == "" {
		return r
	}
	// Shallow clone with new header map so we don't mutate the caller's request permanently
	// beyond this auth attempt (clone request).
	nr := r.Clone(r.Context())
	nr.Header = r.Header.Clone()
	nr.Header.Set("Authorization", "Bearer "+tok)
	return nr
}

func isPlatformAdmin(id *auth.Identity) bool {
	if id == nil {
		return false
	}
	role := strings.ToLower(claimStr(id, "platform_role"))
	return role == "platform_admin" || role == "admin"
}

func claimStr(id *auth.Identity, key string) string {
	if id == nil || id.Claims == nil {
		return ""
	}
	v, ok := id.Claims[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return strings.TrimSpace(claimString(v))
	}
}

func identitySubject(id *auth.Identity) string {
	if id == nil {
		return ""
	}
	return id.Subject
}
