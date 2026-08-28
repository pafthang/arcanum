package svcutil

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/nats-io/nats.go"
)

func TestModuleFuncAndSkipOrder(t *testing.T) {
	var seq []string
	m := func(name string) Module {
		return ModuleFunc(name, func(ctx ModuleContext) (func(), error) {
			if ctx.Process != "testproc" {
				t.Fatalf("process=%q", ctx.Process)
			}
			if ctx.Name != name {
				t.Fatalf("name=%q want %q", ctx.Name, name)
			}
			seq = append(seq, "start:"+name)
			return func() { seq = append(seq, "stop:"+name) }, nil
		})
	}

	h := NewHost("testproc").Modules(m("a"), m("b"), m("c")).Skip("b")
	var started []startedMod
	for _, mod := range h.modules {
		name := mod.Name()
		if h.skip[name] {
			continue
		}
		st, err := mod.Start(ModuleContext{Process: h.name, Name: name})
		if err != nil {
			t.Fatal(err)
		}
		started = append(started, startedMod{name: name, stop: st})
	}
	if len(seq) != 2 || seq[0] != "start:a" || seq[1] != "start:c" {
		t.Fatalf("start seq=%v", seq)
	}
	for i := len(started) - 1; i >= 0; i-- {
		started[i].stop()
	}
	if len(seq) != 4 || seq[2] != "stop:c" || seq[3] != "stop:a" {
		t.Fatalf("full seq=%v (want LIFO stop c then a)", seq)
	}
}

func TestModuleFuncError(t *testing.T) {
	m := ModuleFunc("x", func(ctx ModuleContext) (func(), error) {
		return nil, errors.New("boom")
	})
	_, err := m.Start(ModuleContext{Name: "x"})
	if err == nil || err.Error() != "boom" {
		t.Fatalf("err=%v", err)
	}
}

func TestModuleOf(t *testing.T) {
	var stopped atomic.Bool
	m := ModuleOf("legacy", func(nc *nats.Conn) func() {
		if nc != nil {
			t.Fatal("expected nil nc in unit test")
		}
		return func() { stopped.Store(true) }
	})
	if m.Name() != "legacy" {
		t.Fatal(m.Name())
	}
	st, err := m.Start(ModuleContext{Name: "legacy", NC: nil})
	if err != nil {
		t.Fatal(err)
	}
	st()
	if !stopped.Load() {
		t.Fatal("stop not called")
	}
}

func TestModuleContextDataDir(t *testing.T) {
	c := ModuleContext{Process: "exec", Name: "cron"}
	if got := c.DataDir(); got != DataDir("cron") {
		t.Fatalf("got %q", got)
	}
}
