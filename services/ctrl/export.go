package ctrl

import (
	"github.com/pafthang/arcanum/services/ctrl/internal/config"
	"github.com/pafthang/arcanum/services/ctrl/models"
)

// Re-export key types for consumers of the ctrl package.

type (
	Config = config.Config
)

// FromEnv loads Config from CTRL_* / CONFIG_* environment variables.
func FromEnv() Config {
	return config.FromEnv()
}

// ServiceStatus is re-exported for lifecycle API consumers.
type ServiceStatus = models.ServiceStatus
