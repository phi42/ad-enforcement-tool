package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveConfigPath returns the file path to be used by config set/unset/get.
// Priority order:
//  1. --global flag → global config path
//  2. --config flag (CustomConfigFile) → custom file from the current command
//  3. project config (.ade.yaml in cwd)
func ResolveConfigPath(global bool) (string, error) {
	if global {
		return GlobalConfigPath()
	}
	if custom := CustomConfigFile(); custom != "" {
		return filepath.Abs(custom)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current directory: %w", err)
	}
	return filepath.Join(cwd, FileName+"."+FileExt), nil
}

// ResolveConfigScope returns the human-readable scope label printed alongside
// set/unset confirmations: "[project]", "[global]", or "[custom]".
func ResolveConfigScope(global bool) string {
	if global {
		return "[global]"
	}
	if CustomConfigFile() != "" {
		return "[custom]"
	}
	return "[project]"
}
