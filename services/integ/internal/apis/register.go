package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/integ/internal/config"
	"github.com/pafthang/arcanum/services/integ/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Config config.Config
}

// Register attaches public HTTP, inbound hooks, internal RPC, and work fan-out.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("integ/apis.Register: nil Deps")
	}
	registerConnectors(svc, d)
	registerWebhooks(svc, d)
	registerHooks(svc, d)
	registerInternal(d)
	registerFanout(d)
}
