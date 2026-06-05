package plugin

import (
	"github.com/phi42/ad-enforcement-tool/internal/config"
)

// UpdateRegistry writes or updates the entry for plugin name in the global
// config file: plugin_locations.<name> -> binaryPath, and (if source is
// non-empty) plugin_sources.<name> -> source.
func UpdateRegistry(name, binaryPath, source string) error {
	cfgPath, err := config.GlobalConfigPath()
	if err != nil {
		return err
	}
	cfg, err := config.ReadFile(cfgPath)
	if err != nil {
		return err
	}

	plugins, _ := cfg["plugin_locations"].(map[string]interface{})
	if plugins == nil {
		plugins = make(map[string]interface{})
		cfg["plugin_locations"] = plugins
	}
	plugins[name] = binaryPath

	if source != "" {
		sources, _ := cfg["plugin_sources"].(map[string]interface{})
		if sources == nil {
			sources = make(map[string]interface{})
			cfg["plugin_sources"] = sources
		}
		sources[name] = source
	}

	return config.WriteFile(cfgPath, cfg)
}

// RemoveFromRegistry deletes the plugin and its remembered source entry from
// the global config file. It is not an error if the plugin is absent.
func RemoveFromRegistry(name string) error {
	cfgPath, err := config.GlobalConfigPath()
	if err != nil {
		return err
	}
	cfg, err := config.ReadFile(cfgPath)
	if err != nil {
		return err
	}
	if cfg == nil {
		return nil
	}
	if plugins, ok := cfg["plugin_locations"].(map[string]interface{}); ok {
		delete(plugins, name)
	}
	if sources, ok := cfg["plugin_sources"].(map[string]interface{}); ok {
		delete(sources, name)
	}
	return config.WriteFile(cfgPath, cfg)
}

// ReadRegistry returns the plugin_locations and plugin_sources maps from the
// global config file. Both maps are non-nil even when the file is missing.
func ReadRegistry() (plugins map[string]string, sources map[string]string, err error) {
	plugins = make(map[string]string)
	sources = make(map[string]string)

	cfgPath, err := config.GlobalConfigPath()
	if err != nil {
		return nil, nil, err
	}
	cfg, err := config.ReadFile(cfgPath)
	if err != nil {
		return nil, nil, err
	}

	if raw, ok := cfg["plugin_locations"].(map[string]interface{}); ok {
		for k, v := range raw {
			if s, ok := v.(string); ok {
				plugins[k] = s
			}
		}
	}
	if raw, ok := cfg["plugin_sources"].(map[string]interface{}); ok {
		for k, v := range raw {
			if s, ok := v.(string); ok {
				sources[k] = s
			}
		}
	}
	return plugins, sources, nil
}
