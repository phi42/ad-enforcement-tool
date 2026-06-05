// Package cmd assembles the `ade` (Architectural Decision Enforcement)
// command-line tool as a single cobra command tree.
//
// Embedding from a host module:
//
//	import adecmd "github.com/phi42/ad-enforcement-tool/cmd"
//
//	rootCmd.AddCommand(adecmd.NewEnforceCommand())
//
// This is how `adg enforce ...` is wired up: ADG re-uses the same command
// tree under a different parent.
package cmd

import (
	cmdconfig "github.com/phi42/ad-enforcement-tool/cmd/config"
	cmdplugin "github.com/phi42/ad-enforcement-tool/cmd/plugin"
	"github.com/phi42/ad-enforcement-tool/internal/config"
	"github.com/spf13/cobra"
)

var cfgFile string

var enforceCmd = &cobra.Command{
	Use:   "enforce",
	Short: "Enforce architectural decisions using rule files.",
	Long: `Commands for enforcing architectural decisions: validate rule files, ` +
		`compile rules into test artifacts, verify rules against the codebase, ` +
		`and manage plugins and configuration.`,
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		config.Init(cfgFile)
	},
}

// NewEnforceCommand returns the `enforce` subcommand tree. Callers may rename
// the returned command (e.g. set Use="ade" for the standalone binary).
func NewEnforceCommand() *cobra.Command {
	return enforceCmd
}

func init() {
	enforceCmd.CompletionOptions.HiddenDefaultCmd = true
	enforceCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (overrides default hierarchy: global ade.yaml then .ade.yaml in current directory)")

	enforceCmd.AddCommand(compileCmd, verifyCmd, validateCmd)
	enforceCmd.AddCommand(cmdconfig.New())
	enforceCmd.AddCommand(cmdplugin.New())
}
