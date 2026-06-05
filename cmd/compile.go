package cmd

import (
	"github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/phi42/ad-enforcement-tool/internal/runner"
	"github.com/phi42/ad-enforcement-tool/rule"
	"github.com/spf13/cobra"
)

var compileMode = runner.Mode{
	Name:             "compile",
	Invocation:       rule.InvocationMode_MODE_COMPILE,
	DefaultPluginKey: config.DefaultCompilePlugin,
	DefaultInputKey:  config.DefaultCompileInput,
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile rules into an executable test artifact using the specified plugin.",
	Long: `Compile rules from an ADR rule file into an executable test artifact.

The plugin generates test code (e.g. a Go test file) in the output directory.
Plugin-specific settings (e.g. output directory) are read from plugin_configs.<prefix>.*
in the active config file. Run the generated tests separately to validate the rules.

Examples:
  ade compile -i docs/0001.rule -p archgo
  ade compile -i docs/ -p archgo`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return compileMode.Run(cmd)
	},
}

func init() {
	compileCmd.Flags().StringP("input", "i", "",
		"path to a .rule file or a directory of .rule files (falls back to "+config.DefaultCompileInput+" in config)")
	compileCmd.Flags().StringP("plugin", "p", "",
		"plugin name or path (falls back to "+config.DefaultCompilePlugin+" in config)")
}
