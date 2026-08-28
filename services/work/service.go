package work

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
	"github.com/pafthang/arcanum/services/work/internal/apis"
	"github.com/pafthang/arcanum/services/work/internal/config"
	"github.com/pafthang/arcanum/services/work/internal/store"
)

// Run starts the work mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("work", Version, "Issue aggregate", config.FromEnv)
	defer app.Shutdown()

	dbStore, err := store.OpenStore(app.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() { _ = dbStore.Close() }()

	var sc *spaceclient.Client
	var lc *loggclient.Client
	if app.NC != nil {
		if c, err := spaceclient.New(app.NC); err != nil {
			slog.Warn("space client unavailable", "err", err)
		} else {
			sc = c
		}
		if c, err := loggclient.New(app.NC); err != nil {
			slog.Warn("logg client unavailable", "err", err)
		} else {
			lc = c
		}
	}

	deps := &apis.Deps{
		Store:  dbStore,
		NC:     app.NC,
		Space:  sc,
		Logg:   lc,
		Config: cfg,
	}

	app.WireLifecycle(lifecycle.ReloaderFunc(func() error {
		_ = dbStore.Close()
		s, err := store.OpenStore(app.DataDir)
		if err != nil {
			return err
		}
		dbStore = s
		deps.Store = s
		deps.Config = config.FromEnv()
		return nil
	}))

	apis.Register(app.Svc, deps)

	slog.Info("work service started",
		"data", app.DataDir,
		"version", app.Version,
	)
	app.Wait()
}
