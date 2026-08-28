package config

import (
	"os"
	"strings"

	"github.com/pafthang/arcanum/pkg/svcutil"
	"github.com/pafthang/arcanum/services/agents/internal/providers"
)

// Config is agents process configuration.
type Config struct {
	ListDefaultPerPage int
	ListMaxPerPage     int
	LLMBaseURL         string
	LLMAPIKey          string
	LLMModel           string
	LLMMaxSteps        int
}

// Defaults returns built-in values.
func Defaults() Config {
	return Config{
		ListDefaultPerPage: 50,
		ListMaxPerPage:     200,
		LLMBaseURL:         "https://api.openai.com/v1",
		LLMModel:           "gpt-4o-mini",
		LLMMaxSteps:        8,
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
	c.LLMBaseURL = env("AGENTS_LLM_BASE_URL", c.LLMBaseURL)
	c.LLMAPIKey = env("AGENTS_LLM_API_KEY", "")
	c.LLMModel = env("AGENTS_LLM_MODEL", c.LLMModel)
	c.LLMMaxSteps = svcutil.EnvInt("AGENTS_LLM_MAX_STEPS", c.LLMMaxSteps)
	if c.LLMMaxSteps < 1 {
		c.LLMMaxSteps = 1
	}
	return c
}

func env(key, def string) string {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return def
	}
	return s
}

// Provider returns an OpenAI-compatible client when a key is set.
func (c Config) Provider() providers.Provider {
	if strings.TrimSpace(c.LLMAPIKey) == "" {
		return nil
	}
	return providers.NewOpenAI(c.LLMBaseURL, c.LLMAPIKey, c.LLMModel)
}
