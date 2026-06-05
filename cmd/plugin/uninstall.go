package plugin

import (
	"fmt"
	"os"

	pkg "github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <name>",
	Short: "Remove a plugin from the global plugin directory.",
	Long: `Remove a globally installed plugin binary and its entry in the global config file.

  ade plugin uninstall fscheck`,
	Args: cobra.ExactArgs(1),
	RunE: uninstallRun,
}

func uninstallRun(cmd *cobra.Command, args []string) error {
	name := args[0]

	plugins, _, err := pkg.ReadRegistry()
	if err != nil {
		return fmt.Errorf("reading global config: %w", err)
	}
	binaryPath, registered := plugins[name]
	if !registered {
		return fmt.Errorf("plugin %q is not registered in the global config", name)
	}

	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing plugin binary: %w", err)
	}
	if err := pkg.RemoveFromRegistry(name); err != nil {
		return fmt.Errorf("updating global config: %w", err)
	}

	fmt.Printf("uninstalled plugin %q\n", name)
	return nil
}
