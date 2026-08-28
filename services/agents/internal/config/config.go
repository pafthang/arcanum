package config

import "github.com/pafthang/arcanum/pkg/svcutil"

// Config is agents process configuration.
type Config struct {
	ListDefaultPerPage int
	ListMaxPerPage     int
}

// Defaults returns built-in values.
func Defaults() Config {
	return Config{
		ListDefaultPerPage: 50,
		ListMaxPerPage:     200,
	}
}

// FromEnv loads configuration from environment.
func FromEnv() Config {
	c := Defaults()
	c.ListDefaultPerPage = svcutil.EnvInt("AGENTS_LIST_PER_PAGE", c.ListDefaultPerPage)
	c.ListMaxPerPage = svcutil.EnvInt("AGENTS_LIST_MAX", c.ListMaxPerPage)
	if c.ListMaxPerPage < c.ListDefaultPerPage {
		c.ListMaxPerPage = c.ListDefaultPerPage
	}
	return c
}
