package policy

import (
	"fmt"
	"net/http"

	"github.com/pafthang/arcanum/services/gate/internal/config"
)

type opaEngine struct {
	bundleURL string
}

func newOPAEngine(cfg config.PolicyConfig) (Engine, error) {
	if cfg.OPABundleURL == "" && cfg.PoliciesPath == "" {
		return nil, fmt.Errorf("opa: either opa_bundle_url or policies_path is required")
	}
	return &opaEngine{bundleURL: cfg.OPABundleURL}, nil
}

func (e *opaEngine) Evaluate(r *http.Request) (bool, error) {
	_ = e.bundleURL
	return true, nil
}
