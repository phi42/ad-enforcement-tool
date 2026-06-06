package plugin

import (
	"fmt"
	"os"
	"sort"
	"strings"

	pkg "github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <name>",
	Short: "Re-fetch a release for a remotely installed plugin.",
	Long: `Download a GitHub release for a plugin that was previously installed
with 'ade plugin install <name> --repo github.com/owner/repo' and overwrite the existing binary.

  ade plugin update fscheck

By default the latest release is fetched. If the installed version already
matches the latest, the download is skipped. Use --version to pin a specific
tag (downgrading is allowed):

  ade plugin update fscheck --version v0.1.1

Use --all to update every remotely installed plugin to its latest release.
--all and --version cannot be combined:

  ade plugin update --all

Plugins installed with --path (local mode) cannot be updated this way because
no remote source was recorded. Re-run 'ade plugin install <name> --path <new-path>'
to replace a locally installed plugin.

Authentication

If the GITHUB_TOKEN environment variable is set, it is forwarded as a Bearer
token on every GitHub API and download request. This is required for private
repositories. On GitHub Actions runners GITHUB_TOKEN is set automatically.

For public repositories the variable is optional, but unauthenticated requests
are subject to a rate limit of 60 per hour. If you hit the limit you will see a
403 error. Authenticating raises the limit to 5 000 per hour. A personal access
token with no extra scopes is sufficient. The easiest way to obtain one is via
the GitHub CLI:

  gh auth token

Then export it in your shell profile, for example:

  $env:GITHUB_TOKEN = (gh auth token)   # PowerShell
  export GITHUB_TOKEN=$(gh auth token)  # bash / zsh`,
	Args: cobra.ArbitraryArgs,
	RunE: updateRun,
}

func init() {
	updateCmd.Flags().StringP("version", "v", "", "release tag to fetch (e.g. v0.1.1); defaults to latest")
	updateCmd.Flags().BoolP("all", "a", false, "update all remotely installed plugins to their latest release")
}

func updateRun(cmd *cobra.Command, args []string) error {
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return fmt.Errorf("reading all flag: %w", err)
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return fmt.Errorf("reading version flag: %w", err)
	}

	if all && len(args) > 0 {
		return fmt.Errorf("cannot specify a plugin name together with --all")
	}
	if !all && len(args) != 1 {
		return fmt.Errorf("expected exactly one plugin name, or use --all")
	}
	if all && version != "" {
		return fmt.Errorf("--version and --all cannot be combined")
	}

	plugins, sources, versions, err := pkg.ReadRegistry()
	if err != nil {
		return fmt.Errorf("reading global config: %w", err)
	}

	if all {
		// Collect remote plugins in sorted order for deterministic output.
		// Only include names that have both a registered binary path and a
		// remote source; stale source-only entries are silently skipped.
		var names []string
		for name, source := range sources {
			if source != "" {
				if _, hasPath := plugins[name]; hasPath {
					names = append(names, name)
				}
			}
		}
		sort.Strings(names)

		if len(names) == 0 {
			fmt.Println("no remotely installed plugins found")
			return nil
		}

		var failed []string
		for _, name := range names {
			if err := updateOne(name, plugins[name], sources[name], "", versions[name]); err != nil {
				fmt.Fprintf(os.Stderr, "error updating %q: %v\n", name, err)
				failed = append(failed, name)
			}
		}
		if len(failed) > 0 {
			return fmt.Errorf("failed to update: %s", strings.Join(failed, ", "))
		}
		return nil
	}

	name := args[0]
	binaryPath, registered := plugins[name]
	if !registered {
		return fmt.Errorf("plugin %q is not registered; run 'ade plugin install' first", name)
	}
	source := sources[name]
	if source == "" {
		return fmt.Errorf("plugin %q was installed locally (no remote source recorded); "+
			"use 'ade plugin install %s --path <new-path>' to update it", name, name)
	}
	return updateOne(name, binaryPath, source, version, versions[name])
}

// updateOne downloads the given version (or latest when version is "") for a
// single plugin and updates the registry. When no version is requested and the
// stored tag already matches the latest release, the download is skipped.
func updateOne(name, binaryPath, source, version, installedTag string) error {
	// When no explicit version is requested, check whether the installed tag
	// already matches the latest release before committing to a download.
	if version == "" && installedTag != "" {
		latestTag, err := pkg.FetchLatestTag(source)
		if err == nil && latestTag == installedTag {
			fmt.Printf("plugin %q is already up to date (%s)\n", name, installedTag)
			return nil
		}
	}

	fetchURL := source
	if version != "" {
		fetchURL = source + "@" + version
	}

	resolvedTag, err := pkg.FetchRelease(fetchURL, binaryPath)
	if err != nil {
		return fmt.Errorf("fetching release: %w", err)
	}
	resolvedTag = pkg.StripVersionPrefix(resolvedTag)
	if err := pkg.SetExecutable(binaryPath); err != nil {
		return fmt.Errorf("setting executable permission: %w", err)
	}
	if err := pkg.UpdateRegistry(name, binaryPath, source, resolvedTag); err != nil {
		return fmt.Errorf("updating version in global config: %w", err)
	}

	fmt.Printf("updated plugin %q to %s\n", name, resolvedTag)
	return nil
}
