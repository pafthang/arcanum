package apis

import (
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/pkg/subjects"
	"github.com/pafthang/arcanum/services/comms/internal/config"
	"github.com/pafthang/arcanum/services/comms/internal/store"
	spaceclient "github.com/pafthang/arcanum/services/space/client"
)

// Deps holds runtime dependencies.
type Deps struct {
	Store  *store.Store
	NC     *nats.Conn
	Space  *spaceclient.Client
	Config config.Config
}

// Register attaches public HTTP, WS catalog, and internal NATS endpoints.
func Register(svc mini.Service, d *Deps) {
	if d == nil {
		panic("comms/apis.Register: nil Deps")
	}
	registerChannels(svc, d)
	registerMessages(svc, d)
	registerWS(svc)
	registerInternal(d)
}

func registerWS(svc mini.Service) {
	must(svc.AddPublicWS("channel_ws", mini.WSConfig{
		Path:      "/api/spaces/{spaceId}/channels/{channelId}/ws",
		Subscribe: subjects.EventCommsChannelWSPattern,
	}, mini.WithPublicSubject("comms", "channel.ws"), mini.WithPublicAuth(mini.AuthRequired)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
