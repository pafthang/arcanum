package config

import (
	"os"
	"strings"
)

// Config is runtime process configuration.
type Config struct {
	DockerHost string
}

// FromEnv loads configuration from environment.
func FromEnv() Config {
	return Config{DockerHost: strings.TrimSpace(os.Getenv("RUNTIME_DOCKER_HOST"))}
}

// HasDocker reports whether a docker endpoint is configured.
func (c Config) HasDocker() bool {
	return c.DockerHost != ""
}
