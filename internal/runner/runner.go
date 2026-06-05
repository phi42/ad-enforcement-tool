package runner

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/phi42/ad-enforcement-tool/internal/dsl"
	"github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/phi42/ad-enforcement-tool/rule"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Mode bundles the invariants that differ between `ade compile` and
// `ade verify`: the user-facing mode name, the protobuf invocation mode the
// plugin receives, and the config keys that supply default flag values.
type Mode struct {
	Name             string
	Invocation       rule.InvocationMode
	DefaultPluginKey string
	DefaultInputKey  string
}

// Run resolves --input and --plugin against cmd's flags (falling back to
// configured defaults), queries the plugin for its modes and config prefix,
// parses each rule file, and invokes the plugin once per file.
func (m Mode) Run(cmd *cobra.Command) error {
	v := config.Viper()
	input, err := m.resolveInput(cmd, v)
	if err != nil {
		return err
	}
	pluginName, err := m.resolvePlugin(cmd, v)
	if err != nil {
		return err
	}

	info, err := plugin.QueryInfo(pluginName)
	if err != nil {
		return fmt.Errorf("querying plugin info: %w", err)
	}
	if !info.SupportsMode(m.Name) {
		return fmt.Errorf("plugin %q supports modes %v and cannot be used with %q",
			pluginName, info.Modes, "enforce "+m.Name)
	}

	var pluginConfig map[string]string
	if info.ConfigPrefix != "" {
		pluginConfig = v.GetStringMapString(config.PluginConfigsPrefix + info.ConfigPrefix)
	}

	ruleFiles, err := CollectRuleFiles(input)
	if err != nil {
		return fmt.Errorf("resolving input path: %w", err)
	}

	for _, file := range ruleFiles {
		spec, err := dsl.ParseFile(file)
		if err != nil {
			return err
		}
		spec.PluginConfig = pluginConfig
		spec.Mode = m.Invocation
		if err := plugin.Run(pluginName, spec); err != nil {
			return fmt.Errorf("running plugin: %w", err)
		}
	}
	return nil
}

func (m Mode) resolveInput(cmd *cobra.Command, v *viper.Viper) (string, error) {
	input, err := cmd.Flags().GetString("input")
	if err != nil {
		return "", fmt.Errorf("reading input flag: %w", err)
	}
	if strings.TrimSpace(input) == "" {
		input = v.GetString(m.DefaultInputKey)
	}
	if strings.TrimSpace(input) == "" {
		return "", fmt.Errorf("--input is required (pass as flag or set %s in config)", m.DefaultInputKey)
	}
	return input, nil
}

func (m Mode) resolvePlugin(cmd *cobra.Command, v *viper.Viper) (string, error) {
	pluginName, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return "", fmt.Errorf("reading plugin flag: %w", err)
	}
	if pluginName == "" {
		pluginName = v.GetString(m.DefaultPluginKey)
	}
	if pluginName == "" {
		return "", fmt.Errorf("--plugin is required (pass as flag or set %s in config)", m.DefaultPluginKey)
	}
	if !filepath.IsAbs(pluginName) && filepath.Dir(pluginName) == "." {
		if configured := v.GetString(config.PluginLocationsPrefix + pluginName); configured != "" {
			pluginName = configured
		}
	}
	return pluginName, nil
}
