package plugin

import (
	"fmt"
	"path/filepath"

	pkg "github.com/phi42/ad-enforcement-tool/internal/plugin"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a plugin into the global plugin directory.",
	Long: `Install a plugin and register it in the global config.

The global config is stored in the platform config directory:
  Linux/macOS: $XDG_CONFIG_HOME/ade/ade.yaml (default: ~/.config/ade/ade.yaml)
  Windows:     %APPDATA%\ade\ade.yaml

The <name> argument is always the name under which the plugin is registered.
Exactly one of --path or --repo must be provided.

Local mode: register an existing binary you have already built or downloaded:

  ade plugin install fscheck --path /tmp/downloads/fscheck

Remote mode: download directly from a GitHub release:

  ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo

An optional "@version" suffix pins a specific release tag:

  ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo@v0.1.1

Without a version suffix the latest release is downloaded. Each plugin name
can only be installed once; to switch versions use 'ade plugin update'.

The release must contain assets whose filenames include the OS and architecture,
for example:

  my-plugin-linux-amd64
  my-plugin-darwin-arm64
  my-plugin-windows-amd64.exe

The binary is copied to the platform data directory:
  Linux/macOS: $XDG_DATA_HOME/ade/plugins/<name> (default: ~/.local/share/ade/plugins/<name>)
  Windows:     %APPDATA%\ade\plugins\<name>

An entry is added under plugin_locations.<name> in the global config file.

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
	Args: cobra.ExactArgs(1),
	RunE: installRun,
}

func init() {
	installCmd.Flags().StringP("path", "p", "", "path to an existing local binary to register")
	installCmd.Flags().StringP("repo", "r", "", "GitHub module path to download from (e.g. github.com/owner/repo)")
	installCmd.MarkFlagsMutuallyExclusive("path", "repo")
}

func installRun(cmd *cobra.Command, args []string) error {
	name := args[0]

	localPath, err := cmd.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("reading path flag: %w", err)
	}
	repoURL, err := cmd.Flags().GetString("repo")
	if err != nil {
		return fmt.Errorf("reading repo flag: %w", err)
	}
	if localPath == "" && repoURL == "" {
		return fmt.Errorf("one of --path or --repo is required")
	}

	// Remote installs are rejected when the plugin name is already registered;
	// the user must use 'ade plugin update' to change the version.
	// Local installs are allowed to overwrite an existing entry.
	if repoURL != "" {
		registered, _, _, err := pkg.ReadRegistry()
		if err != nil {
			return fmt.Errorf("reading global config: %w", err)
		}
		if _, exists := registered[name]; exists {
			return fmt.Errorf("plugin %q is already installed; to change its version run: ade plugin update %s --version <version>", name, name)
		}
	}

	pluginDir, err := pkg.GlobalDir()
	if err != nil {
		return fmt.Errorf("resolving plugin directory: %w", err)
	}
	dst := filepath.Join(pluginDir, pkg.BinaryName(name))

	var source, version string
	if localPath != "" {
		if err := pkg.CopyBinary(localPath, dst); err != nil {
			return fmt.Errorf("copying binary: %w", err)
		}
	} else {
		repoURL = pkg.NormaliseModuleURL(repoURL)
		resolvedTag, err := pkg.FetchRelease(repoURL, dst)
		if err != nil {
			return fmt.Errorf("fetching release: %w", err)
		}
		// Store the base module URL (without @version) as the source so that a
		// plain 'ade plugin update' always resolves against the repo, and store
		// the actual downloaded tag as the version.
		source, _ = pkg.SplitVersion(repoURL)
		version = resolvedTag
	}
	if err := pkg.SetExecutable(dst); err != nil {
		return fmt.Errorf("setting executable permission: %w", err)
	}
	if err := pkg.UpdateRegistry(name, dst, source, version); err != nil {
		return fmt.Errorf("updating global config: %w", err)
	}

	fmt.Printf("installed plugin %q to %s\n", name, dst)
	return nil
}
