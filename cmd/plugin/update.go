package plugin

import (
	"fmt"

	pkg "github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Re-fetch the latest release for a remotely installed plugin.",
	Long: `Download the latest GitHub release for a plugin that was previously installed
with 'ade plugin install <name> --repo github.com/owner/repo' and overwrite the existing binary.

  ade plugin update fscheck

Plugins installed with --path (local mode) cannot be updated this way because
no remote source was recorded. Re-run 'ade plugin install <name> --path <new-path>'
to replace a locally installed plugin.

Authentication

If the GITHUB_TOKEN environment variable is set, it is forwarded as a Bearer
token on every GitHub API and download request. This is required for private
repositories and is ignored for public ones.`,
	Args: cobra.ExactArgs(1),
	RunE: updateRun,
}

func updateRun(cmd *cobra.Command, args []string) error {
	name := args[0]

	plugins, sources, err := pkg.ReadRegistry()
	if err != nil {
		return fmt.Errorf("reading global config: %w", err)
	}
	binaryPath, registered := plugins[name]
	if !registered {
		return fmt.Errorf("plugin %q is not registered; run 'ade plugin install' first", name)
	}
	source := sources[name]
	if source == "" {
		return fmt.Errorf("plugin %q was installed locally (no remote source recorded); "+
			"use 'ade plugin install %s --path <new-path>' to update it", name, name)
	}

	if err := pkg.FetchRelease(source, binaryPath); err != nil {
		return fmt.Errorf("fetching release: %w", err)
	}
	if err := pkg.SetExecutable(binaryPath); err != nil {
		return fmt.Errorf("setting executable permission: %w", err)
	}

	fmt.Printf("updated plugin %q at %s\n", name, binaryPath)
	return nil
}
