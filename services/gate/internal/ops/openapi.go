package ops

import (
	"encoding/json"
	"net/http"
)

var openAPISpec = map[string]any{
	"openapi": "3.0.3",
	"info": map[string]any{
		"title":       "Optima Gate",
		"description": "API Gate for Optima platform",
		"version":     "1.0.0",
	},
	"paths": map[string]any{
		"/healthz": map[string]any{
			"get": map[string]any{
				"summary": "Health check",
				"responses": map[string]any{
					"200": map[string]any{"description": "OK"},
				},
			},
		},
		"/readyz": map[string]any{
			"get": map[string]any{
				"summary": "Readiness check",
				"responses": map[string]any{
					"200": map[string]any{"description": "Ready"},
				},
			},
		},
		"/metrics": map[string]any{
			"get": map[string]any{
				"summary": "Prometheus-style metrics",
				"responses": map[string]any{
					"200": map[string]any{"description": "Metrics in text format"},
				},
			},
		},
	},
	"components": map[string]any{
		"securitySchemes": map[string]any{
			"bearerAuth": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
			},
			"apiKeyAuth": map[string]any{
				"type": "apiKey",
				"in":   "header",
				"name": "X-API-Key",
			},
		},
	},
}

// OpenAPIHandler serves the OpenAPI document.
func OpenAPIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(openAPISpec)
}
