# Implementation

This document describes the internal architecture of `ad-enforcement-tool` (the `ade` binary). End users should start with the [README](../README.md), the [user guide](user-guide.md), the [DSL reference](../dsl/dsl-reference.md), or the [CLI reference](cli-reference.md).

## Overview

ADE is responsible for:

- parsing `.rule` files in the ADE DSL into a protobuf intermediate representation (`rule.Spec`);
- locating and invoking plugin binaries that turn the IR into either generated test code (`compile`) or direct verification results (`verify`);
- managing the on-disk inventory of installed plugins and the user's configuration defaults.

Plugins are out-of-process executables that speak a tiny protocol over stdin/stdout. ADE knows nothing about Go, .NET, or any other language: it only forwards the `rule.Spec` and surfaces the plugin's exit code.

## Package layout

```text
ad-enforcement-tool/
├── ade/                         `ade` binary entry point  
├── cmd/                         public command tree (also embedded by adg)
│   ├── root.go                  NewEnforceCommand, --config flag, viper init
│   ├── compile.go               `ade compile`
│   ├── verify.go                `ade verify`
│   ├── validate.go              `ade validate`
│   ├── config/                  `ade config` subcommand group
│   └── plugin/                  `ade plugin` subcommand group
├── dsl/                         public package: language reference only
├── internal/
│   ├── config/                  configuration: keys, file IO, runtime viper
│   ├── dsl/                     DSL parser (ANTLR bridge + semantic validation)
│   ├── parser/                  ANTLR-generated lexer/parser
│   ├── plugin/                  plugin process protocol + install lifecycle
│   └── runner/                  compile/verify command orchestration
├── rule/                        protobuf schema shared with plugins
│   ├── rule.proto
│   └── rule.pb.go               generated; do not edit
├── extras/                      developer tools (not part of the Go module)
│   ├── vscode/                  VS Code syntax-highlighting extension for .rule files
│   ├── plugin-templates/        starter templates for plugin authors (Go, C#, Java)
│   └── ci-templates/            CI workflow templates (GitHub Actions)
└── docs/                        documentation (this directory)
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
       ┌──────────────────────────────────┐
       │  runner.Mode.Run                 │
       │                                  │
       │  resolveInput (flag or config)   │
       │  resolvePlugin (flag or config,  │
       │    expand via plugin_locations)  │
       └────────────────┬─────────────────┘
                        │
                        ▼
       ┌──────────────────────────────────┐
       │  plugin.QueryInfo                │
       │                                  │
       │  exec PLUGIN --info              │
       │  parse {modes, config_prefix}    │
       │  check SupportsMode              │
       └────────────────┬─────────────────┘
                        │
                        ▼
       ┌──────────────────────────────────┐
       │  runner.CollectRuleFiles         │
       │                                  │
       │  single file or walk *.rule      │
       └────────────────┬─────────────────┘
                        │
               for each .rule file
                        │
                        ▼
       ┌──────────────────────────────────┐
       │  dsl.ParseFile                   │
       │                                  │
       │  ANTLR + semantic validation     │
       │  => rule.Spec                    │
       └────────────────┬─────────────────┘
                        │
                        ▼
       ┌──────────────────────────────────┐
       │  plugin.Run                      │
       │                                  │
       │  marshal Spec to proto           │
       │  exec PLUGIN, pipe to stdin      │
       │  forward stdout/stderr           │
       └──────────────────────────────────┘
```

## DSL parsing

`dsl.ParseDSL` is the single entry point for turning a `.rule` source string into a `rule.Spec`. It lives in `internal/dsl/` and runs four phases:

1. Custom block extraction (`extractCustomBlocks`). Every `custom "name" { ... }` block is pulled out of the source, because its body is free-form text the ANTLR lexer cannot tokenise. Each extracted block is replaced with the same number of newline characters so that ANTLR's reported line numbers stay accurate.
2. Lexing and parsing via the ANTLR-generated `internal/parser` package. Errors are collected through a custom `dslErrorListener`.
3. Tree walking via `irVisitor` ([internal/dsl/visitor.go](../internal/dsl/visitor.go)). The visitor dispatches on the verb-phrase context type and delegates to small focused methods (`applyDependOn`, `applyExist`, `applyAnnotated`, ...) that update the rule.
4. Semantic validation ([internal/dsl/validate.go](../internal/dsl/validate.go)): an `adr` block is required; selector references must resolve; file-level checks (`exist`, `contain`) cannot mix with code-level assertions in the same rule; rule names are unique across both regular and custom rules.

## Plugin protocol

A plugin is any executable that responds to two invocations:

1) `plugin --info`

Calling a plugin with the `--info` flag must return a JSON object on stdout with the following fields:

```json
{
  "modes": ["compile", "verify"],
  "config_prefix": "myplugin",
  "version": "1.2.3"
}
```

`modes` is required, `config_prefix` and `version` are optional but recommended. ADE uses this information to determine whether the plugin supports the requested mode and to forward user configuration.

2) `plugin`

Calling a plugin without `--info` means ADE is invoking it to process a `rule.Spec`. ADE marshals the protobuf to the plugin's stdin and expects either generated test files (compile mode) or verification results (verify mode) on stdout. The plugin must exit with code 0 for success or non-zero for failure.

This protocol lives in `internal/plugin/` and the wire types are in [rule/rule.proto](../rule/rule.proto). The full plugin contract is in [docs/plugin-developer-guide.md](plugin-developer-guide.md).

## Configuration model

`internal/config.Init` loads configuration into a single shared viper instance using a two-tier hierarchy that mirrors how most user-level CLI tools behave:

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

Alternatively, if `--config <path>` is passed, the global config is merged with the specified file instead of the default project config.

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
