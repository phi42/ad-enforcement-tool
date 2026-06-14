// Package config implements the `ade config` subcommand group.
package config

import "github.com/spf13/cobra"

var globalFlag bool

// New returns the `config` parent command with its subcommands attached.
func New() *cobra.Command {
	cmd := &cobra.Command{
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
  ade config set defaults.compile.plugin archgo
  ade config set defaults.compile.input ./docs/adr --global
  ade config set plugin_configs.netarchtest.test-project ./src/Tests/ArchTests/Project.csproj --config ./custom.yaml
  ade config get defaults.compile.plugin
  ade config unset defaults.compile.plugin
  ade config set plugin_configs.fscheck.root-dir ./src
  ade config list`,
	}

	cmd.PersistentFlags().BoolVar(&globalFlag, "global", false,
		"target the global config instead of the project config")

	cmd.AddCommand(setCmd, getCmd, unsetCmd, listCmd)
	return cmd
}
