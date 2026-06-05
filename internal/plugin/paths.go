package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// GlobalDir returns the per-user directory where ADE stores plugin binaries
// (Linux/macOS: $XDG_DATA_HOME/ade/plugins; Windows: %APPDATA%\ade\plugins).
func GlobalDir() (string, error) {
	base, err := xdgDataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "ade", "plugins"), nil
}

// BinaryName returns the OS-appropriate filename for a plugin binary
// (".exe" appended on Windows when missing).
func BinaryName(name string) string {
	if runtime.GOOS == "windows" && filepath.Ext(name) != ".exe" {
		return name + ".exe"
	}
	return name
}

// xdgDataHome returns the XDG data home for the current platform.
// On Windows it falls back to %APPDATA%, then $HOME/AppData/Roaming.
// On other platforms it falls back to $HOME/.local/share.
func xdgDataHome() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
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
	return filepath.Join(home, ".local", "share"), nil
}
