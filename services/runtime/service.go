package runtime

import (
	"log"
	"log/slog"

	"github.com/pafthang/arcanum/pkg/lifecycle"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/runtime/internal/apis"
	"github.com/pafthang/arcanum/services/runtime/internal/config"
	"github.com/pafthang/arcanum/services/runtime/internal/docker"
	"github.com/pafthang/arcanum/services/runtime/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Run starts the runtime mini service.
func Run() {
	app, cfg := svcutil.BootstrapWithConfig("runtime", Version, "Dev machines", config.FromEnv)
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
	if eng := docker.New(cfg.DockerHost); eng != nil {
		deps.Docker = eng
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
		deps.Docker = nil
		if eng := docker.New(deps.Config.DockerHost); eng != nil {
			deps.Docker = eng
		}
		return nil
	}))

	apis.Register(app.Svc, deps)
	slog.Info("runtime service started", "data", app.DataDir, "docker", cfg.HasDocker())
	app.Wait()
}
