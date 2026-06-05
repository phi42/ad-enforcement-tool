package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveConfigPath returns the file path to be used by config set/unset/get.
// If configFile is non-empty it takes precedence (resolved to an absolute
// path). If global is true the global config path is returned. Otherwise the
// project config in the current working directory is used.
func ResolveConfigPath(configFile string, global bool) (string, error) {
	if configFile != "" {
		return filepath.Abs(configFile)
	}
	if global {
		return GlobalConfigPath()
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current directory: %w", err)
	}
	return filepath.Join(cwd, FileName+"."+FileExt), nil
}

// ResolveConfigScope returns the human-readable scope label printed alongside
// set/unset confirmations, e.g. "[project]", "[global]", or the file path.
func ResolveConfigScope(configFile string, global bool) string {
	if configFile != "" {
		return "[" + configFile + "]"
	}
	if global {
		return "[global]"
	}
	return "[project]"
}
