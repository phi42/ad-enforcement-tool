package shared

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/phi42/ad-enforcement-tool/internal/domain"
	"github.com/phi42/ad-enforcement-tool/rule"
	"google.golang.org/protobuf/proto"
)

// PluginInfo is the JSON shape returned by a plugin when invoked with --info.
type PluginInfo struct {
	Modes        []string `json:"modes"`
	ConfigPrefix string   `json:"config_prefix"`
}

// QueryPluginInfo runs the plugin with the --info flag and returns the plugin's
// advertised info (modes, config prefix, etc.).
func QueryPluginInfo(plugin string) (*PluginInfo, error) {
	path, err := domain.ResolvePluginPath(plugin)
	if err != nil {
		return nil, err
	}

	var stdout bytes.Buffer
	cmd := exec.Command(path, "--info")
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("querying plugin %q with --info: %w", plugin, err)
	}

	var info PluginInfo
	if err := json.Unmarshal(bytes.TrimSpace(stdout.Bytes()), &info); err != nil {
		return nil, fmt.Errorf("parsing --info response from plugin %q: %w", plugin, err)
	}
	return &info, nil
}

// ValidatePluginMode ensures the plugin advertises the requested mode. It
// returns a descriptive error if the plugin does not support that mode.
func ValidatePluginMode(plugin, requestedMode string) error {
	info, err := QueryPluginInfo(plugin)
	if err != nil {
		return err
	}
	for _, m := range info.Modes {
		if m == requestedMode {
			return nil
		}
	}
	return fmt.Errorf("plugin %q supports modes %v and cannot be used with \"enforce %s\"", plugin, info.Modes, requestedMode)
}

// CompileSpec reads and parses a DSL rule file, returning the Spec.
func CompileSpec(path string) (*rule.Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	ir, err := domain.ParseDSL(string(data))
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return ir, nil
}

// RunPlugin serialises spec as protobuf, pipes it into the plugin process and
// streams stdout/stderr back to the caller's console.
func RunPlugin(plugin string, ir *rule.Spec) error {
	path, err := domain.ResolvePluginPath(plugin)
	if err != nil {
		return err
	}

	payload, err := proto.Marshal(ir)
	if err != nil {
		return fmt.Errorf("unable to marshal Spec to protobuf: %s", err.Error())
	}

	pluginCmd := exec.Command(path)
	pluginCmd.Stdin = bytes.NewReader(payload)
	pluginCmd.Stdout = os.Stdout
	pluginCmd.Stderr = os.Stderr

	if err := pluginCmd.Run(); err != nil {
		return fmt.Errorf("plugin failed: %s", err.Error())
	}
	return nil
}
