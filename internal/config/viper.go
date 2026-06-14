package config

import (
	"os"

	"github.com/spf13/viper"
)

// instance is the process-wide viper that all ADE commands read from.
var instance = viper.New()

// fileUsed is the path of whichever config file actually contributed values
// to instance, or "" if no file was loaded.
var fileUsed string

// customConfigFile holds the path passed via --config, or "" if not used.
var customConfigFile string

// Viper returns the shared viper instance used by every ADE subcommand.
func Viper() *viper.Viper { return instance }

// FileUsed returns the path of the config file that contributed values to
// the shared viper instance, or "" if none was loaded.
func FileUsed() string { return fileUsed }

// CustomConfigFile returns the path passed via --config, or "" if the default
// hierarchy (global + project) is being used.
func CustomConfigFile() string { return customConfigFile }

// Init populates the shared viper using ADE's two-tier hierarchy:
//   - if cfgFile is non-empty, the global config (XDG) is read first (so that
//     plugin installation info is always available), then cfgFile is merged on
//     top; this mirrors the global→project merge used in the default path.
//   - otherwise the global config (XDG) is read first, then the project
//     .ade.yaml in the current working directory is merged on top.
//
// A missing config file is not an error.
func Init(cfgFile string) {
	if cfgFile != "" {
		customConfigFile = cfgFile
		instance.AutomaticEnv()

		// Load global config as the base so plugin info is always available.
		if globalCfg, err := GlobalConfigPath(); err == nil {
			if _, err := os.Stat(globalCfg); err == nil {
				instance.SetConfigFile(globalCfg)
				if err := instance.ReadInConfig(); err == nil {
					fileUsed = instance.ConfigFileUsed()
				}
			}
		}

		// Merge the custom config on top; its values win over global.
		instance.SetConfigFile(cfgFile)
		if err := instance.MergeInConfig(); err == nil {
			fileUsed = instance.ConfigFileUsed()
		}
		return
	}

	instance.AutomaticEnv()

	if globalCfg, err := GlobalConfigPath(); err == nil {
		if _, err := os.Stat(globalCfg); err == nil {
			instance.SetConfigFile(globalCfg)
			if err := instance.ReadInConfig(); err == nil {
				fileUsed = instance.ConfigFileUsed()
			}
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	projectCfg := cwd + string(os.PathSeparator) + FileName + "." + FileExt
	if _, err := os.Stat(projectCfg); err != nil {
		return
	}
	instance.SetConfigFile(projectCfg)
	if err := instance.MergeInConfig(); err == nil {
		fileUsed = instance.ConfigFileUsed()
	}
}
