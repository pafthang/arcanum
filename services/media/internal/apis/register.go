package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/media/internal/config"
	"github.com/pafthang/arcanum/services/media/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Config config.Config
}

// Register attaches public HTTP and internal NATS endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("media/apis.Register: nil Deps")
	}
	registerPublic(svc, d)
	registerDelete(svc, d)
	registerURL(svc, d)
	registerInternal(d)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
