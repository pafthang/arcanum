package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/services/agents/internal/providers"
	"github.com/pafthang/arcanum/services/agents/internal/store"
	"github.com/pafthang/arcanum/services/agents/models"
	workclient "github.com/pafthang/arcanum/services/work/client"
)

// Host runs in-process tools. No Docker socket (that is runtime).
type Host struct {
	Store *store.Store
	Work  *workclient.Client
}

// Definitions is the schema advertised to the model.
func Definitions() []providers.ToolDefinition {
	obj := func(props map[string]any, required []string) map[string]any {
		m := map[string]any{"type": "object", "properties": props}
		if len(required) > 0 {
			m["required"] = required
		}
		return m
	}
	str := map[string]any{"type": "string"}
	return []providers.ToolDefinition{
		providers.FunctionTool("memory_search", "Search agent memory by substring.", obj(map[string]any{"q": str}, []string{"q"})),
		providers.FunctionTool("memory_put", "Write a memory key.", obj(map[string]any{"key": str, "value": str, "tier": str}, []string{"key", "value"})),
		providers.FunctionTool("skill_list", "List skill names and bodies in this space.", obj(map[string]any{}, nil)),
		providers.FunctionTool("work_get_issue", "Fetch an issue in this space by id.", obj(map[string]any{"issueId": str}, []string{"issueId"})),
	}
}

// Exec runs one tool in the run's space/agent.
func (h *Host) Exec(ctx context.Context, run *models.Run, name, argsJSON string) string {
	var args map[string]any
	_ = json.Unmarshal([]byte(argsJSON), &args)
	if args == nil {
		args = map[string]any{}
	}
	str := func(k string) string {
		v, _ := args[k].(string)
		return strings.TrimSpace(v)
	}
	if h == nil || h.Store == nil || run == nil {
		return "{\"error\":\"tools unavailable\"}"
	}
	switch name {
	case "memory_search":
		items, err := h.Store.SearchMemories(ctx, run.SpaceID, run.AgentID, str("q"))
		if err != nil {
			return errJSON(err)
		}
		return mustJSON(items)
	case "memory_put":
		tier := str("tier")
		if tier == "" {
			tier = models.TierWorking
		}
		m, err := h.Store.PutMemory(ctx, run.SpaceID, run.AgentID, tier, str("key"), str("value"))
		if err != nil {
			return errJSON(err)
		}
		return mustJSON(m)
	case "skill_list":
		items, err := h.Store.ListSkills(ctx, run.SpaceID)
		if err != nil {
			return errJSON(err)
		}
		return mustJSON(items)
	case "work_get_issue":
		id := str("issueId")
		if id == "" {
			id = run.IssueID
		}
		if h.Work == nil {
			return "{\"error\":\"work unavailable\"}"
		}
		iss, err := h.Work.GetIssue(ctx, id, run.SpaceID)
		if err != nil {
			return errJSON(err)
		}
		return mustJSON(iss)
	default:
		return fmt.Sprintf("{\"error\":\"unknown tool %s\"}", name)
	}
}

func errJSON(err error) string {
	return mustJSON(map[string]string{"error": err.Error()})
}

func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		return "{\"error\":\"marshal\"}"
	}
	return string(b)
}
