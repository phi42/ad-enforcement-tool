# ADE CLI Reference

Complete reference for every `ade` command and flag. For a guided walkthrough of installing the tool, writing a rule, and wiring it into a pipeline, see the [user guide](user-guide.md). For an introduction to the project, see the [README](../README.md).

## Global flags

| Flag              | Description                                                                                                |
| ----------------- | ---------------------------------------------------------------------------------------------------------- |
| `--config <path>` | Load a specific config file, bypassing the default global + project hierarchy. Applies to every subcommand. |

## `ade validate`

Validate one or more rule files for syntax and semantic errors. No plugin is invoked.

```sh
ade validate -i rules/0001.rule
ade validate -i rules/                       # every .rule file in a directory
ade validate -i rules/0001.rule -i rules/0002.rule
```

| Flag            | Description                                                                       |
| --------------- | --------------------------------------------------------------------------------- |
| `-i`, `--input` | Path to a `.rule` file or a directory of `.rule` files (required, repeatable).    |

## `ade compile`

Compile rules into an executable test artifact using a plugin (e.g. an arch-go or NetArchTest test class).

```sh
ade compile -i docs/0001.rule -p archgo
ade compile -i docs/        -p archgo
```

| Flag             | Description                                                                                  |
| ---------------- | -------------------------------------------------------------------------------------------- |
| `-i`, `--input`  | Path to a `.rule` file or a directory of `.rule` files. Falls back to `defaults.compile.input`. |
| `-p`, `--plugin` | Plugin name or path (see [Plugin resolution](#plugin-resolution)). Falls back to `defaults.compile.plugin`. |

Plugin-specific settings (e.g. the output directory) are read from `plugin_configs.<prefix>.*` in the active config file and forwarded to the plugin in `rule.Spec.PluginConfig`. See [Plugin configuration](#plugin-configuration).

## `ade verify`

Execute rules immediately against the target and exit non-zero if any `error`-severity rule is violated.

```sh
ade verify -i docs/0003.rule -p fscheck
ade verify -i docs/         -p fscheck
```

| Flag             | Description                                                                                |
| ---------------- | ------------------------------------------------------------------------------------------ |
| `-i`, `--input`  | Path to a `.rule` file or a directory of `.rule` files. Falls back to `defaults.verify.input`. |
| `-p`, `--plugin` | Plugin name or path (see [Plugin resolution](#plugin-resolution)). Falls back to `defaults.verify.plugin`. |

Plugin-specific settings (e.g. the root directory for resolving path patterns) are read from `plugin_configs.<prefix>.*` and forwarded to the plugin. See [Plugin configuration](#plugin-configuration).

### Plugin resolution

The `-p` / `--plugin` flag accepts either a name or a path:

- Name (e.g. `archgo`, `netarchtest`) is resolved against `plugin_locations.<name>` entries in the merged config (global first, then project on top). If no entry is found, the current working directory is tried as a fallback. On Windows, `.exe` is appended automatically when the name has no extension.
- Path (absolute or relative) is always used directly.

## `ade plugin install`

Register a plugin binary with ADE. Exactly one of `--path` or `--repo` must be provided.

**Local mode.** Register an existing binary you built or downloaded:

```sh
ade plugin install fscheck --path ./plugins/fscheck/fscheck
```

The binary is copied into the platform data directory and recorded in the global config. Re-running the command with a new `--path` overwrites the existing entry.

**Remote mode.** Download from a GitHub release:

```sh
ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo
```

Without a version suffix, the latest release is downloaded. Pin a specific tag with `@version`:

```sh
ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo@v0.1.1
```

Each plugin name can only be installed once in remote mode; use `ade plugin update` to switch versions.

| Flag             | Description                                                              |
| ---------------- | ------------------------------------------------------------------------ |
| `-p`, `--path`   | Path to an existing local binary to register (local mode).               |
| `-r`, `--repo`   | GitHub module path to download from, e.g. `github.com/owner/repo` (remote mode). |

### Plugin binary location

The installed binary lives in the platform data directory:

| Platform      | Location                                                                           |
| ------------- | ---------------------------------------------------------------------------------- |
| Linux / macOS | `$XDG_DATA_HOME/ade/plugins/<name>` (default: `~/.local/share/ade/plugins/<name>`) |
| Windows       | `%APPDATA%\ade\plugins\<name>`                                                     |

### GitHub release asset naming

For remote mode to work, the release must contain an asset whose filename includes the target OS and architecture:

```
my-plugin-linux-amd64
my-plugin-darwin-arm64
my-plugin-windows-amd64.exe
```

The OS and architecture strings must match Go's `runtime.GOOS` and `runtime.GOARCH` (case-insensitive). The first matching asset is downloaded.

### Authentication and rate limits

When the `GITHUB_TOKEN` environment variable is set, ADE forwards it as a `Bearer` token on every GitHub API and download request. This is required for private repositories. GitHub Actions sets this variable automatically.

For public repositories the variable is optional, but unauthenticated requests are subject to a rate limit of 60 per hour. If you hit the limit you will see a 403. Authenticating raises the limit to 5 000 per hour. A personal access token with no extra scopes is sufficient. The easiest way to obtain one is via the GitHub CLI:

```sh
export GITHUB_TOKEN=$(gh auth token)   # bash / zsh
$env:GITHUB_TOKEN = (gh auth token)    # PowerShell
```

## `ade plugin uninstall`

Remove the binary from the data directory and delete its entry from the global config:

```sh
ade plugin uninstall fscheck
```

## `ade plugin list`

Print every plugin registered in the global config with its binary path, status, installed version, and source:

```sh
ade plugin list
```

```
PLUGIN     PATH                                                  STATUS   VERSION   SOURCE
archgo     /home/user/.local/share/ade/plugins/archgo            ok       v0.2.0    github.com/phi42/ad-plugin-archgo
fscheck    /home/user/.local/share/ade/plugins/fscheck           ok                 (local)
my-plugin  /home/user/.local/share/ade/plugins/my-plugin         missing  v0.1.0    github.com/someone/my-plugin
```

- `STATUS` is `ok` if the binary exists at the registered path, `missing` otherwise.
- `VERSION` is the release tag recorded at install time. It is empty for locally installed plugins.
- `SOURCE` is the GitHub module URL for remote installs, or `(local)` for `--path` installs.

## `ade plugin update`

Re-fetch a GitHub release for a remotely installed plugin and overwrite the existing binary:

```sh
ade plugin update archgo
```

If the installed version already matches the latest, the download is skipped and a message is printed.

```sh
ade plugin update archgo --version v0.1.1   # pin a specific tag (downgrading allowed)
ade plugin update --all                     # update every remotely installed plugin
```

| Flag                | Description                                                       |
| ------------------- | ----------------------------------------------------------------- |
| `-v`, `--version`   | Release tag to fetch (e.g. `v0.1.1`); defaults to the latest tag. |
| `-a`, `--all`       | Update every remotely installed plugin to its latest release.     |

`--all` and `--version` cannot be combined. Plugins installed with `--path` (local mode) are ignored by `--all` and cannot be targeted by `update` individually because no remote source was recorded. Use `ade plugin install <name> --path <new-path>` to replace a locally installed plugin.

## `ade config`

Manage default values for command flags and plugin-specific configuration. By default the project config (`.ade.yaml` in the current directory) is targeted; pass `--global` to target the user-level config, or `--file <path>` to target a specific file.

```sh
ade config set   defaults.compile.plugin archgo
ade config get   defaults.compile.plugin
ade config unset defaults.compile.plugin
ade config list
```

| Flag             | Description                                                                  |
| ---------------- | ---------------------------------------------------------------------------- |
| `--global`       | Target the global config instead of the project config.                      |
| `--file <path>`  | Target a specific config file. Mutually exclusive with `--global`.           |

### Configurable keys

| Key                                  | Purpose                                                                                                                |
| ------------------------------------ | ---------------------------------------------------------------------------------------------------------------------- |
| `defaults.compile.plugin`            | Default value for `ade compile --plugin`.                                                                              |
| `defaults.compile.input`             | Default value for `ade compile --input`.                                                                               |
| `defaults.verify.plugin`             | Default value for `ade verify --plugin`.                                                                               |
| `defaults.verify.input`              | Default value for `ade verify --input`.                                                                                |
| `plugin_locations.<name>`            | Path of an installed plugin binary. Written automatically by `ade plugin install`; can also be edited by hand.         |
| `plugin_sources.<name>`              | GitHub module URL the plugin was installed from. Written automatically; consumed by `ade plugin update` and `list`.    |
| `plugin_versions.<name>`             | Release tag recorded at install time. Written automatically; displayed by `ade plugin list`.                           |
| `plugin_configs.<prefix>.<key>`      | Plugin-specific setting forwarded to the plugin in `rule.Spec.PluginConfig`. See [Plugin configuration](#plugin-configuration). |

`ade config set` rejects any key that is not in this list or under the `plugin_configs.` prefix.

### Configuration hierarchy

ADE uses [Viper](https://pkg.go.dev/github.com/spf13/viper) to merge two config locations on every run, so user-level defaults can be overridden per project:

1. Global (user-level) config, loaded first as the base:

   | Platform      | Path                                                                |
   | ------------- | ------------------------------------------------------------------- |
   | Linux / macOS | `$XDG_CONFIG_HOME/ade/ade.yaml` (default: `~/.config/ade/ade.yaml`) |
   | Windows       | `%APPDATA%\ade\ade.yaml`                                            |

2. Project-level config: `.ade.yaml` in the current working directory, merged on top of the global config. Values defined here override the global config.

Pass `--config <path>` on any command to bypass both files and use a specific config file instead:

```sh
ade compile --config ./my-config.yaml -p archgo -i ./rules
```

### Plugin configuration

Plugins read per-rule configuration from `plugin_configs.<prefix>.<key>`, where `<prefix>` matches the `config_prefix` the plugin advertises in its `--info` response. The entries are forwarded to the plugin as a `map<string, string>` in the `rule.Spec.PluginConfig` field.

For example, to set the output directory for the `archgo` plugin and the root directory for the `fscheck` plugin:

```sh
ade config set plugin_configs.archgo.output-dir   ./internal/archtests
ade config set plugin_configs.fscheck.root-dir    ./src
```

Resulting `.ade.yaml`:

```yaml
plugin_configs:
  archgo:
    output-dir: ./internal/archtests
  fscheck:
    root-dir: ./src
```

The exact key names a plugin understands are documented by the plugin itself.

### `ade config list`

Print every configurable default and every `plugin_configs.*` entry merged into the current view, with its effective value and source:

```
KEY                                          VALUE                       SOURCE
defaults.compile.plugin                      archgo                      [project]
defaults.compile.input                       ./docs/adr                  [global]
defaults.verify.plugin                                                   [not set]
defaults.verify.input                                                    [not set]
plugin_configs.archgo.output-dir             ./internal/archtests        [project]
plugin_configs.fscheck.root-dir              ./src                       [project]
```

The `SOURCE` column tags each value as `[project]`, `[global]`, or `[not set]`, making it explicit where a value comes from after the merge.

### Full configuration example

```yaml
plugin_locations:
  archgo:      /home/user/.local/share/ade/plugins/archgo
  netarchtest: /home/user/.local/share/ade/plugins/netarchtest
  fscheck:     /home/user/.local/share/ade/plugins/fscheck
plugin_sources:
  archgo:      github.com/phi42/ad-plugin-archgo
  netarchtest: github.com/phi42/ad-plugin-netarchtest
  fscheck:     github.com/phi42/ad-plugin-fscheck
plugin_versions:
  archgo:      v0.2.0
  netarchtest: v0.1.1
  fscheck:     v0.1.1
plugin_configs:
  archgo:
    output-dir: ./internal/archtests
  netarchtest:
    output-dir:        ./src/Tests/ArchTests/Generated
    test-project:      ./src/Tests/ArchTests/Project.csproj
    assembly-prefixes: CompanyName.MyApp.
  fscheck:
    root-dir: ./src
defaults:
  compile:
    plugin: archgo
    input:  ./docs/adr
  verify:
    plugin: fscheck
    input:  ./docs/adr
```
