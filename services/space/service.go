package space

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/space/internal/apis"
	"github.com/pafthang/arcanum/services/space/internal/config"
	"github.com/pafthang/arcanum/services/space/internal/store"
)

// Run starts the space mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("space", "0.1.0", "Identity and tenancy", config.FromEnv)
	defer app.Shutdown()

	dbStore, err := store.OpenStore(app.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() { _ = dbStore.Close() }()

	if err := store.Seed(dbStore, cfg.SeedPassword); err != nil {
		log.Fatalf("seed: %v", err)
	}

	deps := &apis.Deps{
		Store:  dbStore,
		NC:     app.NC,
		Config: cfg,
	}

	app.WireLifecycle(lifecycle.ReloaderFunc(func() error {
		_ = dbStore.Close()
		s, err := store.OpenStore(app.DataDir)
		if err != nil {
			return err
		}
		dbStore = s
		if err := store.Seed(dbStore, config.FromEnv().SeedPassword); err != nil {
			return err
		}
		deps.Store = s
		deps.Config = config.FromEnv()
		return nil
	}))

	apis.Register(app.Svc, deps)

	slog.Info("space service started",
		"data", app.DataDir,
		"version", app.Version,
	)
	app.Wait()
}
