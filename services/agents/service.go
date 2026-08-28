package agents

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/agents/internal/apis"
	"github.com/pafthang/arcanum/services/agents/internal/config"
	"github.com/pafthang/arcanum/services/agents/internal/pipeline"
	"github.com/pafthang/arcanum/services/agents/internal/store"
	"github.com/pafthang/arcanum/services/agents/internal/tools"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
	workclient "github.com/pafthang/arcanum/services/work/client"
)

// Run starts the agents mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("agents", Version, "Agent run, session and memory", config.FromEnv)
	defer app.Shutdown()

	dbStore, err := store.OpenStore(app.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() { _ = dbStore.Close() }()

	var sc *spaceclient.Client
	var wc *workclient.Client
	if app.NC != nil {
		if c, err := spaceclient.New(app.NC); err != nil {
			slog.Warn("space client unavailable", "err", err)
		} else {
			sc = c
		}
		if c, err := workclient.New(app.NC); err != nil {
			slog.Warn("work client unavailable", "err", err)
		} else {
			wc = c
		}
	}

	prov := cfg.Provider()
	runner := &pipeline.Runner{
		Store:    dbStore,
		Provider: prov,
		Tools:    &tools.Host{Store: dbStore, Work: wc},
		MaxSteps: cfg.LLMMaxSteps,
	}
	deps := &apis.Deps{
		Store:  dbStore,
		NC:     app.NC,
		Space:  sc,
		Runner: runner,
		Config: cfg,
	}

	app.WireLifecycle(lifecycle.ReloaderFunc(func() error {
		_ = dbStore.Close()
		s, err := store.OpenStore(app.DataDir)
		if err != nil {
			return err
		}
		dbStore = s
		next := config.FromEnv()
		deps.Store = s
		deps.Config = next
		deps.Runner = &pipeline.Runner{
			Store:    s,
			Provider: next.Provider(),
			Tools:    &tools.Host{Store: s, Work: wc},
			MaxSteps: next.LLMMaxSteps,
		}
		return nil
	}))

	apis.Register(app.Svc, deps)
	slog.Info("agents service started", "data", app.DataDir, "version", app.Version, "llm", prov != nil)
	app.Wait()
}
