package config

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	cfg "github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurable defaults and their effective values.",
	RunE:  listRun,
}

func listRun(cmd *cobra.Command, args []string) error {
	globalCfg, _ := cfg.GlobalConfigPath()
	cwd, _ := os.Getwd()
	projectCfg := filepath.Join(cwd, cfg.FileName+"."+cfg.FileExt)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE\tSOURCE")
	for _, key := range cfg.KnownDefaults {
		value, source := resolveEffective(key, globalCfg, projectCfg)
		fmt.Fprintf(w, "%s\t%s\t%s\n", key, value, source)
	}

	// Also print every plugin_configs.* entry currently merged into viper.
	allSettings := cfg.Viper().AllSettings()
	if pluginConfigs, ok := allSettings["plugin_configs"].(map[string]interface{}); ok {
		for prefix, entries := range pluginConfigs {
			entryMap, ok := entries.(map[string]interface{})
			if !ok {
				continue
			}
			for k, v := range entryMap {
				key := cfg.PluginConfigsPrefix + prefix + "." + k
				value, source := resolveEffective(key, globalCfg, projectCfg)
				if value == "" {
					// Fallback: value is in the merged viper state (e.g. via
					// AutomaticEnv) but not in any config file on disk.
					value = fmt.Sprintf("%v", v)
					source = "[env]"
				}
				fmt.Fprintf(w, "%s\t%s\t%s\n", key, value, source)
			}
		}
	}
	return w.Flush()
}

// resolveEffective looks up key in the project config first, then the global
// config, mirroring the precedence used during normal command execution.
// When a custom config file is active (--config flag), the custom file is
// checked first (labelled [custom]), then the global file.
func resolveEffective(key, globalCfg, projectCfg string) (string, string) {
	if customCfg := cfg.CustomConfigFile(); customCfg != "" {
		if val, ok, _ := cfg.GetKey(customCfg, key); ok {
			return val, "[custom]"
		}
		if val, ok, _ := cfg.GetKey(globalCfg, key); ok {
			return val, "[global]"
		}
		return "", "[not set]"
	}
	if val, ok, _ := cfg.GetKey(projectCfg, key); ok {
		return val, "[project]"
	}
	if val, ok, _ := cfg.GetKey(globalCfg, key); ok {
		return val, "[global]"
	}
	return "", "[not set]"
}
