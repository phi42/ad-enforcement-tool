package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CopyBinary copies the file at src to dst, creating parent directories as
// needed. Existing files at dst are truncated.
func CopyBinary(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("creating plugin directory: %w", err)
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0750)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying binary: %w", err)
	}
	return nil
}

// SetExecutable marks the file at path as executable (no-op on Windows, where
// executability is filename-extension driven).
func SetExecutable(path string) error {
	if runtime.GOOS == "windows" {
		return nil
	}
	return os.Chmod(path, 0750)
}

// NormaliseModuleURL strips any http(s):// prefix from a module URL.
func NormaliseModuleURL(moduleURL string) string {
	moduleURL = strings.TrimPrefix(moduleURL, "https://")
	moduleURL = strings.TrimPrefix(moduleURL, "http://")
	return moduleURL
}

// ParseModuleURL splits a "github.com/<owner>/<repo>" module URL into owner
// and repo. An optional "https://" prefix is stripped before parsing.
func ParseModuleURL(moduleURL string) (owner, repo string, err error) {
	moduleURL = NormaliseModuleURL(moduleURL)
	parts := strings.SplitN(moduleURL, "/", 3)
	if len(parts) != 3 || parts[0] != "github.com" {
		return "", "", fmt.Errorf("invalid module URL %q: expected github.com/<owner>/<repo>", moduleURL)
	}
	return parts[1], parts[2], nil
}
