package mini

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/nats-io/nuid"
)

// Stream protocol headers (gate ↔ service).
const (
	// HeaderStreamPhase is begin | data | end.
	HeaderStreamPhase = "X-Mini-Stream"
	// HeaderStreamID correlates chunks of one upload.
	HeaderStreamID = "X-Mini-Stream-Id"
	// HeaderStreamSeq is the 0-based data chunk sequence.
	HeaderStreamSeq = "X-Mini-Stream-Seq"
	// HeaderStreamFilename is the original filename when known.
	HeaderStreamFilename = "X-Mini-Filename"
	// HeaderStreamSize is total content length when known (-1 if unknown).
	HeaderStreamSize = "X-Mini-Stream-Size"
	// HeaderFormPrefix prefixes multipart form field headers: X-Mini-Form-{name}.
	HeaderFormPrefix = "X-Mini-Form-"
	// HeaderFileField is the multipart file field name.
	HeaderFileField = "X-Mini-File-Field"

	StreamPhaseBegin = "begin"
	StreamPhaseData  = "data"
	StreamPhaseEnd   = "end"
)

// StreamMeta describes an inbound stream upload.
type StreamMeta struct {
	ID          string
	Filename    string
	ContentType string
	Size        int64 // -1 if unknown
	Headers     Headers
	// Form holds multipart text fields (from X-Mini-Form-*).
	Form map[string]string
}

// StreamResult is the service response after consuming a stream.
type StreamResult struct {
	Data    []byte
	Headers Headers
	// Status is optional HTTP status for the gate (via WithStatus).
	Status int
}

// StreamFunc processes a streamed upload. r is closed when the gate finishes
// sending data. The function should read until EOF (or ctx cancel).
type StreamFunc func(ctx context.Context, meta StreamMeta, r io.Reader) (*StreamResult, error)

// StreamHandler reassembles begin/data/end messages into an io.Reader for StreamFunc.
// Register with AddEndpoint(..., NewStreamHandler(fn)).
type StreamHandler struct {
	fn       StreamFunc
	mu       sync.Mutex
	sessions map[string]*streamSession
	// SessionTTL drops incomplete uploads (default 2m).
	SessionTTL time.Duration
	// MaxBytes rejects streams larger than this after reassembly buffer
	// when using pipe; 0 = unlimited (still limited by available memory/pipe).
	MaxBytes int64
}

type streamSession struct {
	id     string
	meta   StreamMeta
	pw     *io.PipeWriter
	pr     *io.PipeReader
	done   chan streamOutcome
	cancel context.CancelFunc
	seq    int
	wrote  int64
}

type streamOutcome struct {
	res *StreamResult
	err error
}

// NewStreamHandler creates a Handler for chunked gate uploads.
func NewStreamHandler(fn StreamFunc) *StreamHandler {
	return &StreamHandler{
		fn:         fn,
		sessions:   make(map[string]*streamSession),
		SessionTTL: 2 * time.Minute,
	}
}

// Handle implements Handler.
func (h *StreamHandler) Handle(req Request) {
	if h == nil || h.fn == nil {
		_ = req.Error("500", "stream handler not configured", nil)
		return
	}
	phase := req.Headers().Get(HeaderStreamPhase)
	if phase == "" {
		// Non-stream request: treat body as a one-shot stream.
		h.handleOneshot(req)
		return
	}
	switch phase {
	case StreamPhaseBegin:
		h.handleBegin(req)
	case StreamPhaseData:
		h.handleData(req)
	case StreamPhaseEnd:
		h.handleEnd(req)
	default:
		_ = req.Error("400", "unknown stream phase", nil)
	}
}

func (h *StreamHandler) handleOneshot(req Request) {
	meta := streamMetaFromHeaders(req.Headers(), nuid.Next())
	meta.Size = int64(len(req.Data()))
	ctx := req.Context()
	res, err := h.fn(ctx, meta, &bytesReader{b: req.Data()})
	if err != nil {
		_ = req.Error("500", err.Error(), nil)
		return
	}
	respondStream(req, res)
}

func (h *StreamHandler) handleBegin(req Request) {
	id := req.Headers().Get(HeaderStreamID)
	if id == "" {
		id = nuid.Next()
	}
	meta := streamMetaFromHeaders(req.Headers(), id)
	pr, pw := io.Pipe()
	ttl := h.SessionTTL
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	// Detached from request ctx: begin returns immediately; data/end continue the session.
	ctx, cancel := context.WithTimeout(context.Background(), ttl)
	sess := &streamSession{
		id:     id,
		meta:   meta,
		pw:     pw,
		pr:     pr,
		done:   make(chan streamOutcome, 1),
		cancel: cancel,
	}
	h.mu.Lock()
	// cleanup stale
	for k, s := range h.sessions {
		if s == nil {
			delete(h.sessions, k)
		}
	}
	h.sessions[id] = sess
	h.mu.Unlock()

	go func() {
		res, err := h.fn(ctx, meta, pr)
		_ = pr.Close()
		sess.done <- streamOutcome{res: res, err: err}
		// TTL cleanup
		time.AfterFunc(ttl, func() {
			h.mu.Lock()
			if h.sessions[id] == sess {
				delete(h.sessions, id)
			}
			h.mu.Unlock()
		})
	}()

	// ACK begin so gate can proceed with data chunks.
	_ = req.Respond([]byte(`{"ok":true,"stream_id":"`+id+`"}`), WithHeaders(Headers{
		"Content-Type": []string{"application/json"},
		HeaderStreamID: []string{id},
	}))
}

func (h *StreamHandler) handleData(req Request) {
	id := req.Headers().Get(HeaderStreamID)
	h.mu.Lock()
	sess := h.sessions[id]
	h.mu.Unlock()
	if sess == nil {
		_ = req.Error("404", "unknown stream id", nil)
		return
	}
	data := req.Data()
	if h.MaxBytes > 0 && sess.wrote+int64(len(data)) > h.MaxBytes {
		sess.cancel()
		_ = sess.pw.CloseWithError(fmt.Errorf("stream exceeds max bytes"))
		_ = req.Error("413", "stream too large", nil)
		return
	}
	if len(data) > 0 {
		if _, err := sess.pw.Write(data); err != nil {
			_ = req.Error("500", "stream write: "+err.Error(), nil)
			return
		}
		sess.wrote += int64(len(data))
	}
	sess.seq++
	_ = req.Respond([]byte(`{"ok":true}`))
}

func (h *StreamHandler) handleEnd(req Request) {
	id := req.Headers().Get(HeaderStreamID)
	h.mu.Lock()
	sess := h.sessions[id]
	if sess != nil {
		delete(h.sessions, id)
	}
	h.mu.Unlock()
	if sess == nil {
		_ = req.Error("404", "unknown stream id", nil)
		return
	}
	// Final data may be on END body.
	if len(req.Data()) > 0 {
		_, _ = sess.pw.Write(req.Data())
	}
	_ = sess.pw.Close()

	select {
	case out := <-sess.done:
		if out.err != nil {
			_ = req.Error("500", out.err.Error(), nil)
			return
		}
		respondStream(req, out.res)
	case <-req.Context().Done():
		sess.cancel()
		_ = req.Error("504", "stream handler timeout", nil)
	}
}

func respondStream(req Request, res *StreamResult) {
	if res == nil {
		_ = req.Respond(nil)
		return
	}
	opts := []RespondOpt{}
	if res.Headers != nil {
		opts = append(opts, WithHeaders(res.Headers))
	}
	if res.Status > 0 {
		opts = append(opts, WithStatus(res.Status))
	}
	_ = req.Respond(res.Data, opts...)
}

func streamMetaFromHeaders(h Headers, id string) StreamMeta {
	meta := StreamMeta{
		ID:          id,
		Filename:    h.Get(HeaderStreamFilename),
		ContentType: h.Get("Content-Type"),
		Size:        -1,
		Headers:     h,
		Form:        map[string]string{},
	}
	if raw := h.Get(HeaderStreamSize); raw != "" {
		var n int64
		if _, err := fmt.Sscan(raw, &n); err == nil {
			meta.Size = n
		}
	}
	for k, vals := range h {
		if len(vals) == 0 {
			continue
		}
		if len(k) > len(HeaderFormPrefix) && k[:len(HeaderFormPrefix)] == HeaderFormPrefix {
			meta.Form[k[len(HeaderFormPrefix):]] = vals[0]
		}
	}
	return meta
}

// bytesReader is a tiny non-copying reader over a byte slice.
type bytesReader struct {
	b []byte
	i int
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.i >= len(r.b) {
		return 0, io.EOF
	}
	n := copy(p, r.b[r.i:])
	r.i += n
	return n, nil
}

// FormValue returns a multipart form field from request headers (gate-injected).
func FormValue(req Request, name string) string {
	if req == nil {
		return ""
	}
	return req.Headers().Get(HeaderFormPrefix + name)
}

// Filename returns X-Mini-Filename when set by the gate.
func Filename(req Request) string {
	if req == nil {
		return ""
	}
	return req.Headers().Get(HeaderStreamFilename)
}
