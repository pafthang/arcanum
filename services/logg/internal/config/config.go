package config

import (
	"github.com/pafthang/arcanum/pkg/svcutil"
)

// Config represents the logg service configuration.
type Config struct {
	ListDefaultPerPage int
	ListMaxPerPage     int
}

// Defaults returns the default configuration.
func Defaults() Config {
	return Config{
		ListDefaultPerPage: 100,
		ListMaxPerPage:     500,
	}
}

// FromEnv loads the configuration from environment variables.
func FromEnv() Config {
	c := Defaults()
	c.ListDefaultPerPage = svcutil.EnvInt("LOGG_LIST_PER_PAGE", c.ListDefaultPerPage)
	c.ListMaxPerPage = svcutil.EnvInt("LOGG_LIST_MAX", c.ListMaxPerPage)

	if c.ListMaxPerPage < c.ListDefaultPerPage {
		c.ListMaxPerPage = c.ListDefaultPerPage
	}

	return c
}
