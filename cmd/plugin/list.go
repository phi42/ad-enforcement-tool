package plugin

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	pkg "github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all globally registered plugins.",
	Long: `Print all plugins registered in the global config file with their paths and status.

The global config is stored in the platform config directory:
  Linux/macOS: $XDG_CONFIG_HOME/ade/ade.yaml (default: ~/.config/ade/ade.yaml)
  Windows:     %APPDATA%\ade\ade.yaml

The STATUS column shows:
  ok       — binary exists at the registered path
  missing  — binary not found (may need to reinstall)

The SOURCE column shows the GitHub module URL for remotely installed plugins,
or "(local)" for plugins installed with --path.`,
	Args: cobra.NoArgs,
	RunE: listRun,
}

func listRun(cmd *cobra.Command, args []string) error {
	plugins, sources, versions, err := pkg.ReadRegistry()
	if err != nil {
		return fmt.Errorf("reading global config: %w", err)
	}
	if len(plugins) == 0 {
		fmt.Println("no plugins registered; use 'ade plugin install' to add one")
		return nil
	}

	names := make([]string, 0, len(plugins))
	for name := range plugins {
		names = append(names, name)
	}
	sort.Strings(names)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "PLUGIN\tPATH\tSTATUS\tVERSION\tSOURCE")
	for _, name := range names {
		path := plugins[name]
		status := "ok"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			status = "missing"
		}
		source := sources[name]
		version := ""
		if source == "" {
			source = "(local)"
		} else {
			version = versions[name]
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", name, path, status, version, source)
	}
	return w.Flush()
}
