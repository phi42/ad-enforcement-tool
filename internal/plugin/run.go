package plugin

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/phi42/ad-enforcement-tool/rule"
	"google.golang.org/protobuf/proto"
)

// Run invokes the plugin at nameOrPath with the serialised spec piped to its
// stdin. The plugin's stdout and stderr are forwarded to the host's. Any
// non-zero exit from the plugin surfaces as an error.
func Run(nameOrPath string, spec *rule.Spec) error {
	path, err := ResolvePath(nameOrPath)
	if err != nil {
		return err
	}

	payload, err := proto.Marshal(spec)
	if err != nil {
		return fmt.Errorf("marshalling rule.Spec to protobuf: %w", err)
	}

	cmd := exec.Command(path)
	cmd.Stdin = bytes.NewReader(payload)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("plugin %q failed: %w", nameOrPath, err)
	}
	return nil
}
