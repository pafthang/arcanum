package apis

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/httpx"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/integ/internal/hmacx"
	"github.com/pafthang/arcanum/services/integ/models"
)

var issueKeyRe = regexp.MustCompile(`\b[A-Z][A-Z0-9]+-\d+\b|#\d+\b`)

func registerHooks(svc mini.Service, d *Deps) {
	must(svc.AddEndpoint("hook_ingest", mini.HandlerFunc(func(req mini.Request) {
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		handleIngest(req, d, id, false)
	}),
		mini.WithPublicHTTP("POST", "/api/integ/hooks/{connectorId}"),
		mini.WithPublicSubject("integ", "hook.ingest"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	must(svc.AddEndpoint("hook_github", mini.HandlerFunc(func(req mini.Request) {
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		handleIngest(req, d, id, true)
	}),
		mini.WithPublicHTTP("POST", "/api/integ/hooks/github/{connectorId}"),
		mini.WithPublicSubject("integ", "hook.github"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	must(svc.AddEndpoint("hook_telegram", mini.HandlerFunc(func(req mini.Request) {
		id := strings.TrimSpace(mini.PathParam(req, "connectorId"))
		handleIngest(req, d, id, false)
	}),
		mini.WithPublicHTTP("POST", "/api/integ/hooks/telegram/{connectorId}"),
		mini.WithPublicSubject("integ", "hook.telegram"),
		mini.WithPublicAuth(mini.AuthNone),
	))

	if d.NC == nil {
		return
	}
	_, _ = d.NC.Subscribe(subjects.InternalIntegIngest, func(msg *nats.Msg) {
		var in models.IngestRequest
		if err := json.Unmarshal(msg.Data, &in); err != nil {
			respondErr(d.NC, msg, "400", "bad request")
			return
		}
		c, err := d.Store.GetConnector(context.Background(), in.ConnectorID)
		if err != nil {
			respondErr(d.NC, msg, "500", err.Error())
			return
		}
		if c == nil {
			respondErr(d.NC, msg, "404", "not found")
			return
		}
		ev, err := ingestPayload(context.Background(), d, c, in.Source, in.Payload, "")
		if err != nil {
			respondErr(d.NC, msg, "400", err.Error())
			return
		}
		respondJSON(msg, ev)
	})
}

func handleIngest(req mini.Request, d *Deps, connectorID string, github bool) {
	if connectorID == "" {
		httpx.Error(req, 400, "connectorId path required.", nil)
		return
	}
	c, err := d.Store.GetConnector(req.Context(), connectorID)
	if err != nil {
		httpx.Error(req, 500, err.Error(), nil)
		return
	}
	if c == nil {
		httpx.Error(req, 404, "Connector not found.", nil)
		return
	}
	if github && c.Kind != models.KindGitHub {
		httpx.Error(req, 400, "github hook requires a github connector.", nil)
		return
	}
	body := req.Data()
	if !verifyHook(c, req, body, github) {
		httpx.Error(req, 401, "Invalid hook signature.", nil)
		return
	}
	payload := payloadFromBody(body)
	source := c.Kind
	if github {
		source = models.KindGitHub
	}
	externalID := firstNonEmpty(
		httpx.Header(req, "X-GitHub-Delivery"),
		httpx.Header(req, "X-Request-Id"),
	)
	ev, err := ingestPayload(req.Context(), d, c, source, payload, externalID)
	if err != nil {
		httpx.Error(req, 400, err.Error(), nil)
		return
	}
	httpx.JSON(req, 202, ev)
}

func verifyHook(c *models.Connector, req mini.Request, body []byte, github bool) bool {
	if c.Secret == "" {
		return true
	}
	if github || c.Kind == models.KindGitHub {
		if hmacx.VerifyGitHub(c.Secret, httpx.Header(req, "X-Hub-Signature-256"), body) {
			return true
		}
	}
	if hmacx.VerifyBearer(c.Secret, httpx.Header(req, "Authorization")) {
		return true
	}
	if hmacx.VerifyHex(c.Secret, httpx.Header(req, "X-Arcanum-Signature"), body) {
		return true
	}
	return hmacx.VerifyHex(c.Secret, httpx.Header(req, "X-Hub-Signature-256"), body)
}

func payloadFromBody(body []byte) map[string]any {
	body = bytesTrim(body)
	if len(body) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil && m != nil {
		return m
	}
	return map[string]any{"raw": string(body)}
}

func bytesTrim(b []byte) []byte {
	return []byte(strings.TrimSpace(string(b)))
}

func ingestPayload(ctx context.Context, d *Deps, c *models.Connector, source string, payload map[string]any, externalID string) (*models.InboundEvent, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	keys := extractIssueKeys(flattenText(payload)...)
	ev, err := d.Store.RecordInbound(ctx, c.SpaceID, c.ID, source, externalID, payload, keys)
	if err != nil {
		return nil, err
	}
	publish(d, subjects.EventIntegInbound, "integ.inbound", ev)
	if c.Kind == models.KindGitHub || source == models.KindGitHub {
		act := githubActivity(c, payload, keys, externalID)
		publish(d, subjects.EventIntegGitHubActivity, "integ.github.activity", act)
	}
	if models.ChannelKind(c.Kind) {
		msg := channelMessage(c, payload, externalID)
		publish(d, subjects.EventIntegChannelMessage, "integ.channel.message", msg)
		publish(d, subjects.EventIntegMessageInbound, "integ.message.inbound", msg)
	}
	return ev, nil
}

func githubActivity(c *models.Connector, payload map[string]any, keys []string, externalID string) models.GitHubActivity {
	repo := nestedString(payload, "repository", "full_name")
	if repo == "" {
		owner := nestedString(payload, "repository", "owner", "login")
		name := nestedString(payload, "repository", "name")
		if owner != "" && name != "" {
			repo = owner + "/" + name
		}
	}
	title := firstNonEmpty(
		nestedString(payload, "pull_request", "title"),
		nestedString(payload, "issue", "title"),
		nestedString(payload, "head_commit", "message"),
		nestedString(payload, "title"),
	)
	return models.GitHubActivity{
		ConnectorID: c.ID,
		SpaceID:     c.SpaceID,
		Action:      stringField(payload, "action"),
		Repo:        repo,
		Ref:         firstNonEmpty(stringField(payload, "ref"), nestedString(payload, "pull_request", "head", "ref")),
		Title:       title,
		IssueKeys:   keys,
		ExternalID:  externalID,
	}
}

func channelMessage(c *models.Connector, payload map[string]any, externalID string) models.ChannelMessage {
	text := firstNonEmpty(
		nestedString(payload, "message", "text"),
		nestedString(payload, "text"),
		nestedString(payload, "content"),
		nestedString(payload, "event", "text"),
	)
	chatID := firstNonEmpty(
		anyString(nestedAny(payload, "message", "chat", "id")),
		stringField(payload, "chatId"),
		stringField(payload, "channel"),
	)
	user := firstNonEmpty(
		nestedString(payload, "message", "from", "username"),
		nestedString(payload, "user"),
		nestedString(payload, "author", "id"),
	)
	return models.ChannelMessage{
		ConnectorID:  c.ID,
		SpaceID:      c.SpaceID,
		Kind:         c.Kind,
		ChatID:       chatID,
		ExternalUser: user,
		Text:         text,
		ExternalID:   externalID,
	}
}

func extractIssueKeys(texts ...string) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, t := range texts {
		for _, m := range issueKeyRe.FindAllString(t, -1) {
			if _, ok := seen[m]; ok {
				continue
			}
			seen[m] = struct{}{}
			out = append(out, m)
		}
	}
	return out
}

func flattenText(v any) []string {
	var out []string
	switch t := v.(type) {
	case string:
		if strings.TrimSpace(t) != "" {
			out = append(out, t)
		}
	case map[string]any:
		for _, child := range t {
			out = append(out, flattenText(child)...)
		}
	case []any:
		for _, child := range t {
			out = append(out, flattenText(child)...)
		}
	}
	return out
}

func stringField(m map[string]any, key string) string {
	v, _ := m[key].(string)
	return strings.TrimSpace(v)
}

func nestedAny(m map[string]any, keys ...string) any {
	var cur any = m
	for _, k := range keys {
		obj, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = obj[k]
	}
	return cur
}

func nestedString(m map[string]any, keys ...string) string {
	v := nestedAny(m, keys...)
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func anyString(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case float64:
		return strconv.FormatInt(int64(t), 10)
	case int:
		return strconv.Itoa(t)
	case json.Number:
		return t.String()
	default:
		return ""
	}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}
