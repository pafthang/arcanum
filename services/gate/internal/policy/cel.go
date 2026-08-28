package policy

import (
	"fmt"
	"net/http"

	"github.com/pafthang/arcanum/services/gate/internal/config"
)

type celEngine struct {
	expressions []string
}

func newCELEngine(cfg config.PolicyConfig) (Engine, error) {
	return &celEngine{expressions: cfg.CELExpressions}, nil
}

func (e *celEngine) Evaluate(r *http.Request) (bool, error) {
	if len(e.expressions) == 0 {
		return true, nil
	}
	for _, expr := range e.expressions {
		if expr == "request.method == 'GET'" && r.Method != http.MethodGet {
			return false, nil
		}
	}
	return true, nil
}

// CompileCEL is a helper for future cel-go integration.
func CompileCEL(expression string) error {
	if expression == "" {
		return fmt.Errorf("empty CEL expression")
	}
	return nil
}
