package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnvAndServerOptionsDefaults(t *testing.T) {
	t.Setenv("NATS_CONFIG", "")
	t.Setenv("NATS_HOST", "127.0.0.1")
	t.Setenv("NATS_PORT", "14223")
	t.Setenv("NATS_HTTP_PORT", "0")
	t.Setenv("NATS_JETSTREAM", "true")
	t.Setenv("NATS_SERVER_NAME", "test-nats")
	dir := t.TempDir()
	t.Setenv("NATS_STORE_DIR", dir)
	t.Setenv("NATS_DEBUG", "")
	t.Setenv("NATS_TRACE", "")

	c := FromEnv()
	opts, err := c.ServerOptions()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Host != "127.0.0.1" || opts.Port != 14223 {
		t.Fatalf("bind = %s:%d", opts.Host, opts.Port)
	}
	if !opts.JetStream {
		t.Fatal("expected JetStream")
	}
	if opts.StoreDir != dir {
		t.Fatalf("store = %q", opts.StoreDir)
	}
	if !opts.NoSigs {
		t.Fatal("expected NoSigs for embedded mode")
	}
	if opts.ServerName != "test-nats" {
		t.Fatalf("name = %q", opts.ServerName)
	}
}

func TestFromEnvDisableJetStream(t *testing.T) {
	t.Setenv("NATS_CONFIG", "")
	t.Setenv("NATS_PORT", "14224")
	t.Setenv("NATS_HTTP_PORT", "0")
	t.Setenv("NATS_JETSTREAM", "false")
	t.Setenv("NATS_STORE_DIR", t.TempDir())

	c := FromEnv()
	opts, err := c.ServerOptions()
	if err != nil {
		t.Fatal(err)
	}
	if opts.JetStream {
		t.Fatal("expected JetStream off")
	}
}

func TestServerOptionsFromConfigFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nats.conf")
	content := "port: 14225\nhttp_port: 0\njetstream: enabled\nstore_dir: " + dir + "\n"
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("NATS_CONFIG", cfgPath)
	t.Setenv("NATS_HOST", "127.0.0.1")
	// clear overrides that would fight the conf
	t.Setenv("NATS_PORT", "")
	t.Setenv("NATS_HTTP_PORT", "")
	t.Setenv("NATS_JETSTREAM", "")
	t.Setenv("NATS_STORE_DIR", "")
	t.Setenv("NATS_SERVER_NAME", "")

	c := FromEnv()
	opts, err := c.ServerOptions()
	if err != nil {
		t.Fatal(err)
	}
	if opts.Port != 14225 {
		t.Fatalf("port = %d", opts.Port)
	}
	if !opts.JetStream {
		t.Fatal("expected JetStream from conf")
	}
	if !opts.NoSigs {
		t.Fatal("expected NoSigs")
	}
	if opts.ConfigFile != cfgPath {
		t.Fatalf("config = %q", opts.ConfigFile)
	}
}
