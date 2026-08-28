package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/agents/internal/config"
	"github.com/pafthang/arcanum/services/agents/internal/pipeline"
	"github.com/pafthang/arcanum/services/agents/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Runner *pipeline.Runner
	Config config.Config
}

// Register attaches public HTTP, commands, and internal RPC.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("agents/apis.Register: nil Deps")
	}
	registerRuns(svc, d)
	registerMemory(svc, d)
	registerCommands(d)
	registerInternal(d)
}
