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

// Viper returns the shared viper instance used by every ADE subcommand.
func Viper() *viper.Viper { return instance }

// FileUsed returns the path of the config file that contributed values to
// the shared viper instance, or "" if none was loaded.
func FileUsed() string { return fileUsed }

// Init populates the shared viper using ADE's two-tier hierarchy:
//   - if cfgFile is non-empty, only that file is loaded;
//   - otherwise the global config (XDG) is read first, then the project
//     .ade.yaml in the current working directory is merged on top.
//
// A missing config file is not an error.
func Init(cfgFile string) {
	if cfgFile != "" {
		instance.SetConfigFile(cfgFile)
		instance.AutomaticEnv()
		if err := instance.ReadInConfig(); err == nil {
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
