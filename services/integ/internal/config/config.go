package config

// Config is the integ service configuration.
type Config struct{}

// FromEnv loads config from environment.
func FromEnv() Config {
	return Config{}
}
