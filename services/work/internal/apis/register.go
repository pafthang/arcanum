package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
	mediaclient "github.com/pafthang/arcanum/services/media/client"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
	"github.com/pafthang/arcanum/services/work/internal/config"
	"github.com/pafthang/arcanum/services/work/internal/store"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Logg   *loggclient.Client
	Media  *mediaclient.Client
	Config config.Config
}

// Register attaches public HTTP and internal NATS endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("work/apis.Register: nil Deps")
	}
	registerIssues(svc, d)
	registerComments(svc, d)
	registerLabels(svc, d)
	registerInternal(d)
}
