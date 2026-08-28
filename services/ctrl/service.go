// Package ctrl is the platform control plane: lifecycle API, optional process
// supervision, and edge config KV + admin HTTP.
//
// Layout (edge style):
//
//	service.go   — Run (process entry; supports -up supervise)
//	export.go    — public type aliases + config helpers
//	cmd/ctrl/    — thin main → ctrl.Run()
//	internal/    — implementation
//	  apis/        mini endpoints Register (lifecycle, inventory)
//	  config/      CTRL_*, CONFIG_BUCKET, ADMIN_*
//	  models/      DTOs
//	  supervise/   process supervisor for local stack
//	  edgecfg/     JetStream KV routes / WS ACL
//	  admin/       standalone /_admin HTTP for edgecfg
//
// Entrypoint: go run ./services/ctrl/cmd/ctrl  (or ./cmd/ctrl)
package ctrl

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/ctrl/internal/apis"
	ctrlconfig "github.com/pafthang/arcanum/services/ctrl/internal/config"
	"github.com/pafthang/arcanum/services/ctrl/internal/edgecfg"
	"github.com/pafthang/arcanum/services/ctrl/internal/supervise"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
)

// Run starts the ctrl mini service and blocks until signal.
//
// Flags / env:
//
//	-up / CTRL_UP=1   — also supervise all enabled services from cfgs/
//	CFG_DIR           — config directory (default cfgs)
func Run() {
	upFlag := flag.Bool("up", false, "supervise all enabled services from cfgs/ (local stack)")
	flag.Parse()

	// Own file config first (shared secrets, NATS_URL, …).
	svcutil.LoadConfig("ctrl")
	cfg := ctrlconfig.FromEnv()
	if *upFlag {
		cfg.Up = true
	}

	var cancelSupervise context.CancelFunc
	if cfg.Up {
		var ctx context.Context
		ctx, cancelSupervise = context.WithCancel(context.Background())
		skip := map[string]struct{}{"ctrl": {}}
		if cfg.SkipNATS {
			skip["nats"] = struct{}{}
			slog.Info("ctrl: not supervising nats (external broker)", "nats_url", cfg.NATSURL)
		}
		go func() {
			if err := supervise.StartSupervisor(ctx, supervise.SupervisorConfig{
				Logg: slog.Default(),
				Skip: skip,
			}); err != nil {
				slog.Error("ctrl: supervisor stopped", "err", err)
			}
		}()
		time.Sleep(cfg.SuperviseStartupDelay)
	}

	app := svcutil.Bootstrap("ctrl", "0.3.0", "Platform control: lifecycle + status + optional supervisor")
	defer app.Shutdown()
	if cancelSupervise != nil {
		defer cancelSupervise()
	}

	edgeStore, err := edgecfg.Open(app.Ctx, app.NC, cfg.ConfigBucket)
	if err != nil {
		slog.Error("ctrl: init edgecfg store failed", "err", err)
	}

	apis.Register(app.Svc, &apis.Deps{
		NC:        app.NC,
		Cfg:       cfg,
		Logger:    loggclient.MustNew(app.NC),
		EdgeStore: edgeStore,
	})

	slog.Info("ctrl service started",
		"supervise", cfg.Up,
		"cfg_dir", mini.ConfigDir(),
		"config_bucket", cfg.ConfigBucket,
	)

	app.WireLifecycle(nil)
	app.Wait()
}
