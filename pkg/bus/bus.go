// Package bus provides optional JetStream helpers with core NATS fallback.
//
// Enable with NATS JetStream (nats-server -js). If JetStream is unavailable,
// Subscribe falls back to core NATS queue subscriptions.
//
// Stream flavours:
//
//	MICRO_EVENTS       — LimitsPolicy, subjects events.> (fanout / replay)
//	MICRO_COMMANDS     — WorkQueuePolicy, subjects commands.> (work + retry)
//	MICRO_COMMANDS_DLQ — LimitsPolicy, subjects commands.dlq.> (dead letter)
package bus

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	// StreamName is the shared domain events stream.
	StreamName = "MICRO_EVENTS"
	// DefaultSubjects captured by the stream.
	DefaultSubjects = "events.>"

	// CommandStreamName is the work-queue stream for worker commands.
	CommandStreamName = "MICRO_COMMANDS"
	// DefaultCommandSubjects captured by the command stream.
	DefaultCommandSubjects = "commands.>"

	// CommandDLQStreamName stores exhausted / poison command messages.
	CommandDLQStreamName = "MICRO_COMMANDS_DLQ"
	// DefaultCommandDLQSubjects for dead-lettered commands.
	DefaultCommandDLQSubjects = "commands.dlq.>"
)

// EnsureStream creates (or updates) the MICRO_EVENTS stream.
func EnsureStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream(nats.MaxWait(3 * time.Second))
	if err != nil {
		return nil, err
	}
	subjects := strings.Split(env("BUS_STREAM_SUBJECTS", DefaultSubjects), ",")
	cfg := &nats.StreamConfig{
		Name:       env("BUS_STREAM_NAME", StreamName),
		Subjects:   subjects,
		Storage:    nats.FileStorage,
		Retention:  nats.LimitsPolicy,
		MaxAge:     7 * 24 * time.Hour,
		Duplicates: 2 * time.Minute,
	}
	if err := ensureStreamCfg(js, cfg); err != nil {
		return js, err
	}
	return js, nil
}

// EnsureCommandStream creates (or updates) MICRO_COMMANDS with WorkQueue retention.
func EnsureCommandStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream(nats.MaxWait(3 * time.Second))
	if err != nil {
		return nil, err
	}
	subjects := strings.Split(env("BUS_COMMAND_SUBJECTS", DefaultCommandSubjects), ",")
	cfg := &nats.StreamConfig{
		Name:       env("BUS_COMMAND_STREAM", CommandStreamName),
		Subjects:   subjects,
		Storage:    nats.FileStorage,
		Retention:  nats.WorkQueuePolicy,
		MaxAge:     24 * time.Hour,
		Duplicates: 2 * time.Minute,
	}
	if err := ensureStreamCfg(js, cfg); err != nil {
		return js, err
	}
	return js, nil
}

// EnsureCommandDLQStream creates the dead-letter stream for commands.
func EnsureCommandDLQStream(nc *nats.Conn) (nats.JetStreamContext, error) {
	js, err := nc.JetStream(nats.MaxWait(3 * time.Second))
	if err != nil {
		return nil, err
	}
	subjects := strings.Split(env("BUS_COMMAND_DLQ_SUBJECTS", DefaultCommandDLQSubjects), ",")
	cfg := &nats.StreamConfig{
		Name:      env("BUS_COMMAND_DLQ_STREAM", CommandDLQStreamName),
		Subjects:  subjects,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour,
	}
	if err := ensureStreamCfg(js, cfg); err != nil {
		return js, err
	}
	return js, nil
}

func ensureStreamCfg(js nats.JetStreamContext, cfg *nats.StreamConfig) error {
	if _, err := js.StreamInfo(cfg.Name); err != nil {
		if _, err := js.AddStream(cfg); err != nil {
			return fmt.Errorf("add stream %s: %w", cfg.Name, err)
		}
		slog.Info("bus: stream created", "name", cfg.Name, "subjects", cfg.Subjects, "retention", fmt.Sprintf("%v", cfg.Retention))
	} else {
		if _, err := js.UpdateStream(cfg); err != nil {
			slog.Debug("bus: update stream", "name", cfg.Name, "err", err)
		}
	}
	return nil
}

// Handler processes a NATS message. For JetStream event subscribe, Ack is called on return.
type Handler func(msg *nats.Msg)

// CommandHandler processes a work command. Return non-nil to request retry (Nak).
// After MaxDeliver attempts the message is dead-lettered and terminated.
type CommandHandler func(msg *nats.Msg) error

// CommandOpts configures work-queue consumers.
type CommandOpts struct {
	// MaxDeliver is max redeliveries before DLQ (default 5, min 1).
	MaxDeliver int
	// AckWait before redelivery (default 2m).
	AckWait time.Duration
	// DLQSubject receives poison messages. Empty disables DLQ publish.
	// When unset in opts, defaults to commands.dlq.<rest of filter>.
	DLQSubject string
	// DisableDLQ skips dead-letter publishing (still Term after max deliver).
	DisableDLQ bool
	// NakDelay base delay for retries (actual = NakDelay * delivery#). Default 2s.
	NakDelay time.Duration
}

func defaultCommandOpts(filterSubject string) CommandOpts {
	max := envInt("BUS_COMMAND_MAX_DELIVER", 5)
	if max < 1 {
		max = 1
	}
	return CommandOpts{
		MaxDeliver: max,
		AckWait:    envDuration("BUS_COMMAND_ACK_WAIT", 2*time.Minute),
		NakDelay:   envDuration("BUS_COMMAND_NAK_DELAY", 2*time.Second),
		DLQSubject: dlqSubjectFor(filterSubject),
	}
}

// Subscribe binds a durable consumer when JetStream is up; otherwise core queue sub.
func Subscribe(nc *nats.Conn, filterSubject, durable string, h Handler) (*nats.Subscription, error) {
	js, err := EnsureStream(nc)
	if err != nil || js == nil || os.Getenv("BUS_FORCE_CORE") == "1" {
		slog.Info("bus: core NATS subscribe", "subject", filterSubject, "queue", durable)
		return nc.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) { h(msg) })
	}

	sub, err := js.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) {
		h(msg)
		_ = msg.Ack()
	}, nats.Durable(durable), nats.ManualAck(), nats.AckExplicit(), nats.DeliverAll())
	if err != nil {
		slog.Warn("bus: jetstream subscribe failed, falling back to core", "err", err)
		return nc.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) { h(msg) })
	}
	slog.Info("bus: jetstream subscribe", "subject", filterSubject, "durable", durable)
	return sub, nil
}

// SubscribeCommand binds a WorkQueue consumer with retry + optional DLQ.
//
//	h returns error → NakWithDelay until MaxDeliver, then DLQ + Term.
//	h returns nil → Ack (removed from work queue).
func SubscribeCommand(nc *nats.Conn, filterSubject, durable string, h CommandHandler, opts ...CommandOpts) (*nats.Subscription, error) {
	o := defaultCommandOpts(filterSubject)
	if len(opts) > 0 {
		in := opts[0]
		if in.MaxDeliver > 0 {
			o.MaxDeliver = in.MaxDeliver
		}
		if in.AckWait > 0 {
			o.AckWait = in.AckWait
		}
		if in.NakDelay > 0 {
			o.NakDelay = in.NakDelay
		}
		if in.DisableDLQ {
			o.DisableDLQ = true
			o.DLQSubject = ""
		} else if in.DLQSubject != "" {
			o.DLQSubject = in.DLQSubject
		}
	}

	handle := func(msg *nats.Msg, jsMode bool) {
		err := safeCommand(h, msg)
		if !jsMode {
			if err != nil {
				slog.Warn("bus: command failed (core, no retry)", "subject", msg.Subject, "err", err)
			}
			return
		}
		if err == nil {
			_ = msg.Ack()
			return
		}
		delivered := uint64(1)
		if meta, merr := msg.Metadata(); merr == nil && meta != nil {
			delivered = meta.NumDelivered
		}
		if int(delivered) >= o.MaxDeliver {
			slog.Error("bus: command exhausted deliveries → DLQ",
				"subject", msg.Subject, "delivered", delivered, "max", o.MaxDeliver, "err", err)
			if !o.DisableDLQ && o.DLQSubject != "" {
				_ = PublishDLQ(nc, o.DLQSubject, msg.Subject, msg.Data, err, int(delivered))
			}
			if terr := msg.Term(); terr != nil {
				_ = msg.Ack()
			}
			return
		}
		delay := o.NakDelay * time.Duration(delivered)
		if delay <= 0 {
			delay = 2 * time.Second
		}
		slog.Warn("bus: command retry", "subject", msg.Subject, "delivered", delivered, "delay", delay, "err", err)
		if nerr := msg.NakWithDelay(delay); nerr != nil {
			_ = msg.Nak()
		}
	}

	if os.Getenv("BUS_FORCE_CORE") == "1" {
		slog.Info("bus: core command subscribe", "subject", filterSubject, "queue", durable)
		return nc.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) { handle(msg, false) })
	}
	js, err := EnsureCommandStream(nc)
	if err != nil || js == nil {
		slog.Info("bus: core command subscribe (no jetstream)", "subject", filterSubject)
		return nc.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) { handle(msg, false) })
	}
	if !o.DisableDLQ && o.DLQSubject != "" {
		_, _ = EnsureCommandDLQStream(nc)
	}

	sub, err := js.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) {
		handle(msg, true)
	},
		nats.Durable(durable),
		nats.ManualAck(),
		nats.AckExplicit(),
		nats.DeliverAll(),
		nats.AckWait(o.AckWait),
		nats.MaxDeliver(o.MaxDeliver),
	)
	if err != nil {
		slog.Warn("bus: command jetstream subscribe failed, core fallback", "err", err)
		return nc.QueueSubscribe(filterSubject, durable, func(msg *nats.Msg) { handle(msg, false) })
	}
	slog.Info("bus: command workqueue subscribe",
		"subject", filterSubject, "durable", durable, "maxDeliver", o.MaxDeliver, "dlq", o.DLQSubject)
	return sub, nil
}

func safeCommand(h CommandHandler, msg *nats.Msg) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("panic: %v", rec)
		}
	}()
	return h(msg)
}

// DLQEnvelope is stored on the dead-letter subject.
type DLQEnvelope struct {
	OriginalSubject string `json:"originalSubject"`
	Delivered       int    `json:"delivered"`
	Error           string `json:"error"`
	TS              string `json:"ts"`
	// Payload is the original command body (base64 in JSON).
	Payload []byte `json:"payload"`
}

// PublishDLQ writes a poison message to the DLQ stream/subject.
func PublishDLQ(nc *nats.Conn, dlqSubject, originalSubject string, payload []byte, cause error, delivered int) error {
	env := DLQEnvelope{
		OriginalSubject: originalSubject,
		Delivered:       delivered,
		TS:              time.Now().UTC().Format(time.RFC3339Nano),
		Payload:         payload,
	}
	if cause != nil {
		env.Error = cause.Error()
	}
	raw, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if os.Getenv("BUS_FORCE_CORE") != "1" {
		if js, err := EnsureCommandDLQStream(nc); err == nil && js != nil {
			if _, err := js.Publish(dlqSubject, raw); err == nil {
				return nil
			}
		}
	}
	return nc.Publish(dlqSubject, raw)
}

// Publish publishes to core NATS (and JetStream if stream captures the subject).
func Publish(nc *nats.Conn, subject string, data []byte) error {
	js, err := nc.JetStream(nats.MaxWait(2 * time.Second))
	if err == nil {
		if _, err := js.Publish(subject, data); err == nil {
			return nil
		}
	}
	return nc.Publish(subject, data)
}

// PublishCommand publishes a worker command (prefers MICRO_COMMANDS JetStream).
func PublishCommand(nc *nats.Conn, subject string, data []byte) error {
	if os.Getenv("BUS_FORCE_CORE") != "1" {
		js, err := EnsureCommandStream(nc)
		if err == nil && js != nil {
			if _, err := js.Publish(subject, data); err == nil {
				return nil
			}
			slog.Debug("bus: command js publish failed", "err", err)
		}
	}
	return nc.Publish(subject, data)
}

// DLQRecord is one retained dead-letter message.
type DLQRecord struct {
	Seq             uint64    `json:"seq"`
	Subject         string    `json:"subject"`
	Time            time.Time `json:"time"`
	OriginalSubject string    `json:"originalSubject"`
	Delivered       int       `json:"delivered"`
	Error           string    `json:"error"`
	TS              string    `json:"ts"`
	Payload         []byte    `json:"payload,omitempty"`
	// Parsed is the original command payload when it was JSON.
	Parsed any `json:"parsed,omitempty"`
}

// ListDLQ returns the newest dead-letter messages (up to limit).
func ListDLQ(nc *nats.Conn, filterSubject string, limit int) ([]DLQRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	js, err := EnsureCommandDLQStream(nc)
	if err != nil {
		return nil, err
	}
	stream := env("BUS_COMMAND_DLQ_STREAM", CommandDLQStreamName)
	info, err := js.StreamInfo(stream)
	if err != nil {
		return nil, err
	}
	if info.State.Msgs == 0 {
		return []DLQRecord{}, nil
	}
	// Walk from last sequence backwards.
	var out []DLQRecord
	last := info.State.LastSeq
	first := info.State.FirstSeq
	for seq := last; seq >= first && len(out) < limit; seq-- {
		rmsg, err := js.GetMsg(stream, seq)
		if err != nil {
			continue
		}
		if filterSubject != "" && rmsg.Subject != filterSubject {
			continue
		}
		rec := DLQRecord{
			Seq:     rmsg.Sequence,
			Subject: rmsg.Subject,
			Time:    rmsg.Time,
		}
		var env DLQEnvelope
		if json.Unmarshal(rmsg.Data, &env) == nil && (env.OriginalSubject != "" || len(env.Payload) > 0) {
			rec.OriginalSubject = env.OriginalSubject
			rec.Delivered = env.Delivered
			rec.Error = env.Error
			rec.TS = env.TS
			rec.Payload = env.Payload
			var parsed any
			if json.Unmarshal(env.Payload, &parsed) == nil {
				rec.Parsed = parsed
			}
		} else {
			rec.Payload = rmsg.Data
			var parsed any
			if json.Unmarshal(rmsg.Data, &parsed) == nil {
				rec.Parsed = parsed
			}
		}
		out = append(out, rec)
		if seq == 0 {
			break
		}
	}
	return out, nil
}

// RequeueDLQ republishes the original payload onto the work queue and deletes the DLQ entry.
// targetSubject overrides env.OriginalSubject when non-empty.
func RequeueDLQ(nc *nats.Conn, seq uint64, targetSubject string) (DLQRecord, error) {
	var zero DLQRecord
	js, err := EnsureCommandDLQStream(nc)
	if err != nil {
		return zero, err
	}
	stream := env("BUS_COMMAND_DLQ_STREAM", CommandDLQStreamName)
	rmsg, err := js.GetMsg(stream, seq)
	if err != nil {
		return zero, fmt.Errorf("get dlq msg %d: %w", seq, err)
	}
	rec := DLQRecord{Seq: rmsg.Sequence, Subject: rmsg.Subject, Time: rmsg.Time}
	var env DLQEnvelope
	payload := rmsg.Data
	orig := targetSubject
	if json.Unmarshal(rmsg.Data, &env) == nil && len(env.Payload) > 0 {
		payload = env.Payload
		rec.OriginalSubject = env.OriginalSubject
		rec.Delivered = env.Delivered
		rec.Error = env.Error
		rec.TS = env.TS
		rec.Payload = env.Payload
		if orig == "" {
			orig = env.OriginalSubject
		}
	} else {
		rec.Payload = payload
	}
	if orig == "" {
		// fallback: strip dlq prefix
		if strings.HasPrefix(rmsg.Subject, "commands.dlq.") {
			orig = "commands." + strings.TrimPrefix(rmsg.Subject, "commands.dlq.")
		} else {
			return zero, fmt.Errorf("cannot determine original subject")
		}
	}
	if err := PublishCommand(nc, orig, payload); err != nil {
		return zero, err
	}
	if err := js.DeleteMsg(stream, seq); err != nil {
		slog.Warn("bus: requeue ok but delete dlq failed", "seq", seq, "err", err)
	}
	rec.OriginalSubject = orig
	var parsed any
	if json.Unmarshal(payload, &parsed) == nil {
		rec.Parsed = parsed
	}
	return rec, nil
}

func dlqSubjectFor(filter string) string {
	// commands.exec.run → commands.dlq.exec.run
	if strings.HasPrefix(filter, "commands.") {
		return "commands.dlq." + strings.TrimPrefix(filter, "commands.")
	}
	return "commands.dlq." + filter
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func envDuration(k string, def time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Second
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return def
	}
	return d
}
