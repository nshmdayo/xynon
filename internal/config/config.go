package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the full proxy configuration.
type Config struct {
	Listen  string         `yaml:"listen"`
	Plugins PluginsConfig  `yaml:"plugins"`
}

// PluginsConfig holds plugin directory and the ordered chain.
type PluginsConfig struct {
	Dir   string        `yaml:"dir"`
	Chain []PluginEntry `yaml:"chain"`
}

// PluginEntry identifies one plugin in the chain.
type PluginEntry struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"` // "wasm" or "rpc"
}

// Limits holds per-plugin sandbox limits.
type Limits struct {
	TimeoutMs uint64 `yaml:"timeout_ms"`
	MemoryMB  uint32 `yaml:"memory_mb"`
}

// Load reads a YAML config file from path and validates required fields.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: cannot read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: invalid YAML in %q: %w", path, err)
	}

	if err := validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save serialises cfg as YAML and writes it to path, overwriting the file.
// Existing comments are not preserved.
func Save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config: marshal: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("config: write %q: %w", path, err)
	}
	return nil
}

func validate(cfg *Config) error {
	var missing []string
	if cfg.Listen == "" {
		missing = append(missing, "listen")
	}
	if cfg.Plugins.Dir == "" {
		missing = append(missing, "plugins.dir")
	}
	if len(missing) > 0 {
		return errors.New("config: missing required fields: " + strings.Join(missing, ", "))
	}
	return nil
}
