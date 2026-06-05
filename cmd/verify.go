package cmd

import (
	"github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/phi42/ad-enforcement-tool/internal/runner"
	"github.com/phi42/ad-enforcement-tool/rule"
	"github.com/spf13/cobra"
)

var verifyMode = runner.Mode{
	Name:             "verify",
	Invocation:       rule.InvocationMode_MODE_VERIFY,
	DefaultPluginKey: config.DefaultVerifyPlugin,
	DefaultInputKey:  config.DefaultVerifyInput,
}

var verifyCmd = &cobra.Command{
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return verifyMode.Run(cmd)
	},
}

func init() {
	verifyCmd.Flags().StringP("input", "i", "",
		"input path containing decision (falls back to "+config.DefaultVerifyInput+" in config)")
	verifyCmd.Flags().StringP("plugin", "p", "",
		"plugin name or path (falls back to "+config.DefaultVerifyPlugin+" in config)")
}
