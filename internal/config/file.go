package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SetKey sets a dotted key (e.g. "defaults.compile.plugin") to value in the
// YAML config at cfgPath. Intermediate map nodes are created as needed.
func SetKey(cfgPath, dottedKey, value string) error {
	cfg, err := ReadFile(cfgPath)
	if err != nil {
		return err
	}
	setNestedKey(cfg, strings.Split(dottedKey, "."), value)
	return WriteFile(cfgPath, cfg)
}

// UnsetKey removes a dotted key from the YAML config at cfgPath. Empty
// parent maps are pruned. Missing keys are not an error.
func UnsetKey(cfgPath, dottedKey string) error {
	cfg, err := ReadFile(cfgPath)
	if err != nil {
		return err
	}
	unsetNestedKey(cfg, strings.Split(dottedKey, "."))
	return WriteFile(cfgPath, cfg)
}

// GetKey reads a single dotted key from the YAML config at cfgPath. Returns
// the value and true if found, or "" and false otherwise.
func GetKey(cfgPath, dottedKey string) (string, bool, error) {
	cfg, err := ReadFile(cfgPath)
	if err != nil {
		return "", false, err
	}
	var current interface{} = cfg
	for _, p := range strings.Split(dottedKey, ".") {
		m, ok := current.(map[string]interface{})
		if !ok {
			return "", false, nil
		}
		current, ok = m[p]
		if !ok {
			return "", false, nil
		}
	}
	if s, ok := current.(string); ok {
		return s, true, nil
	}
	return fmt.Sprintf("%v", current), true, nil
}

// ReadFile loads cfgPath as a generic YAML map. A missing file yields an
// empty map without error.
func ReadFile(cfgPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]interface{}), nil
		}
		return nil, fmt.Errorf("reading config %s: %w", cfgPath, err)
	}
	var cfg map[string]interface{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", cfgPath, err)
	}
	if cfg == nil {
		return make(map[string]interface{}), nil
	}
	return cfg, nil
}

// WriteFile serialises cfg as YAML and writes it to cfgPath, creating the
// parent directory if needed.
func WriteFile(cfgPath string, cfg map[string]interface{}) error {
	if err := os.MkdirAll(filepath.Dir(cfgPath), 0750); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}
	updated, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("serializing config: %w", err)
	}
	return os.WriteFile(cfgPath, updated, 0600)
}

func setNestedKey(cfg map[string]interface{}, parts []string, value interface{}) {
	for i := 0; i < len(parts)-1; i++ {
		child, ok := cfg[parts[i]].(map[string]interface{})
		if !ok {
			child = make(map[string]interface{})
			cfg[parts[i]] = child
		}
		cfg = child
	}
	cfg[parts[len(parts)-1]] = value
}

func unsetNestedKey(cfg map[string]interface{}, parts []string) {
	if len(parts) == 0 {
		return
	}
	if len(parts) == 1 {
		delete(cfg, parts[0])
		return
	}
	child, ok := cfg[parts[0]].(map[string]interface{})
	if !ok {
		return
	}
	unsetNestedKey(child, parts[1:])
	if len(child) == 0 {
		delete(cfg, parts[0])
	}
}
