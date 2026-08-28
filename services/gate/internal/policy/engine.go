package policy

import (
	"fmt"
	"net/http"

	"github.com/pafthang/arcanum/services/gate/internal/config"
)

// Engine is the policy engine interface.
type Engine interface {
	Evaluate(r *http.Request) (allowed bool, err error)
}

// NewEngine builds an engine from config.
func NewEngine(cfg config.PolicyConfig) (Engine, error) {
	switch cfg.Engine {
	case "", "none":
		return &noopEngine{}, nil
	case "cel":
		return newCELEngine(cfg)
	case "opa":
		return newOPAEngine(cfg)
	default:
		return nil, fmt.Errorf("unknown policy engine: %s", cfg.Engine)
	}
}

// Middleware wraps a handler with a policy check.
func Middleware(eng Engine) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, err := eng.Evaluate(r)
			if err != nil {
				http.Error(w, "policy evaluation error", http.StatusInternalServerError)
				return
			}
			if !allowed {
				http.Error(w, "forbidden by policy", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type noopEngine struct{}

func (n *noopEngine) Evaluate(r *http.Request) (bool, error) {
	return true, nil
}
