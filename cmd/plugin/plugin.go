// Package plugin implements the `ade plugin` subcommand group.
package plugin

import "github.com/spf13/cobra"

// New returns the `plugin` parent command with its subcommands attached.
func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage enforcement plugins.",
		Long:  `Install, uninstall, update, and list enforcement plugins.`,
	}
	cmd.AddCommand(installCmd, uninstallCmd, updateCmd, listCmd)
	return cmd
}
