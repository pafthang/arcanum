package logg

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/logg/internal/apis"
	"github.com/pafthang/arcanum/services/logg/internal/config"
	"github.com/pafthang/arcanum/services/logg/internal/store"
)

// Run starts the logg mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("logg", "0.1.0", "Centralized logging service", config.FromEnv)
	defer app.Shutdown()

	dbStore, err := store.OpenStore(app.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() { _ = dbStore.Close() }()

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
		deps.Store = s
		deps.Config = config.FromEnv()
		return nil
	}))

	apis.Register(app.Svc, deps)

	slog.Info("logg service started",
		"data", app.DataDir,
		"version", app.Version,
	)
	app.Wait()
}
