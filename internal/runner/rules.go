package runner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CollectRuleFiles returns the list of rule files at path. If path is a
// regular file it is returned verbatim; if it is a directory the tree is
// walked recursively for files with a ".rule" suffix.
func CollectRuleFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("stating %s: %w", path, err)
	}
	if !info.IsDir() {
		return []string{path}, nil
	}

	var files []string
	err = filepath.WalkDir(path, func(p string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(p, ".rule") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking %s: %w", path, err)
	}
	return files, nil
}
