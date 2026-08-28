// Package apis registers ctrl platform lifecycle / inventory endpoints.
package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/ctrl/internal/config"
	"github.com/pafthang/arcanum/services/ctrl/internal/edgecfg"
	loggclient "github.com/pafthang/arcanum/services/logg/client"
	loggmodels "github.com/pafthang/arcanum/services/logg/models"
)

// Deps are runtime dependencies for handlers.
type Deps struct {
	NC        *nats.Conn
	Cfg       config.Config
	Logger    *loggclient.Client
	EdgeStore *edgecfg.Store
}

func (d *Deps) recordActivity(a loggmodels.Activity) {
	if d == nil || d.Logger == nil {
		return
	}
	d.Logger.AppendActivityAsync(&a)
}

// Register attaches public platform control endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("ctrl/apis.Register: nil Deps")
	}
	registerLifecycle(svc, d)
	registerInventory(svc, d)
	registerEdgecfg(svc, d)
}
