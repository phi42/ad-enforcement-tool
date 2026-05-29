package enforce

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/phi42/ad-enforcement-tool/internal/application/shared"
	verifyapp "github.com/phi42/ad-enforcement-tool/internal/application/verify"
	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/spf13/cobra"
)

const (
	FLAG_VERIFY_INPUT       = "input"
	FLAG_VERIFY_INPUT_SHORT = "i"
	FLAG_VERIFY_INPUT_USAGE = "input path containing decision (falls back to defaults.verify.input in config)"

	FLAG_VERIFY_PLUGIN       = "plugin"
	FLAG_VERIFY_PLUGIN_SHORT = "p"
	FLAG_VERIFY_PLUGIN_USAGE = "plugin name or path (falls back to defaults.verify.plugin in config)"
)

var enforceVerifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "Verify rules immediately against the target using the specified plugin.",
	Long: `Execute rules from an ADR rule file directly against the target and report results.

Unlike 'compile', which generates test code for later execution, 'verify'
runs assertions immediately using the plugin and exits non-zero if any rule is violated.

Plugin-specific settings (e.g. root directory) are read from plugin_configs.<prefix>.*
in the active config file.

Examples:
  ade verify -i docs/0003.rule -p fscheck
  ade verify -i docs/ -p fscheck`,
	Run: enforceVerifyCommand,
}

func init() {
	enforceCmd.AddCommand(enforceVerifyCmd)

	enforceVerifyCmd.Flags().StringP(FLAG_VERIFY_INPUT, FLAG_VERIFY_INPUT_SHORT, "", FLAG_VERIFY_INPUT_USAGE)

	enforceVerifyCmd.Flags().StringP(FLAG_VERIFY_PLUGIN, FLAG_VERIFY_PLUGIN_SHORT, "", FLAG_VERIFY_PLUGIN_USAGE)
}

func enforceVerifyCommand(cmd *cobra.Command, args []string) {
	input, err := cmd.Flags().GetString(FLAG_VERIFY_INPUT)
	domain.CheckFatalError(err, "reading input flag")
	if strings.TrimSpace(input) == "" {
		input = adeViper.GetString(domain.CONFIG_DEFAULT_VERIFY_INPUT)
	}
	if strings.TrimSpace(input) == "" {
		domain.CheckFatalError(fmt.Errorf("--input is required (pass as flag or set %s in config)", domain.CONFIG_DEFAULT_VERIFY_INPUT), "resolving input")
	}

	plugin, err := cmd.Flags().GetString(FLAG_VERIFY_PLUGIN)
	domain.CheckFatalError(err, "reading plugin flag")
	if plugin == "" {
		plugin = adeViper.GetString(domain.CONFIG_DEFAULT_VERIFY_PLUGIN)
	}
	if plugin == "" {
		domain.CheckFatalError(fmt.Errorf("--plugin is required (pass as flag or set %s in config)", domain.CONFIG_DEFAULT_VERIFY_PLUGIN), "resolving plugin")
	}
	if !filepath.IsAbs(plugin) && filepath.Dir(plugin) == "." {
		if configPath := adeViper.GetString(domain.CONFIG_PLUGIN_KEY_PREFIX + plugin); configPath != "" {
			plugin = configPath
		}
	}

	info, err := shared.QueryPluginInfo(plugin)
	domain.CheckFatalError(err, "querying plugin info")

	validMode := false
	for _, m := range info.Modes {
		if m == "verify" {
			validMode = true
			break
		}
	}
	if !validMode {
		domain.CheckFatalError(fmt.Errorf("plugin %q supports modes %v and cannot be used with \"enforce verify\"", plugin, info.Modes), "validating plugin mode")
	}

	var pluginConfig map[string]string
	if info.ConfigPrefix != "" {
		pluginConfig = adeViper.GetStringMapString(domain.CONFIG_PLUGIN_CONFIGS_PREFIX + info.ConfigPrefix)
	}

	ruleFiles, err := collectRuleFilePaths(input)
	domain.CheckFatalError(err, "resolving input path")

	for _, f := range ruleFiles {
		verifyapp.Verify(verifyapp.VerifyInfo{
			InputFile:    f,
			PluginName:   plugin,
			PluginConfig: pluginConfig,
		})
	}
}
