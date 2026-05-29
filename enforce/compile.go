package enforce

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	compileapp "github.com/phi42/ad-enforcement-tool/internal/application/compile"
	"github.com/phi42/ad-enforcement-tool/internal/application/shared"
	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/spf13/cobra"
)

const (
	FLAG_COMPILE_INPUT       = "input"
	FLAG_COMPILE_INPUT_SHORT = "i"
	FLAG_COMPILE_INPUT_USAGE = "path to a .rule file or a directory of .rule files (falls back to defaults.compile.input in config)"

	FLAG_COMPILE_PLUGIN       = "plugin"
	FLAG_COMPILE_PLUGIN_SHORT = "p"
	FLAG_COMPILE_PLUGIN_USAGE = "plugin name or path (falls back to defaults.compile.plugin in config)"
)

var enforceCompileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile rules into an executable test artifact using the specified plugin.",
	Long: `Compile rules from an ADR rule file into an executable test artifact.

The plugin generates test code (e.g. a Go test file) in the output directory.
Plugin-specific settings (e.g. output directory) are read from plugin_configs.<prefix>.*
in the active config file. Run the generated tests separately to validate the rules.

Examples:
  ade compile -i docs/0001.rule -p arch-go
  ade compile -i docs/ -p arch-go`,
	Run: enforceCompileCommand,
}

func init() {
	enforceCmd.AddCommand(enforceCompileCmd)

	enforceCompileCmd.Flags().StringP(FLAG_COMPILE_INPUT, FLAG_COMPILE_INPUT_SHORT, "", FLAG_COMPILE_INPUT_USAGE)

	enforceCompileCmd.Flags().StringP(FLAG_COMPILE_PLUGIN, FLAG_COMPILE_PLUGIN_SHORT, "", FLAG_COMPILE_PLUGIN_USAGE)
}

func enforceCompileCommand(cmd *cobra.Command, args []string) {
	input, err := cmd.Flags().GetString(FLAG_COMPILE_INPUT)
	domain.CheckFatalError(err, "reading input flag")
	if strings.TrimSpace(input) == "" {
		input = adeViper.GetString(domain.CONFIG_DEFAULT_COMPILE_INPUT)
	}
	if strings.TrimSpace(input) == "" {
		domain.CheckFatalError(fmt.Errorf("--input is required (pass as flag or set %s in config)", domain.CONFIG_DEFAULT_COMPILE_INPUT), "resolving input")
	}

	plugin, err := cmd.Flags().GetString(FLAG_COMPILE_PLUGIN)
	domain.CheckFatalError(err, "reading plugin flag")
	if plugin == "" {
		plugin = adeViper.GetString(domain.CONFIG_DEFAULT_COMPILE_PLUGIN)
	}
	if plugin == "" {
		domain.CheckFatalError(fmt.Errorf("--plugin is required (pass as flag or set %s in config)", domain.CONFIG_DEFAULT_COMPILE_PLUGIN), "resolving plugin")
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
		if m == "compile" {
			validMode = true
			break
		}
	}
	if !validMode {
		domain.CheckFatalError(fmt.Errorf("plugin %q supports modes %v and cannot be used with \"enforce compile\"", plugin, info.Modes), "validating plugin mode")
	}

	var pluginConfig map[string]string
	if info.ConfigPrefix != "" {
		pluginConfig = adeViper.GetStringMapString(domain.CONFIG_PLUGIN_CONFIGS_PREFIX + info.ConfigPrefix)
	}

	ruleFiles, err := collectRuleFilePaths(input)
	domain.CheckFatalError(err, "resolving input path")

	for _, f := range ruleFiles {
		compileapp.Compile(compileapp.CompileInfo{
			InputFile:    f,
			PluginName:   plugin,
			PluginConfig: pluginConfig,
		})
	}
}

// collectRuleFilePaths returns .rule file paths from a file or directory.
func collectRuleFilePaths(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{path}, nil
	}
	var files []string
	err = filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !fi.IsDir() && strings.HasSuffix(p, ".rule") {
			files = append(files, p)
		}
		return nil
	})
	return files, err
}
