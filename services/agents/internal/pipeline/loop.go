package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/services/agents/internal/providers"
	"github.com/pafthang/arcanum/services/agents/internal/tools"
)

const defaultMaxSteps = 8

func (r *Runner) thinkAct(ctx context.Context, st *State) error {
	if r.Provider == nil {
		st.Notes[StageThink] = "local"
		st.Notes[StageAct] = "noop"
		st.Notes[StageObserve] = st.Run.Input
		return nil
	}
	host := r.Tools
	if host == nil {
		host = &tools.Host{Store: r.Store}
	}
	max := r.MaxSteps
	if max <= 0 {
		max = defaultMaxSteps
	}
	msgs := buildMessages(st)
	defs := tools.Definitions()
	var last string
	for step := 0; step < max; step++ {
		res, err := r.Provider.Chat(ctx, providers.ChatRequest{Messages: msgs, Tools: defs})
		if err != nil {
			return err
		}
		if len(res.ToolCalls) == 0 {
			last = strings.TrimSpace(res.Content)
			st.Notes[StageThink] = r.Provider.Name()
			st.Notes[StageAct] = fmt.Sprintf("steps=%d", step+1)
			st.Notes[StageObserve] = last
			if last != "" {
				st.Output = last
			}
			return nil
		}
		msgs = append(msgs, providers.Message{Role: "assistant", Content: res.Content, ToolCalls: res.ToolCalls})
		var observed []string
		for _, tc := range res.ToolCalls {
			out := host.Exec(ctx, st.Run, tc.Name, tc.Arguments)
			observed = append(observed, tc.Name)
			msgs = append(msgs, providers.Message{
				Role:       "tool",
				Name:       tc.Name,
				ToolCallID: tc.ID,
				Content:    out,
			})
		}
		st.Notes[StageAct] = strings.Join(observed, ",")
	}
	return fmt.Errorf("tool loop exceeded %d steps", max)
}

func buildMessages(st *State) []providers.Message {
	var b strings.Builder
	b.WriteString("You are an Arcanum space agent. Use tools when they help. Stay in the current space.\n")
	if st.Run != nil && st.Run.IssueID != "" {
		b.WriteString("Bound issue: " + st.Run.IssueID + "\n")
	}
	if len(st.Skills) > 0 {
		b.WriteString("\nSkills:\n")
		for _, sk := range st.Skills {
			b.WriteString("- " + sk.Name + "\n" + sk.Body + "\n")
		}
	}
	if len(st.Memories) > 0 {
		b.WriteString("\nMemory:\n")
		for i, m := range st.Memories {
			if i >= 12 {
				break
			}
			b.WriteString("- " + m.Key + ": " + m.Value + "\n")
		}
	}
	user := ""
	if st.Run != nil {
		user = strings.TrimSpace(st.Run.Input)
	}
	if user == "" {
		user = "Summarize available context."
	}
	return []providers.Message{
		{Role: "system", Content: b.String()},
		{Role: "user", Content: user},
	}
}
