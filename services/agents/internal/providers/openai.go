package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAI talks to any OpenAI-compatible /v1/chat/completions endpoint.
type OpenAI struct {
	NameValue  string
	BaseURL    string
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAI builds a provider. BaseURL empty means https://api.openai.com/v1.
func NewOpenAI(baseURL, apiKey, model string) *OpenAI {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAI{
		NameValue:  "openai",
		BaseURL:    baseURL,
		APIKey:     apiKey,
		Model:      model,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenAI) Name() string         { return p.NameValue }
func (p *OpenAI) DefaultModel() string { return p.Model }

func (p *OpenAI) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := strings.TrimSpace(req.Model)
	if model == "" {
		model = p.Model
	}
	body := map[string]any{
		"model":    model,
		"messages": encodeMessages(req.Messages),
	}
	if len(req.Tools) > 0 {
		body["tools"] = encodeTools(req.Tools)
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	res, err := p.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	data, _ := io.ReadAll(res.Body)
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("openai %d: %s", res.StatusCode, truncate(data))
	}
	return parseChat(data)
}

func encodeMessages(in []Message) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, m := range in {
		row := map[string]any{"role": m.Role, "content": m.Content}
		if m.ToolCallID != "" {
			row["tool_call_id"] = m.ToolCallID
		}
		if m.Name != "" {
			row["name"] = m.Name
		}
		if len(m.ToolCalls) > 0 {
			calls := make([]map[string]any, 0, len(m.ToolCalls))
			for _, tc := range m.ToolCalls {
				calls = append(calls, map[string]any{
					"id":   tc.ID,
					"type": "function",
					"function": map[string]any{
						"name":      tc.Name,
						"arguments": tc.Arguments,
					},
				})
			}
			row["tool_calls"] = calls
		}
		out = append(out, row)
	}
	return out
}

func encodeTools(in []ToolDefinition) []map[string]any {
	out := make([]map[string]any, 0, len(in))
	for _, t := range in {
		out = append(out, map[string]any{
			"type": "function",
			"function": map[string]any{
				"name":        t.Function.Name,
				"description": t.Function.Description,
				"parameters":  t.Function.Parameters,
			},
		})
	}
	return out
}

func parseChat(data []byte) (*ChatResponse, error) {
	var raw struct {
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	if raw.Error != nil && raw.Error.Message != "" {
		return nil, fmt.Errorf("openai: %s", raw.Error.Message)
	}
	if len(raw.Choices) == 0 {
		return nil, fmt.Errorf("openai: empty choices")
	}
	ch := raw.Choices[0]
	out := &ChatResponse{Content: strings.TrimSpace(ch.Message.Content), FinishReason: ch.FinishReason}
	for _, tc := range ch.Message.ToolCalls {
		out.ToolCalls = append(out.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments})
	}
	if len(out.ToolCalls) > 0 && out.FinishReason == "" {
		out.FinishReason = "tool_calls"
	}
	return out, nil
}

func truncate(b []byte) string {
	s := strings.TrimSpace(string(b))
	if len(s) > 240 {
		return s[:240]
	}
	return s
}
