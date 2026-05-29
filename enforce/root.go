package enforce

import (
	"os"

	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/phi42/ad-enforcement-tool/internal/pluginstore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Viper instance used by all enforcement commands for config resolution.
var adeViper = viper.New()

var cfgFile string
var configFileUsed string

var enforceCmd = &cobra.Command{
	Use:   "enforce",
	Short: "Enforce architectural decisions using rule files.",
	Long:  `Commands for enforcing architectural decisions: validate rule files, compile rules into test artifacts, verify rules against the codebase, and manage plugins and configuration.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initADEConfig()
	},
}

// NewEnforceCommand returns the 'enforce' subcommand for use as an adg subgroup.
// The returned command has Use="enforce"; callers can rename it for other contexts.
func NewEnforceCommand() *cobra.Command {
	return enforceCmd
}

func init() {
	enforceCmd.CompletionOptions.HiddenDefaultCmd = true
	enforceCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (overrides default hierarchy: global ade.yaml then .ade.yaml in current directory)")
}

func initADEConfig() {
	if cfgFile != "" {
		adeViper.SetConfigFile(cfgFile)
		adeViper.AutomaticEnv()
		if err := adeViper.ReadInConfig(); err == nil {
			configFileUsed = adeViper.ConfigFileUsed()
		}
		return
	}

	adeViper.AutomaticEnv()

	// Load the global config from the XDG config directory as the fallback base.
	if globalCfg, err := pluginstore.GlobalConfigPath(); err == nil {
		if _, err := os.Stat(globalCfg); err == nil {
			adeViper.SetConfigFile(globalCfg)
			if err := adeViper.ReadInConfig(); err == nil {
				configFileUsed = adeViper.ConfigFileUsed()
			}
		}
	}

	// Merge the project-level config on top; its values take precedence.
	if cwd, err := os.Getwd(); err == nil {
		pv := viper.New()
		pv.AddConfigPath(cwd)
		pv.SetConfigType(domain.CONFIG_FILE_EXT)
		pv.SetConfigName(domain.CONFIG_FILE_NAME)
		if err := pv.ReadInConfig(); err == nil {
			if mergeErr := adeViper.MergeConfigMap(pv.AllSettings()); mergeErr == nil {
				configFileUsed = pv.ConfigFileUsed()
			}
		}
	}
}


