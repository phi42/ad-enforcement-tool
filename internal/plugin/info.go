package plugin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// Info is the JSON shape a plugin returns when invoked with --info.
type Info struct {
	Modes        []string `json:"modes"`
	ConfigPrefix string   `json:"config_prefix"`
}

// SupportsMode reports whether the plugin advertises mode in its Modes list.
func (i *Info) SupportsMode(mode string) bool {
	for _, m := range i.Modes {
		if m == mode {
			return true
		}
	}
	return false
}

// QueryInfo runs the plugin with --info and parses the JSON it prints to
// stdout. nameOrPath is resolved by ResolvePath.
func QueryInfo(nameOrPath string) (*Info, error) {
	path, err := ResolvePath(nameOrPath)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd := exec.Command(path, "--info")
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("querying plugin %q with --info: %w", nameOrPath, err)
	}

	var info Info
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &info); err != nil {
		return nil, fmt.Errorf("parsing --info response from plugin %q: %w", nameOrPath, err)
	}
	return &info, nil
}
