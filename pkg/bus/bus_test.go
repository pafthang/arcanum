package bus

import (
	"testing"
	"time"
)

func TestDlqSubjectFor(t *testing.T) {
	if got := dlqSubjectFor("commands.exec.run"); got != "commands.dlq.exec.run" {
		t.Fatalf("got %s", got)
	}
}

func TestDefaultCommandOpts(t *testing.T) {
	o := defaultCommandOpts("commands.exec.run")
	if o.MaxDeliver < 1 {
		t.Fatal("max deliver")
	}
	if o.DLQSubject != "commands.dlq.exec.run" {
		t.Fatalf("dlq %s", o.DLQSubject)
	}
	if o.AckWait < time.Second {
		t.Fatal("ack wait")
	}
}
