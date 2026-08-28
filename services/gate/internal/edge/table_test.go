package edge

import (
	"testing"

	"github.com/pafthang/arcanum/pkg/mini"
)

func TestTableHTTPAndWS(t *testing.T) {
	tbl := NewTable()
	tbl.Replace([]mini.PublicRoute{
		{
			Kind:    mini.TransportHTTP,
			Method:  "GET",
			Path:    "/api/spaces/{spaceId}/office/servers",
			Subject: "public.office.servers.list",
		},
		{
			Kind:      mini.TransportWS,
			Method:    mini.WSMethod,
			Path:      "/api/spaces/{spaceId}/office/ssh/{connectionId}/ws",
			Subscribe: "public.office.ssh.{connectionId}.out",
			Publish:   "public.office.ssh.{connectionId}.in",
			Subject:   "public.office.ssh.{connectionId}.out",
		},
	})
	if tbl.Len() != 2 {
		t.Fatalf("len=%d", tbl.Len())
	}
	if tbl.HTTPLen() != 1 {
		t.Fatalf("httpLen=%d", tbl.HTTPLen())
	}

	r, params, ok := tbl.Match("GET", "/api/spaces/t1/office/servers")
	if !ok || r == nil || params["spaceId"] != "t1" {
		t.Fatalf("http match failed: ok=%v params=%v", ok, params)
	}

	wr, wp, ok := tbl.MatchWS("/api/spaces/t1/office/ssh/cabc/ws")
	if !ok || wr == nil {
		t.Fatal("ws match failed")
	}
	if wp["spaceId"] != "t1" || wp["connectionId"] != "cabc" {
		t.Fatalf("ws params=%v", wp)
	}
	sub := mini.ExpandSubject(wr.Subscribe, wp)
	if sub != "public.office.ssh.cabc.out" {
		t.Fatalf("subscribe expand=%q", sub)
	}
}
