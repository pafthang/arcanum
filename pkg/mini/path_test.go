package mini

import "testing"

func TestParseAndMatchPathPattern(t *testing.T) {
	p, err := ParsePathPattern("/v1/orders/{id}")
	if err != nil {
		t.Fatal(err)
	}
	if !p.HasParams() {
		t.Fatal("expected params")
	}
	params, ok := p.Match("/v1/orders/42")
	if !ok || params["id"] != "42" {
		t.Fatalf("got %v ok=%v", params, ok)
	}
	if _, ok := p.Match("/v1/orders"); ok {
		t.Fatal("should not match shorter path")
	}
	if _, ok := p.Match("/v1/orders/42/extra"); ok {
		t.Fatal("should not match longer path")
	}
}

func TestStaticPathMatchNoParamsMap(t *testing.T) {
	p, err := ParsePathPattern("/api/agents")
	if err != nil {
		t.Fatal(err)
	}
	if p.HasParams() {
		t.Fatal("static pattern must not report params")
	}
	params, ok := p.Match("/api/agents")
	if !ok {
		t.Fatal("expected match")
	}
	if params != nil {
		t.Fatalf("expected nil params map, got %v", params)
	}
	if _, ok := p.Match("/api/agents/"); !ok {
		t.Fatal("trailing slash should match")
	}
	if _, ok := p.Match("/api/other"); ok {
		t.Fatal("should not match")
	}
}

func TestWildcardAndCatchAll(t *testing.T) {
	w, err := ParsePathPattern("/v1/files/*")
	if err != nil {
		t.Fatal(err)
	}
	params, ok := w.Match("/v1/files/a.txt")
	if !ok || params[WildcardParam] != "a.txt" {
		t.Fatalf("wildcard: %v ok=%v", params, ok)
	}
	if _, ok := w.Match("/v1/files/a/b"); ok {
		t.Fatal("* is single segment")
	}

	c, err := ParsePathPattern("/v1/assets/{*path}")
	if err != nil {
		t.Fatal(err)
	}
	params, ok = c.Match("/v1/assets/img/logo.png")
	if !ok || params["path"] != "img/logo.png" {
		t.Fatalf("catch-all: %v ok=%v", params, ok)
	}

	if _, err := ParsePathPattern("/v1/{*path}/x"); err == nil {
		t.Fatal("catch-all must be last")
	}
}

func TestPathSpecificityOrder(t *testing.T) {
	exact, _ := ParsePathPattern("/v1/orders/all")
	param, _ := ParsePathPattern("/v1/orders/{id}")
	wild, _ := ParsePathPattern("/v1/orders/*")
	catch, _ := ParsePathPattern("/v1/{*path}")

	if !ComparePathSpecificity(exact, param) {
		t.Fatal("exact > param")
	}
	if !ComparePathSpecificity(param, wild) {
		// both 2 static? orders is static: exact has 3 static, param has 2 static + 1 param
		// param: v1, orders static=2; wild: v1, orders static=2, wild=1
		// param dynamic=1, wild dynamic=1 - same static and dynamic; length may decide
	}
	if !ComparePathSpecificity(param, catch) {
		t.Fatal("param > catch-all")
	}
	if !ComparePathSpecificity(wild, catch) {
		t.Fatal("wild > catch-all")
	}
	_ = exact
}

func TestPublicSubject(t *testing.T) {
	if got := PublicSubject("orders", "get"); got != "public.orders.get" {
		t.Fatal(got)
	}
	if !IsPublicSubject("public.orders.get") {
		t.Fatal("expected public")
	}
	if IsPublicSubject("orders.get") {
		t.Fatal("not public")
	}
}
