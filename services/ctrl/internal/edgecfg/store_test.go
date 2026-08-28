package edgecfg_test

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/pafthang/arcanum/pkg/mini"
	"github.com/pafthang/arcanum/services/ctrl/internal/edgecfg"
)

func startJS(t *testing.T) *nats.Conn {
	t.Helper()
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
	}
	s, err := server.NewServer(opts)
	if err != nil {
		t.Fatal(err)
	}
	go s.Start()
	if !s.ReadyForConnections(10 * time.Second) {
		t.Fatal("server not ready")
	}
	t.Cleanup(s.Shutdown)

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(nc.Close)
	return nc
}

func TestStorePutListDelete(t *testing.T) {
	nc := startJS(t)
	ctx := context.Background()
	store, err := edgecfg.Open(ctx, nc, "TEST_MINI_CP")
	if err != nil {
		t.Fatal(err)
	}

	doc, err := store.PutRoute(ctx, edgecfg.RouteDoc{
		Method:  "GET",
		Path:    "/v1/orders",
		Subject: "public.orders.list",
		Service: "orders",
	})
	if err != nil {
		t.Fatal(err)
	}
	if doc.ID == "" {
		t.Fatal("id empty")
	}

	list, err := store.ListRoutes(ctx, false)
	if err != nil || len(list) != 1 {
		t.Fatalf("list: %v %v", list, err)
	}
	pr, err := list[0].ToPublicRoute()
	if err != nil || pr.Subject != "public.orders.list" {
		t.Fatal(pr, err)
	}

	_, err = store.PutWSACL(ctx, edgecfg.WSACLDoc{
		Path:        "/v1/rooms/{room}/ws",
		RequireAuth: true,
		ClaimEquals: map[string]string{"tenant": "{room}"},
	})
	if err != nil {
		t.Fatal(err)
	}
	acls, err := store.ListWSACL(ctx, false)
	if err != nil || len(acls) != 1 {
		t.Fatal(acls, err)
	}

	if err := store.DeleteRoute(ctx, doc.ID); err != nil {
		t.Fatal(err)
	}
	list, err = store.ListRoutes(ctx, false)
	if err != nil || len(list) != 0 {
		t.Fatalf("after delete: %v %v", list, err)
	}
}

func TestWatcherReload(t *testing.T) {
	nc := startJS(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := edgecfg.Open(ctx, nc, "TEST_MINI_CP_W")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.PutRoute(ctx, edgecfg.RouteDoc{
		Kind: mini.TransportWS, Path: "/v1/ws",
		Subscribe: "public.events",
	})
	if err != nil {
		t.Fatal(err)
	}

	w := edgecfg.NewWatcher(store, nil)
	if err := w.Reload(ctx); err != nil {
		t.Fatal(err)
	}
	snap := w.Snapshot()
	if len(snap.Routes) != 1 {
		t.Fatalf("routes: %+v", snap.Routes)
	}
	if snap.Routes[0].Method != mini.WSMethod {
		t.Fatal(snap.Routes[0])
	}
}

func TestGatewayControlKVTruth(t *testing.T) {
	// integration-style: config store + gate refresh
	nc := startJS(t)
	ctx := context.Background()
	store, err := edgecfg.Open(ctx, nc, "TEST_GW_CP")
	if err != nil {
		t.Fatal(err)
	}
	_, err = store.PutRoute(ctx, edgecfg.RouteDoc{
		Method: "GET", Path: "/v1/cp", Subject: "public.cp",
	})
	if err != nil {
		t.Fatal(err)
	}

	// import gate package
	// done in gate test file to avoid cycle - skip here
	_ = store
}
