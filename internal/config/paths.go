package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// xdgConfigHome returns the XDG config home for the current platform.
// On Windows it falls back to %APPDATA%, then $HOME/AppData/Roaming.
// On other platforms it falls back to $HOME/.config.
func xdgConfigHome() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	if runtime.GOOS == "windows" {
		if dir := os.Getenv("APPDATA"); dir != "" {
			return dir, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine home directory: %w", err)
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, "AppData", "Roaming"), nil
	}
	return filepath.Join(home, ".config"), nil
}

// GlobalConfigPath returns the per-user ADE config file path
// (Linux/macOS: $XDG_CONFIG_HOME/ade/ade.yaml; Windows: %APPDATA%\ade\ade.yaml).
func GlobalConfigPath() (string, error) {
	base, err := xdgConfigHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ade", "ade.yaml"), nil
}
