package mini

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	// DefaultConfigDir is the directory of per-service config files.
	DefaultConfigDir = "cfgs"

	// EnvConfigDir overrides the config directory (default: cfgs).
	EnvConfigDir = "CFG_DIR"

	// EnvConfigFile forces a single config file for the current process
	// (skips common + <name>.json lookup by name).
	EnvConfigFile = "CFG_FILE"

	// EnvConfigSkip disables file config loading when set to a truthy value.
	EnvConfigSkip = "CFG_SKIP"
)

// ServiceFileConfig is the on-disk shape of cfgs/<name>.json (and comm.json).
//
// Env values are applied to the process only when the key is not already set
// (process / shell env wins). Used by domain services via svcutil and by
// ctrl when supervising a stack.
type ServiceFileConfig struct {
	// Name is optional; defaults to the file stem when empty.
	Name string `json:"name,omitempty"`

	// Description is free-form documentation for operators.
	Description string `json:"description,omitempty"`

	// Enabled controls whether ctrl -up should start this service.
	// nil / omitted means true.
	Enabled *bool `json:"enabled,omitempty"`

	// Order is start order for supervised stacks (lower first). Default 100.
	Order int `json:"order,omitempty"`

	// Command overrides the default start command for ctrl supervise.
	// Empty → ["go", "run", "./cmd/<name>"].
	Command []string `json:"command,omitempty"`

	// Env is applied with ApplyServiceConfig (only unset keys).
	Env map[string]string `json:"env,omitempty"`
}

// IsEnabled reports whether the service should be started by a supervisor.
func (c ServiceFileConfig) IsEnabled() bool {
	if c.Enabled == nil {
		return true
	}
	return *c.Enabled
}

// StartOrder returns the effective start order (lower first).
// Explicit Order wins; nats defaults to 0; everything else defaults to 100.
func (c ServiceFileConfig) StartOrder() int {
	return effectiveOrder(c)
}

// ConfigDir returns CFG_DIR or DefaultConfigDir.
func ConfigDir() string {
	if v := strings.TrimSpace(os.Getenv(EnvConfigDir)); v != "" {
		return v
	}
	return DefaultConfigDir
}

// ConfigSkip reports whether file config loading is disabled.
func ConfigSkip() bool {
	return envTruthy(os.Getenv(EnvConfigSkip))
}

// LoadServiceFile loads and merges comm.json (or _comm.json) with
// cfgs/<name>.json. Missing files yield an empty config and nil error.
// Env maps are shallow-merged (service overrides common).
func LoadServiceFile(name string) (ServiceFileConfig, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return ServiceFileConfig{}, fmt.Errorf("%w: empty service name", ErrConfigValidation)
	}
	if ConfigSkip() {
		return ServiceFileConfig{Name: name}, nil
	}

	dir := ConfigDir()
	var out ServiceFileConfig

	for _, common := range []string{"comm.json", "_comm.json"} {
		path := filepath.Join(dir, common)
		c, err := readServiceFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return ServiceFileConfig{}, err
		}
		out = mergeServiceFile(out, c)
		break // only one common file
	}

	// Explicit single file wins over name-based path.
	if f := strings.TrimSpace(os.Getenv(EnvConfigFile)); f != "" {
		c, err := readServiceFile(f)
		if err != nil {
			return ServiceFileConfig{}, err
		}
		out = mergeServiceFile(out, c)
	} else {
		path := filepath.Join(dir, name+".json")
		c, err := readServiceFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return ServiceFileConfig{}, err
			}
		} else {
			out = mergeServiceFile(out, c)
		}
	}

	if out.Name == "" {
		out.Name = name
	}
	return out, nil
}

// ApplyServiceConfig loads the service file config and sets process env for
// every key that is not already set. Safe to call multiple times.
// Missing config files are a no-op.
func ApplyServiceConfig(name string) (ServiceFileConfig, error) {
	cfg, err := LoadServiceFile(name)
	if err != nil {
		return cfg, err
	}
	applyEnv(cfg.Env)
	return cfg, nil
}

// ListServiceFiles returns service names for *.json entries in ConfigDir,
// excluding comm.json / _comm.json / stack.json. Sorted by StartOrder then name.
func ListServiceFiles() ([]ServiceFileConfig, error) {
	dir := ConfigDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}

	var out []ServiceFileConfig
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		base := e.Name()
		if !strings.HasSuffix(base, ".json") {
			continue
		}
		stem := strings.TrimSuffix(base, ".json")
		switch stem {
		case "common", "_common", "comm", "_comm", "stack":
			continue
		}
		cfg, err := LoadServiceFile(stem)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		if cfg.Name == "" {
			cfg.Name = stem
		}
		out = append(out, cfg)
	}

	sort.SliceStable(out, func(i, j int) bool {
		oi, oj := effectiveOrder(out[i]), effectiveOrder(out[j])
		if oi != oj {
			return oi < oj
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// effectiveOrder: explicit order wins; nats defaults to 0; others to 100.
func effectiveOrder(c ServiceFileConfig) int {
	// JSON omitempty means 0 is ambiguous. Convention:
	// - if order key was set to 0 in file, Order is 0
	// - we treat Order==0 as "default" unless name is nats
	if c.Order != 0 {
		return c.Order
	}
	if c.Name == "nats" {
		return 0
	}
	return 100
}

func readServiceFile(path string) (ServiceFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ServiceFileConfig{}, err
	}
	data = stripJSONComments(data)
	var c ServiceFileConfig
	if err := json.Unmarshal(data, &c); err != nil {
		return ServiceFileConfig{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

// SaveServiceFile writes the ServiceFileConfig back to cfgs/<name>.json.
func SaveServiceFile(name string, c ServiceFileConfig) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("empty service name")
	}
	dir := ConfigDir()
	path := filepath.Join(dir, name+".json")

	// Read existing to preserve any comments or fields we don't map if we want?
	// Actually, just overwrite. JSON Marshal is fine.
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	// append a newline
	data = append(data, '\n')

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func mergeServiceFile(base, over ServiceFileConfig) ServiceFileConfig {
	if over.Name != "" {
		base.Name = over.Name
	}
	if over.Description != "" {
		base.Description = over.Description
	}
	if over.Enabled != nil {
		base.Enabled = over.Enabled
	}
	if over.Order != 0 {
		base.Order = over.Order
	}
	if len(over.Command) > 0 {
		base.Command = append([]string(nil), over.Command...)
	}
	if len(over.Env) > 0 {
		if base.Env == nil {
			base.Env = map[string]string{}
		}
		for k, v := range over.Env {
			base.Env[k] = v
		}
	}
	return base
}

func applyEnv(env map[string]string) {
	for k, v := range env {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if os.Getenv(k) != "" {
			continue // process env wins
		}
		_ = os.Setenv(k, v)
	}
}

func envTruthy(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

// stripJSONComments removes // line comments for operator-friendly cfgs.
// Strings are left intact (no comment stripping inside quotes).
func stripJSONComments(data []byte) []byte {
	var b strings.Builder
	b.Grow(len(data))
	inStr := false
	esc := false
	for i := 0; i < len(data); i++ {
		c := data[i]
		if inStr {
			b.WriteByte(c)
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		if c == '"' {
			inStr = true
			b.WriteByte(c)
			continue
		}
		if c == '/' && i+1 < len(data) && data[i+1] == '/' {
			// skip to end of line
			for i < len(data) && data[i] != '\n' {
				i++
			}
			if i < len(data) {
				b.WriteByte('\n')
			}
			continue
		}
		b.WriteByte(c)
	}
	return []byte(b.String())
}

// ChildEnv builds an environment block for a supervised child:
// parent environ, then file env for keys not already set.
func ChildEnv(cfg ServiceFileConfig) []string {
	env := os.Environ()
	have := map[string]struct{}{}
	for _, e := range env {
		if k, _, ok := strings.Cut(e, "="); ok {
			have[k] = struct{}{}
		}
	}
	for k, v := range cfg.Env {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		if _, ok := have[k]; ok {
			continue
		}
		env = append(env, k+"="+v)
		have[k] = struct{}{}
	}
	// Ensure CFG_DIR is visible to children so they load the same tree.
	if _, ok := have[EnvConfigDir]; !ok {
		env = append(env, EnvConfigDir+"="+ConfigDir())
	}
	return env
}

// DefaultCommand returns the default process command for a service name.
func DefaultCommand(name string) []string {
	return []string{"go", "run", "./cmd/" + name}
}

// ResolveCommand returns cfg.Command or DefaultCommand(name).
func ResolveCommand(cfg ServiceFileConfig) []string {
	if len(cfg.Command) > 0 {
		return append([]string(nil), cfg.Command...)
	}
	name := cfg.Name
	if name == "" {
		name = "service"
	}
	return DefaultCommand(name)
}
