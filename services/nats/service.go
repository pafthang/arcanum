// Package nats runs an embedded NATS Server as a platform process.
//
// Layout (edge style):
//
//	service.go   — Run (process entry)
//	export.go    — Config + FromEnv
//	cmd/nats/    — thin main → nats.Run()
//	internal/
//	  config/      NATS_* env + server options
//
// Unlike domain services this process does not connect as a NATS client or
// register mini endpoints — it *is* the message bus that every other service
// connects to (including JetStream for bus + control-plane KV).
//
// Environment (see config.FromEnv / cfgs/nats.json):
//
//	NATS_HOST, NATS_PORT, NATS_HTTP_PORT
//	NATS_SERVER_NAME, NATS_JETSTREAM, NATS_STORE_DIR
//	NATS_CONFIG, NATS_DEBUG, NATS_TRACE
package nats

import (
	"log"
	"log/slog"
	"path/filepath"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/pafthang/arcanum/pkg/logx"
	"github.com/pafthang/arcanum/pkg/svcutil"
	natsconfig "github.com/pafthang/arcanum/services/nats/internal/config"
)

// Run starts the embedded NATS server and blocks until SIGINT/SIGTERM.
func Run() {
	// Load cfgs/nats.json (+ common) before reading env.
	svcutil.LoadConfig("nats")

	cfg := natsconfig.FromEnv()

	_ = logx.Setup(logx.Config{
		Term: true,
		File: filepath.Join(svcutil.DataDir("nats"), "service.log"),
		NATS: nil, // NATS doesn't log to NATS to avoid loops
	})

	opts, err := cfg.ServerOptions()
	if err != nil {
		log.Fatalf("nats options: %v", err)
	}

	ns, err := natsserver.NewServer(opts)
	if err != nil {
		log.Fatalf("nats new server: %v", err)
	}
	ns.ConfigureLogger()

	go ns.Start()
	if !ns.ReadyForConnections(10 * time.Second) {
		log.Fatal("nats server not ready for connections")
	}

	slog.Info("nats service started",
		"url", ns.ClientURL(),
		"host", opts.Host,
		"port", opts.Port,
		"http_port", opts.HTTPPort,
		"jetstream", opts.JetStream,
		"store_dir", opts.StoreDir,
		"server_name", opts.ServerName,
		"config", opts.ConfigFile,
	)

	svcutil.WaitSignal()
	slog.Info("nats service shutting down")
	ns.Shutdown()
	ns.WaitForShutdown()
}
