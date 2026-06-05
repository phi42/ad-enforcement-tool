# ADE CLI Reference

This document is a complete reference for the `ade` commands. For an introduction, see the [README](../README.md).

## Global flags

These flags apply to all `ade` subcommands:

| Flag              | Description                                                                |
| ----------------- | -------------------------------------------------------------------------- |
| `--config <path>` | Load a specific config file, bypassing the default global + project merge. |
| `--debug`         | Enable DEBUG-level logging.                                                |
| `--no-warnings`   | Suppress WARN-level log lines.                                             |
| `--quiet`         | Suppress all output except errors.                                         |

## `ade validate`

Validate rule file syntax without compiling or executing any plugin.

```sh
ade validate -i ./rules/0001-domain-is-independent.rule
ade validate -i ./rules/     # validate every .rule file in a directory
```

| Flag            | Description                                                   |
| --------------- | ------------------------------------------------------------- |
| `-i`, `--input` | Path to a `.rule` file or a directory (required, repeatable). |

## `ade compile`

Compile rules into executable architecture tests using a plugin.

```sh
ade compile -i my-adr.rule -p archgo -o ./internal
```

| Flag             | Description                                                    |
| ---------------- | -------------------------------------------------------------- |
| `-i`, `--input`  | Path to a `.rule` file or a directory (required).              |
| `-p`, `--plugin` | Plugin name or path (see "Plugin resolution" below).           |
| `-o`, `--output` | Directory that will receive generated test files (must exist). |

`-p` and `-o` can be omitted when their config defaults are set; see [Configuration](#configuration) below.

### Plugin resolution

The `-p` / `--plugin` flag accepts either a name or a path:

- Name (e.g., `archgo`, `netarch`) is resolved in this order:
  1. `plugin_locations.<name>` entry in the global config.
  2. `plugin_locations.<name>` entry in the project-level `.ade.yaml`.
  3. Current working directory (fallback).

  On Windows, `.exe` is appended automatically if the name has no extension.

- Path (absolute or relative) is always used directly.

## `ade verify`

Execute rules directly against the target, without generating test code.

```sh
ade verify -i my-adr.rule -p fscheck
ade verify -i my-adr.rule -p fscheck -r ./src
```

| Flag             | Description                                                                      |
| ---------------- | -------------------------------------------------------------------------------- |
| `-i`, `--input`  | Path to a `.rule` file or a directory (required).                                |
| `-p`, `--plugin` | Plugin name or path.                                                             |
| `-r`, `--root`   | Root directory for resolving path patterns (default: current working directory). |

`-p` can be omitted when `defaults.verify.plugin` is set; see [Configuration](#configuration) below.

## `ade plugin install`

Register a plugin binary with ADE. Works in two modes.

Local mode: register a binary you already built or downloaded:

```sh
ade plugin install fscheck --path ./plugins/fscheck/fscheck
```

Remote mode: download from a GitHub release:

```sh
ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo
```

In remote mode the plugin name is taken from the `<name>` argument. The binary is placed in the platform data directory:

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

The tool matches `runtime.GOOS` and `runtime.GOARCH` (case-insensitive) against asset filenames and downloads the first match.

### Authentication

When the `GITHUB_TOKEN` environment variable is set, ADE sends it as a Bearer token on all GitHub API and download requests. Private repositories require a token; public ones do not. GitHub Actions sets this variable automatically. On a developer machine:

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

Print all plugins registered in the global config with their paths and a status indicator (`ok` or `missing`):

```sh
ade plugin list
```

```
PLUGIN     PATH                                                  STATUS   SOURCE
archgo     /home/user/.local/share/ade/plugins/archgo            ok       github.com/phi42/ad-plugin-archgo
fscheck    /home/user/.local/share/ade/plugins/fscheck           ok       github.com/phi42/ad-plugin-fscheck
my-plugin  /home/user/.local/share/ade/plugins/my-plugin         missing  github.com/someone/my-plugin
```

## `ade plugin update`

Re-fetch the latest GitHub release for a remotely installed plugin:

```sh
ade plugin update archgo
```

Plugins installed with `--path` (local mode) cannot be updated this way because no remote source was recorded. Use `ade plugin install <name> --path <new-path>` to replace a locally installed plugin.

## Configuration

Manage defaults for frequently used command flags. By default the project config (`.ade.yaml` in the current directory) is targeted; pass `--global` to target the user-level config instead.

```sh
ade config set   defaults.compile.plugin archgo
ade config get   defaults.compile.plugin
ade config unset defaults.compile.plugin
ade config list
```

ADE uses [Viper](https://pkg.go.dev/github.com/spf13/viper) to load configuration from YAML files, merging two locations on every run so that user-level defaults can be overridden by a project-specific file.

### File hierarchy

1. Global (user-level) config: loaded first as the base, applies to all projects on the machine.

   | Platform      | Path                                                                |
   | ------------- | ------------------------------------------------------------------- |
   | Linux / macOS | `$XDG_CONFIG_HOME/ade/ade.yaml` (default: `~/.config/ade/ade.yaml`) |
   | Windows       | `%APPDATA%\ade\ade.yaml`                                            |

2. Project-level config: `.ade.yaml` in the current working directory, merged on top of the global config. Values defined here override the global config.

Pass `--config <path>` to bypass both files and use a specific config file instead:

```sh
ade compile --config ./my-config.yaml -p netarch -i ./rules
```

### Plugin entries

Plugin paths go under the `plugin_locations:` key. `ade plugin install` writes these entries automatically, but you can also edit them by hand:

```yaml
plugin_locations:
  netarch: /home/user/.local/share/ade/plugins/netarch
  archgo:  /home/user/.local/share/ade/plugins/archgo
  fscheck: /home/user/.local/share/ade/plugins/fscheck
```

A bare plugin name on the command line (e.g., `-p netarch`) is resolved against these entries; see [Plugin resolution](#plugin-resolution) above.

ADE also records paths of remotely installed plugins under `plugin_sources.<name>` so that `ade plugin update` can re-fetch without the user having to remember the URL:

```yaml
plugin_sources:
  archgo: github.com/phi42/ad-plugin-archgo
```

### Defaults

Defaults for frequently used flags live under the `defaults:` key:

```yaml
plugin_locations:
  archgo:  /home/user/.local/share/ade/plugins/archgo
  fscheck: /home/user/.local/share/ade/plugins/fscheck
defaults:
  compile:
    plugin: archgo
    output: ./internal
  verify:
    plugin: fscheck
```

When a flag is configured as a default, it can be omitted on the command line. If a flag is omitted and no default is configured, the command exits with an error naming the missing flag and the config key to set.

#### Configurable keys

| Key                       | Flag replaced     | Command   |
| ------------------------- | ----------------- | --------- |
| `defaults.compile.plugin` | `--plugin` / `-p` | `compile` |
| `defaults.compile.output` | `--output` / `-o` | `compile` |
| `defaults.verify.plugin`  | `--plugin` / `-p` | `verify`  |

#### Managing defaults

```sh
ade config set   defaults.compile.plugin archgo                 # project-level
ade config set   defaults.verify.plugin  fscheck --global       # global-level
ade config get   defaults.compile.plugin                        # print effective value
ade config unset defaults.compile.output                        # remove value
ade config list                                                 # show all keys, values, and source
```

`ade config list` tags each value with its source (`[project]`, `[global]`, or `[not set]`) so it is clear where a value comes from after the merge.
