package plugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolvePath turns a plugin name or path into an absolute executable path.
//
// If nameOrPath looks like a path (absolute or contains a separator) it is
// returned verbatim after a stat-check. Otherwise the current working
// directory is searched, with ".exe" appended on Windows when the name has no
// extension.
func ResolvePath(nameOrPath string) (string, error) {
	if filepath.IsAbs(nameOrPath) || filepath.Dir(nameOrPath) != "." {
		clean := filepath.Clean(nameOrPath)
		if _, err := os.Stat(clean); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("plugin path does not exist: %s", clean)
			}
			return "", fmt.Errorf("unable to access plugin path %s: %w", clean, err)
		}
		return clean, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("unable to get current working directory: %w", err)
	}
	if path := findExecutable(cwd, nameOrPath); path != "" {
		return path, nil
	}
	return "", fmt.Errorf("plugin not found: tried '%s'", filepath.Join(cwd, nameOrPath))
}

// findExecutable looks for name (and, on Windows, name+".exe") inside dir.
func findExecutable(dir, name string) string {
	path := filepath.Join(dir, name)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	if filepath.Ext(name) == "" {
		withExe := filepath.Join(dir, name+".exe")
		if _, err := os.Stat(withExe); err == nil {
			return withExe
		}
	}
	return ""
}
