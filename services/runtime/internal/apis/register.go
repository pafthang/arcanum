package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/runtime/internal/config"
	"github.com/pafthang/arcanum/services/runtime/internal/docker"
	"github.com/pafthang/arcanum/services/runtime/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Config config.Config
	Docker docker.Engine
}

// Register attaches public HTTP endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("runtime/apis.Register: nil Deps")
	}
	registerPublic(svc, d)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
