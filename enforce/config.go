package enforce

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/phi42/ad-enforcement-tool/internal/pluginstore"
	"github.com/spf13/cobra"
)

var globalFlag bool

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage ADE configuration defaults.",
	Long: `Get, set, and unset default values for command flags and plugin configuration.

Defaults are stored in the project config (.ade.yaml in the current directory)
unless --global is specified, in which case the global config is used.

Configurable keys:
  defaults.compile.plugin    Default plugin for 'compile'
  defaults.compile.input     Default input path for 'compile'
  defaults.verify.plugin     Default plugin for 'verify'
  defaults.verify.input      Default input path for 'verify'

Plugin configuration (open namespace, prefix must match the plugin's config_prefix):
  plugin_configs.<prefix>.<key>    Plugin-specific setting forwarded to the plugin

Examples:
  ade config set defaults.compile.plugin arch-go
  ade config set defaults.compile.input ./docs/adr --global
  ade config get defaults.compile.plugin
  ade config unset defaults.compile.plugin
  ade config set plugin_configs.fscheck.root-dir ./src
  ade config list`,
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration default.",
	Args:  cobra.ExactArgs(2),
	Run:   configSetCommand,
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get the effective value of a configuration key.",
	Args:  cobra.ExactArgs(1),
	Run:   configGetCommand,
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key>",
	Short: "Remove a configuration default.",
	Args:  cobra.ExactArgs(1),
	Run:   configUnsetCommand,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configurable defaults and their effective values.",
	Run:   configListCommand,
}

func init() {
	enforceCmd.AddCommand(configCmd)
	configCmd.PersistentFlags().BoolVar(&globalFlag, "global", false, "target the global config instead of the project config")
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configListCmd)
}

func resolveConfigPath() (string, error) {
	if globalFlag {
		return pluginstore.GlobalConfigPath()
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current directory: %w", err)
	}
	return filepath.Join(cwd, domain.CONFIG_FILE_NAME+"."+domain.CONFIG_FILE_EXT), nil
}

func configSetCommand(cmd *cobra.Command, args []string) {
	key, value := args[0], args[1]
	isKnown := slices.Contains(domain.KnownDefaults, key)
	isPluginConfig := strings.HasPrefix(key, domain.CONFIG_PLUGIN_CONFIGS_PREFIX)
	if !isKnown && !isPluginConfig {
		fmt.Fprintf(os.Stderr, "Error: unknown config key %q\nAllowed keys:\n", key)
		for _, k := range domain.KnownDefaults {
			fmt.Fprintf(os.Stderr, "  %s\n", k)
		}
		fmt.Fprintf(os.Stderr, "  %s<prefix>.<key>  (plugin-specific config)\n", domain.CONFIG_PLUGIN_CONFIGS_PREFIX)
		os.Exit(1)
	}

	cfgPath, err := resolveConfigPath()
	domain.CheckFatalError(err, "resolving config path")

	domain.CheckFatalError(pluginstore.SetDefault(cfgPath, key, value), "setting config value")

	scope := "project"
	if globalFlag {
		scope = "global"
	}
	fmt.Printf("Set %s = %s [%s]\n", key, value, scope)
}

func configGetCommand(cmd *cobra.Command, args []string) {
	key := args[0]
	value := adeViper.GetString(key)
	if value == "" {
		fmt.Fprintf(os.Stderr, "Error: %s is not set\n", key)
		os.Exit(1)
	}
	fmt.Println(value)
}

func configUnsetCommand(cmd *cobra.Command, args []string) {
	key := args[0]
	isKnown := slices.Contains(domain.KnownDefaults, key)
	isPluginConfig := strings.HasPrefix(key, domain.CONFIG_PLUGIN_CONFIGS_PREFIX)
	if !isKnown && !isPluginConfig {
		fmt.Fprintf(os.Stderr, "Error: unknown config key %q\nAllowed keys:\n", key)
		for _, k := range domain.KnownDefaults {
			fmt.Fprintf(os.Stderr, "  %s\n", k)
		}
		fmt.Fprintf(os.Stderr, "  %s<prefix>.<key>  (plugin-specific config)\n", domain.CONFIG_PLUGIN_CONFIGS_PREFIX)
		os.Exit(1)
	}

	cfgPath, err := resolveConfigPath()
	domain.CheckFatalError(err, "resolving config path")

	domain.CheckFatalError(pluginstore.UnsetDefault(cfgPath, key), "unsetting config value")

	scope := "project"
	if globalFlag {
		scope = "global"
	}
	fmt.Printf("Unset %s [%s]\n", key, scope)
}

func configListCommand(cmd *cobra.Command, args []string) {
	globalCfg, _ := pluginstore.GlobalConfigPath()
	cwd, _ := os.Getwd()
	projectCfg := filepath.Join(cwd, domain.CONFIG_FILE_NAME+"."+domain.CONFIG_FILE_EXT)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "KEY\tVALUE\tSOURCE")
	for _, key := range domain.KnownDefaults {
		value, source := resolveEffective(key, globalCfg, projectCfg)
		fmt.Fprintf(w, "%s\t%s\t%s\n", key, value, source)
	}

	// Print all plugin_configs.* entries from the merged config.
	allSettings := adeViper.AllSettings()
	if pluginConfigs, ok := allSettings["plugin_configs"].(map[string]interface{}); ok {
		for prefix, entries := range pluginConfigs {
			if entryMap, ok := entries.(map[string]interface{}); ok {
				for k, v := range entryMap {
					key := domain.CONFIG_PLUGIN_CONFIGS_PREFIX + prefix + "." + k
					value, source := resolveEffective(key, globalCfg, projectCfg)
					if value == "" {
						value = fmt.Sprintf("%v", v)
						source = "[merged]"
					}
					fmt.Fprintf(w, "%s\t%s\t%s\n", key, value, source)
				}
			}
		}
	}
	w.Flush()
}

func resolveEffective(key, globalCfg, projectCfg string) (string, string) {
	// Check project config first (highest precedence).
	if val, ok, _ := pluginstore.GetDefault(projectCfg, key); ok {
		return val, "[project]"
	}
	// Then global config.
	if val, ok, _ := pluginstore.GetDefault(globalCfg, key); ok {
		return val, "[global]"
	}
	return "", "[not set]"
}
