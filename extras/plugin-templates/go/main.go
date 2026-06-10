// Plugin template for ADE enforcement plugins written in Go.
//
// This is the minimal skeleton needed to qualify as an ADE plugin:
//   - Responds to --info with a JSON object listing supported modes.
//   - Reads a serialized rule.Spec from stdin and prints its contents.
//
// Replace the printSpec call with your actual compile / verify logic.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/phi42/ad-enforcement-tool/rule"
	"google.golang.org/protobuf/proto"
)

type pluginInfo struct {
	Modes        []string `json:"modes"`
	ConfigPrefix string   `json:"config_prefix"`
}

// info declares which ADE invocation modes this plugin supports and the
// config key prefix used in .rule files for plugin-specific options.
// TODO: remove any modes your plugin does not implement.
// TODO: set ConfigPrefix to the prefix your plugin reads from plugin_config.
var info = pluginInfo{
	Modes:        []string{"compile", "verify"},
	ConfigPrefix: "my-plugin",
}

func main() {
	// --info must be answered before anything else.  ADE calls this before
	// every invocation to verify the plugin supports the requested mode.
	if len(os.Args) == 2 && os.Args[1] == "--info" {
		out, err := json.Marshal(info)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: marshaling plugin info: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(out))
		os.Exit(0)
	}

	// When stdin is a terminal the user ran the binary directly; show help.
	if fi, err := os.Stdin.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) != 0 {
		fmt.Fprintln(os.Stderr, "Usage: pipe an ADE Spec protobuf message to stdin")
		fmt.Fprintln(os.Stderr, "       plugin --info")
		os.Exit(0)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	var spec rule.Spec
	if err := proto.Unmarshal(data, &spec); err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot unmarshal Spec protobuf: %v\n", err)
		os.Exit(1)
	}

	// TODO: replace this with your actual plugin logic.
	printSpec(&spec)
}

// printSpec prints a brief summary to confirm the plugin received and
// deserialised the Spec correctly.  Remove this once you have real logic.
func printSpec(spec *rule.Spec) {
	adr := spec.GetAdr()
	fmt.Printf("received Spec: ADR [%s] %q -- %d selector(s), %d rule(s), mode=%s\n",
		adr.GetId(), adr.GetTitle(),
		len(spec.GetSelectors()), len(spec.GetRules()),
		spec.GetMode())
}
