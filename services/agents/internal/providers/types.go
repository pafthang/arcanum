package providers

import "context"

// Provider is an OpenAI-compatible chat backend (GoClaw-shaped, smaller).
type Provider interface {
	Name() string
	DefaultModel() string
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
}

// ChatRequest is one model turn.
type ChatRequest struct {
	Messages []Message
	Tools    []ToolDefinition
	Model    string
}

// ChatResponse is one model reply.
type ChatResponse struct {
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
}

// Message is a chat message.
type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Name       string
}

// ToolCall is a function invocation requested by the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// ToolDefinition is an OpenAI function tool schema.
type ToolDefinition struct {
	Type     string
	Function ToolFunction
}

// ToolFunction describes a callable tool.
type ToolFunction struct {
	Name        string
	Description string
	Parameters  map[string]any
}

// FunctionTool builds a function tool definition.
func FunctionTool(name, desc string, params map[string]any) ToolDefinition {
	if params == nil {
		params = map[string]any{"type": "object", "properties": map[string]any{}}
	}
	return ToolDefinition{Type: "function", Function: ToolFunction{Name: name, Description: desc, Parameters: params}}
}
