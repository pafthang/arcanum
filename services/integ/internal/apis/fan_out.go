package apis

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/events"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/integ/internal/hmacx"
	"github.com/pafthang/arcanum/services/integ/models"
)

func registerFanout(d *Deps) {
	if d.NC == nil {
		return
	}
	_, err := d.NC.Subscribe("events.work.>", func(msg *nats.Msg) {
		env, data, err := events.Decode[map[string]any](msg.Data)
		if err != nil {
			return
		}
		spaceID, _ := data["spaceId"].(string)
		if spaceID == "" {
			return
		}
		deliverOutbound(d, spaceID, env.Type, msg.Data)
	})
	if err != nil {
		slog.Error("integ: subscribe events.work.>", "err", err)
	}
}

func deliverOutbound(d *Deps, spaceID, typ string, payload []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	hooks, err := d.Store.ListActiveWebhooksForEvent(ctx, spaceID, typ)
	if err != nil || len(hooks) == 0 {
		return
	}
	client := &http.Client{Timeout: 6 * time.Second}
	for i := range hooks {
		h := hooks[i]
		status := models.DeliverySent
		lastErr := ""
		delivered := time.Now().UTC().Format(time.RFC3339Nano)
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.URL, bytes.NewReader(payload))
		if err != nil {
			status = models.DeliveryFailed
			lastErr = err.Error()
			delivered = ""
		} else {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Arcanum-Event", typ)
			if h.Secret != "" {
				req.Header.Set("X-Arcanum-Signature", hmacx.SignHex(h.Secret, payload))
			}
			resp, err := client.Do(req)
			if err != nil {
				status = models.DeliveryFailed
				lastErr = err.Error()
				delivered = ""
			} else {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= 300 {
					status = models.DeliveryFailed
					lastErr = resp.Status
					delivered = ""
				}
			}
		}
		del, recErr := d.Store.RecordDelivery(ctx, spaceID, "webhook", h.ID, typ, string(payload), status, lastErr, 1, delivered)
		if recErr != nil {
			continue
		}
		if status == models.DeliverySent {
			publish(d, subjects.EventIntegWebhookDelivered, "integ.webhook.delivered", del)
		}
	}
}
