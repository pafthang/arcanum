package media

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/media/internal/apis"
	"github.com/pafthang/arcanum/services/media/internal/config"
	"github.com/pafthang/arcanum/services/media/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Run starts the media mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("media", Version, "Blob store", config.FromEnv)
	defer app.Shutdown()

	dbStore, err := store.OpenStore(app.DataDir)
	if err != nil {
		log.Fatalf("open store: %v", err)
	}
	defer func() { _ = dbStore.Close() }()

	var sc *spaceclient.Client
	if app.NC != nil {
		if c, err := spaceclient.New(app.NC); err != nil {
			slog.Warn("space client unavailable", "err", err)
		} else {
			sc = c
		}
	}

	deps := &apis.Deps{
		Store:  dbStore,
		NC:     app.NC,
		Space:  sc,
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

	slog.Info("media service started",
		"data", app.DataDir,
		"version", app.Version,
		"maxBytes", cfg.MaxBytes,
	)
	app.Wait()
}
