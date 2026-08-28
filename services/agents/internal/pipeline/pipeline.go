package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pafthang/arcanum/services/agents/internal/providers"
	"github.com/pafthang/arcanum/services/agents/internal/store"
	"github.com/pafthang/arcanum/services/agents/internal/tools"
	"github.com/pafthang/arcanum/services/agents/models"
)

// Stage names match GoClaw's 8-stage agent loop.
const (
	StageContext   = "context"
	StageHistory   = "history"
	StagePrompt    = "prompt"
	StageThink     = "think"
	StageAct       = "act"
	StageObserve   = "observe"
	StageMemory    = "memory"
	StageSummarize = "summarize"
)

// DefaultStages is the always-on execution path.
var DefaultStages = []string{
	StageContext,
	StageHistory,
	StagePrompt,
	StageThink,
	StageAct,
	StageObserve,
	StageMemory,
	StageSummarize,
}

// State is mutable run state shared across stages.
type State struct {
	Run      *models.Run
	Session  *models.Session
	Stages   []string
	Notes    map[string]string
	Memories []models.Memory
	Skills   []models.Skill
	Output   string
}

// Runner executes stages against the agents store.
type Runner struct {
	Store    *store.Store
	Provider providers.Provider
	Tools    *tools.Host
	MaxSteps int
}

// Execute runs the pipeline for an existing queued/running run.
func (r *Runner) Execute(ctx context.Context, run *models.Run) (*models.Run, error) {
	if r == nil || r.Store == nil {
		return nil, fmt.Errorf("pipeline runner not configured")
	}
	if run == nil {
		return nil, fmt.Errorf("run required")
	}

	cur, err := r.Store.GetRun(ctx, run.ID)
	if err != nil {
		return nil, err
	}
	if cur == nil {
		return nil, fmt.Errorf("run not found")
	}
	if cur.Status == models.StatusCancelling {
		return r.Store.FinishRun(ctx, cur.ID, models.StatusCancelled, "", "cancelled")
	}

	cur, err = r.Store.MarkRunning(ctx, cur.ID)
	if err != nil {
		return nil, err
	}
	if cur.Status == models.StatusCancelling {
		return r.Store.FinishRun(ctx, cur.ID, models.StatusCancelled, "", "cancelled")
	}

	st := &State{
		Run:    cur,
		Stages: append([]string{}, DefaultStages...),
		Notes:  map[string]string{},
	}
	st.Memories, _ = r.Store.ListMemories(ctx, cur.SpaceID, cur.AgentID)
	st.Skills, _ = r.Store.ListSkills(ctx, cur.SpaceID)

	for _, name := range st.Stages {
		live, err := r.Store.GetRun(ctx, cur.ID)
		if err != nil {
			return nil, err
		}
		if live != nil && (live.Status == models.StatusCancelling || live.Status == models.StatusCancelled) {
			return r.Store.FinishRun(ctx, cur.ID, models.StatusCancelled, st.Output, "cancelled")
		}
		if err := r.step(ctx, name, st); err != nil {
			payload, _ := json.Marshal(st.Notes)
			_, _ = r.Store.UpsertSession(ctx, cur.ID, cur.SpaceID, name, string(payload))
			return r.Store.FinishRun(ctx, cur.ID, models.StatusFailed, st.Output, err.Error())
		}
		payload, _ := json.Marshal(st.Notes)
		sess, err := r.Store.UpsertSession(ctx, cur.ID, cur.SpaceID, name, string(payload))
		if err != nil {
			return r.Store.FinishRun(ctx, cur.ID, models.StatusFailed, st.Output, err.Error())
		}
		st.Session = sess
	}

	return r.Store.FinishRun(ctx, cur.ID, models.StatusSucceeded, st.Output, "")
}

func (r *Runner) step(ctx context.Context, name string, st *State) error {
	switch name {
	case StageContext:
		st.Notes[name] = fmt.Sprintf("space=%s agent=%s issue=%s", st.Run.SpaceID, st.Run.AgentID, st.Run.IssueID)
	case StageHistory:
		st.Notes[name] = fmt.Sprintf("memories=%d", len(st.Memories))
	case StagePrompt:
		var names []string
		for _, sk := range st.Skills {
			names = append(names, sk.Name)
		}
		st.Notes[name] = fmt.Sprintf("skills=%s input=%d", strings.Join(names, ","), len(st.Run.Input))
	case StageThink:
		return r.thinkAct(ctx, st)
	case StageAct, StageObserve:
		if st.Notes[StageThink] == "" {
			return r.thinkAct(ctx, st)
		}
	case StageMemory:
		if strings.TrimSpace(st.Run.Input) != "" {
			_, err := r.Store.PutMemory(ctx, st.Run.SpaceID, st.Run.AgentID, models.TierWorking, "last_input", st.Run.Input)
			if err != nil {
				return err
			}
		}
		st.Notes[name] = models.TierWorking
	case StageSummarize:
		if strings.TrimSpace(st.Output) == "" {
			st.Output = summarize(st)
		}
		st.Notes[name] = st.Output
	default:
		return fmt.Errorf("unknown stage %s", name)
	}
	return nil
}

func summarize(st *State) string {
	in := strings.TrimSpace(st.Run.Input)
	if in == "" {
		in = "(empty)"
	}
	issue := st.Run.IssueID
	if issue == "" {
		issue = "-"
	}
	return fmt.Sprintf("run %s agent %s issue %s: %s", st.Run.ID, st.Run.AgentID, issue, in)
}
