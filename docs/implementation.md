# Implementation

This document describes the internal architecture of `ad-enforcement-tool` (the `ade` binary). It targets contributors who want to understand or modify the code. End users should start with the [README](../README.md), the [user guide](user-guide.md), the [DSL reference](../dsl/dsl-reference.md), or the [CLI reference](cli-reference.md).

## Overview

ADE is the orchestrator in a host-and-plugins architecture. The host (this repository) is responsible for:

- parsing `.rule` files in the ADE DSL into a protobuf intermediate representation (`rule.Spec`);
- locating and invoking plugin binaries that turn the IR into either generated test code (`compile`) or direct verification results (`verify`);
- managing the on-disk inventory of installed plugins and the user's configuration defaults.

Plugins are out-of-process executables that speak a tiny protocol over stdin/stdout. The host knows nothing about Go, .NET, or any other language: it only forwards the `rule.Spec` and surfaces the plugin's exit code.

```text
┌──────────────────────────────────────────────┐
│                  ade host                    │
│                                              │
│  .rule files ──► parse ──► IR (rule.Spec)    │
│                  resolve plugin & config     │
│                  invoke plugin (proto stdin) │
└──────────────────────────┬───────────────────┘
                           │
                           ▼
              plugin (separate binary)
          archgo / netarch / fscheck / ...
                           │
                           ▼
          generated tests / verify report
```

## Package layout

```text
ad-enforcement-tool/
├── ade/
│   └── main.go                  binary entry point
├── cmd/                         public command tree (also embedded by adg)
│   ├── root.go                  NewEnforceCommand, --config flag, viper init
│   ├── compile.go               `ade compile`
│   ├── verify.go                `ade verify`
│   ├── validate.go              `ade validate`
│   ├── config/                  `ade config` subcommand group
│   │   ├── config.go            New(): builds the parent + persistent flags
│   │   ├── set.go               `ade config set <key> <value>`
│   │   ├── get.go               `ade config get <key>`
│   │   ├── unset.go             `ade config unset <key>`
│   │   └── list.go              `ade config list`
│   └── plugin/                  `ade plugin` subcommand group
│       ├── plugin.go            New(): builds the parent
│       ├── install.go           `ade plugin install <name> --path|--repo`
│       ├── uninstall.go         `ade plugin uninstall <name>`
│       ├── update.go            `ade plugin update <name>`
│       └── list.go              `ade plugin list`
├── dsl/                         public package: language reference only
│   ├── reference.go             Reference (//go:embed dsl-reference.md), Validate
│   └── dsl-reference.md         the language reference (embedded as dsl.Reference)
├── internal/
│   ├── config/                  configuration: keys, file IO, runtime viper
│   │   ├── keys.go              YAML key constants and KnownDefaults
│   │   ├── viper.go             Viper(), Init(cfgFile), FileUsed()
│   │   ├── paths.go             xdgConfigHome, GlobalConfigPath
│   │   ├── scope.go             ResolveConfigPath, ResolveConfigScope
│   │   ├── validation.go        ValidateKey
│   │   └── file.go              ReadFile/WriteFile + SetKey/GetKey/UnsetKey
│   ├── dsl/                     DSL parser (ANTLR bridge + semantic validation)
│   │   ├── parse.go             ParseDSL, ParseFile, custom-block extraction
│   │   ├── visitor.go           ANTLR tree visitor (verb-phrase dispatch)
│   │   └── validate.go          semantic validation of the built IR
│   ├── parser/                  ANTLR-generated lexer/parser (do not edit)
│   │   └── ade_*.go
│   ├── plugin/                  plugin process protocol + install lifecycle
│   │   ├── resolve.go           ResolvePath
│   │   ├── info.go              Info, QueryInfo, SupportsMode
│   │   ├── run.go               Run (marshal IR + exec)
│   │   ├── paths.go             GlobalDir, BinaryName, xdgDataHome
│   │   ├── binary.go            CopyBinary, SetExecutable, NormaliseModuleURL
│   │   ├── github.go            FetchRelease (GitHub releases API)
│   │   └── registry.go          UpdateRegistry/RemoveFromRegistry/ReadRegistry
│   └── runner/                  compile/verify command orchestration
│       ├── runner.go            Mode.Run -- shared compile/verify pipeline
│       └── rules.go             CollectRuleFiles
├── rule/                        protobuf schema shared with plugins
│   ├── rule.proto
│   └── rule.pb.go               generated; do not edit
├── extras/                      non-Go developer tools (not part of the Go module)
│   ├── vscode/                  VS Code syntax-highlighting extension for .rule files
│   ├── plugin-templates/        starter templates for plugin authors (Go, C#, Java)
│   └── ci-templates/            CI workflow templates (GitHub Actions)
└── docs/                        human-facing documentation (this directory)
```

The `internal/` boundary is enforced by Go's import rules: only this module's packages can import anything under `internal/`. Anything that needs to be reusable by ADG, plugins, or third-party code lives outside `internal/`: [`cmd`](../cmd/), [`dsl`](../dsl/), and [`rule`](../rule/).

## Compile and verify pipeline

`ade compile` and `ade verify` share the same orchestration; only the protobuf invocation mode and the config keys for default values differ. Both commands route through `internal/runner.Mode.Run`:

```text
┌─────────────────────┐   ┌─────────────────────┐
│  ade compile        │   │  ade verify         │
│    -i FILE          │   │    -i FILE          │
│    -p PLUGIN        │   │    -p PLUGIN        │
└──────────┬──────────┘   └──────────┬──────────┘
           │                         │
           └────────────┬────────────┘
                        │
                        ▼
       ┌────────────────────────────────┐
       │  runner.Mode.Run               │
       │                                │
       │  resolveInput (flag or config) │
       │  resolvePlugin (flag or config,│
       │    expand via plugin_locations)│
       └────────────────┬───────────────┘
                        │
                        ▼
       ┌────────────────────────────────┐
       │  plugin.QueryInfo              │
       │                                │
       │  exec PLUGIN --info            │
       │  parse {modes, config_prefix}  │
       │  check SupportsMode            │
       └────────────────┬───────────────┘
                        │
                        ▼
       ┌────────────────────────────────┐
       │  runner.CollectRuleFiles       │
       │                                │
       │  single file or walk *.rule    │
       └────────────────┬───────────────┘
                        │
               for each .rule file
                        │
                        ▼
       ┌────────────────────────────────┐
       │  dsl.ParseFile                 │
       │                                │
       │  ANTLR + semantic validation   │
       │  => rule.Spec                  │
       └────────────────┬───────────────┘
                        │
                        ▼
       ┌────────────────────────────────┐
       │  plugin.Run                    │
       │                                │
       │  marshal Spec to proto         │
       │  exec PLUGIN, pipe to stdin    │
       │  forward stdout/stderr         │
       └────────────────────────────────┘
```

Failures bubble up as Go errors; cobra's `RunE` plumbing surfaces them as the command's exit code (and prints the error to stderr because `enforceCmd` sets `SilenceUsage: true`).

## DSL parsing

`dsl.ParseDSL` is the single entry point for turning a `.rule` source string into a `rule.Spec`. It lives in `internal/dsl/` and runs four phases:

1. Custom block extraction (`extractCustomBlocks`). Every `custom "name" { ... }` block is pulled out of the source, because its body is free-form text the ANTLR lexer cannot tokenise. Each extracted block is replaced with the same number of newline characters so that ANTLR's reported line numbers stay accurate.
2. Lexing and parsing via the ANTLR-generated `internal/parser` package. Errors are collected through a custom `dslErrorListener`.
3. Tree walking via `irVisitor` ([internal/dsl/visitor.go](../internal/dsl/visitor.go)). The visitor dispatches on the verb-phrase context type and delegates to small focused methods (`applyDependOn`, `applyExist`, `applyAnnotated`, ...) that update the rule.
4. Semantic validation ([internal/dsl/validate.go](../internal/dsl/validate.go)): an `adr` block is required; selector references must resolve; file-level checks (`exist`, `contain`) cannot mix with code-level assertions in the same rule; rule names are unique across both regular and custom rules.

The public `dsl` package exposes only [dsl/reference.go](../dsl/reference.go): the embedded DSL reference string and a `Validate` convenience wrapper that delegates to `internal/dsl.ParseDSL`.

### When to extend the parser

- Adding a new verb phrase (e.g. `must be exported`): add the rule to `internal/parser/ADE.g4`, regenerate the parser (see [docs/plugin-developer-guide.md](plugin-developer-guide.md)), then add a new `apply*` method in `internal/dsl/visitor.go` and dispatch to it from `visitVerbPhrase`. If the new phrase produces a new `rule.RuleKind`, list it in `codeRulesWithSubject` in `internal/dsl/validate.go` so its selector references are checked.
- Adding a new selector kind (e.g. `enum`): extend the `selectorType` rule in the grammar, add the new kind to `rule.proto`, and update `getSelectorKind` in `internal/dsl/visitor.go`.
- Adding new semantic checks: add them in `internal/dsl/validate.go`, ideally alongside the existing helpers (`validateAssertionShape`, `validateSelectorRefs`).

## Plugin protocol

A plugin is any executable that responds to two invocations:

`plugin --info` (host -> stdout)

```json
{
  "modes": ["compile", "verify"],
  "config_prefix": "myplugin"
}
```

`modes` lists which `enforce` subcommands the plugin supports. `config_prefix` is the second segment under `plugin_configs.` from which user configuration is forwarded.

`plugin` (host -> serialised `rule.Spec` on stdin)

The host marshals the rule.Spec protobuf and pipes it to the plugin's stdin. The plugin's stdout and stderr are forwarded directly to the host's. A non-zero exit code from the plugin causes the host to exit non-zero.

This protocol lives in `internal/plugin/` and the wire types are in [rule/rule.proto](../rule/rule.proto). The full plugin contract for authors is in [docs/plugin-developer-guide.md](plugin-developer-guide.md).

## Configuration model

`internal/config.Init` (called from `cmd/root.go`'s `PersistentPreRun`) loads configuration into a single shared viper instance using a two-tier hierarchy that mirrors how most user-level CLI tools behave:

```text
┌─────────────────────────────────────────────────────┐
│  1. Global config                                   │
│     Linux/macOS: $XDG_CONFIG_HOME/ade/ade.yaml      │
│     Windows:     %APPDATA%\ade\ade.yaml             │
│     (read first, becomes the base)                  │
└────────────────────────────┬────────────────────────┘
                             │ merged
                             ▼
┌─────────────────────────────────────────────────────┐
│  2. Project config                                  │
│     ./.ade.yaml (current working directory)         │
│     (values override the global config)             │
└─────────────────────────────────────────────────────┘
```

If `--config <path>` is passed, both default files are skipped and only that file is loaded.

The full set of recognised keys lives in [internal/config/keys.go](../internal/config/keys.go):

| Constant                       | YAML key                        | Purpose                                                      |
| ------------------------------ | ------------------------------- | ------------------------------------------------------------ |
| `config.DefaultCompilePlugin`  | `defaults.compile.plugin`       | Default plugin for `ade compile`                             |
| `config.DefaultCompileInput`   | `defaults.compile.input`        | Default `--input` for `ade compile`                          |
| `config.DefaultVerifyPlugin`   | `defaults.verify.plugin`        | Default plugin for `ade verify`                              |
| `config.DefaultVerifyInput`    | `defaults.verify.input`         | Default `--input` for `ade verify`                           |
| `config.PluginLocationsPrefix` | `plugin_locations.<name>`       | Path of an installed plugin binary                           |
| `config.PluginSourcesPrefix`   | `plugin_sources.<name>`         | GitHub module URL the plugin was installed from              |
| `config.PluginConfigsPrefix`   | `plugin_configs.<prefix>.<key>` | Plugin-specific config forwarded as `rule.Spec.PluginConfig` |

The `ade config` subcommands operate on these keys via `config.SetKey` / `GetKey` / `UnsetKey`, which read and rewrite the YAML file directly and so do not depend on the merged viper instance.

## Plugin lifecycle (install / update / uninstall / list)

`ade plugin install` runs in one of two modes:

- Local (`--path`): copy the binary at the given path into `plugin.GlobalDir()` and record `plugin_locations.<name>` in the global config.
- Remote (`--repo`): call the GitHub releases API, pick the asset whose filename contains the current GOOS and GOARCH, download it via the API endpoint (so the `Authorization: Bearer $GITHUB_TOKEN` header reaches only api.github.com, not the CDN), and record both `plugin_locations.<name>` and `plugin_sources.<name>`.

`ade plugin update` re-runs the remote download path against `plugin_sources.<name>`. Plugins installed locally cannot be updated this way: re-run `ade plugin install ... --path` to replace them.

`ade plugin uninstall` deletes the binary on disk and removes both YAML entries.

`ade plugin list` reads the global config, stats every recorded path, and prints a table with an `ok`/`missing` status column.

The split is by concern rather than by command: a single command typically calls into two or three of these files (`binary.go`, `github.go`, `registry.go`).

## Error handling

The CLI is structured around cobra's `RunE` callback, not the older `Run` callback. Every command returns a regular Go `error` and lets cobra surface it. The root command sets `SilenceUsage: true`, so a returned error prints just the error text and a non-zero exit code, without the usage banner.

Within commands, errors are wrapped with `fmt.Errorf("doing X: %w", err)` so that the final message always identifies which step failed. There is no `os.Exit` outside `ade/main.go`.

## Regenerating the ANTLR parser and the protobuf

These steps are out of scope for normal development but are documented in [docs/plugin-developer-guide.md](plugin-developer-guide.md). The generated files live under [internal/parser/](../internal/parser/) (ANTLR) and [rule/rule.pb.go](../rule/rule.pb.go) (protoc-gen-go). Do not hand-edit them; regenerate from `internal/parser/ADE.g4` and `rule/rule.proto` respectively.

## Plugin starter templates

Starter templates for authoring a new plugin in Go, C#, or Java live in [extras/plugin-templates/](../extras/plugin-templates/). Each subdirectory contains a minimal but runnable plugin skeleton and a `test-pipe` script that exercises the `--info` and stdin/stdout protocol locally without needing a full ADE installation.
