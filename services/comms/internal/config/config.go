package config

import "github.com/pafthang/arcanum/pkg/svcutil"

// Config is the comms service configuration.
type Config struct {
	ListDefaultPerPage int
	ListMaxPerPage     int
}

// Defaults returns the default configuration.
func Defaults() Config {
	return Config{
		ListDefaultPerPage: 50,
		ListMaxPerPage:     200,
	}
}

// FromEnv loads configuration from environment.
func FromEnv() Config {
	c := Defaults()
	c.ListDefaultPerPage = svcutil.EnvInt("COMMS_LIST_PER_PAGE", c.ListDefaultPerPage)
	c.ListMaxPerPage = svcutil.EnvInt("COMMS_LIST_MAX", c.ListMaxPerPage)
	if c.ListMaxPerPage < c.ListDefaultPerPage {
		c.ListMaxPerPage = c.ListDefaultPerPage
	}
	return c
}
