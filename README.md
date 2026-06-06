# ADE

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/phi42/ad-enforcement-tool.svg)](https://pkg.go.dev/github.com/phi42/ad-enforcement-tool)
[![Go Version](https://img.shields.io/github/go-mod/go-version/phi42/ad-enforcement-tool)](./go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/phi42/ad-enforcement-tool)](https://goreportcard.com/report/github.com/phi42/ad-enforcement-tool)
[![Latest Release](https://img.shields.io/github/v/release/phi42/ad-enforcement-tool?sort=semver)](https://github.com/phi42/ad-enforcement-tool/releases)

ADE (Architectural Decision Enforcement) is a command-line tool written in Go for enforcing architectural decisions. Rules are written in a small domain-specific language and stored in `.rule` files. ADE either compiles each rule into an executable architecture test or verifies it directly against the codebase, delegating language-specific work to plugins.

ADE is part of the [ADG](https://github.com/adr/ad-guidance-tool) ecosystem. The `adg enforce` command is powered by this module: ADG imports the [`cmd`](cmd/) package and registers its command tree under `adg enforce`.

## Installation

### Via Go

```bash
go install github.com/phi42/ad-enforcement-tool/ade@latest
```

The `ade` binary is placed in `$GOPATH/bin` (typically `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows). Make sure that directory is in your `PATH`.

### Via release archive

Precompiled executables for major operating systems are available on the [releases page](https://github.com/phi42/ad-enforcement-tool/releases).

## Usage

```text
ade validate    Validate rule file syntax
ade compile     Compile rules into architecture tests using a plugin
ade verify      Verify rules directly against the target using a plugin
ade plugin      Manage enforcement plugins (install, uninstall, list, update)
ade config      Manage configuration defaults
```

Run `ade --help` or `ade <command> --help` for command-level details, or read [docs/enforcement.md](docs/enforcement.md) for the full CLI reference.

### Writing a rule file

Rules are written in a small DSL and saved as `.rule` files alongside the ADRs they encode:

```dsl
adr "0001" "Use Clean Architecture"

component "Domain" = "MyApp.Domain"
component "Infra"  = "MyApp.Infrastructure"

code "domain_isolated" {
  Domain must not depend on Infra
  severity error
}

file "tests_exist" {
  path "tests/ArchTests/" must exist
  severity error
}
```

See [dsl/dsl.md](dsl/dsl.md) for the full DSL reference.

### Validating rule files

Check rule files for syntax errors without invoking any plugin:

```bash
ade validate -i my-adr.rule
ade validate -i rules/     # validate every .rule file in a directory
```

### Compiling into architecture tests

Compile rule files into executable tests using a language-specific plugin:

```bash
ade compile -i my-adr.rule -p archgo  -o ./internal      # Go (arch-go)
ade compile -i my-adr.rule -p netarch -o ./src/Tests     # .NET (NetArchTest)
```

The generated test file is compiled and run as part of the project's normal test pipeline (`go test`, `dotnet test`, ...).

### Verifying directly

`file` rules can be verified immediately against the filesystem without generating any test code:

```bash
ade verify -i my-adr.rule -p fscheck
```

## Plugins

Enforcement is delegated to plugins, separate executables that receive the parsed rule IR (`rule.Spec` protobuf) and either generate tests (`compile` mode) or perform checks directly (`verify` mode).

| Plugin                                                                    | Target    | Description                                                                                      |
| ------------------------------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------ |
| [`ad-plugin-archgo`](https://github.com/phi42/ad-plugin-archgo)           | Go        | Compiles `code` rules into [arch-go](https://github.com/arch-go/arch-go) tests                   |
| [`ad-plugin-netarchtest`](https://github.com/phi42/ad-plugin-netarchtest) | .NET / C# | Compiles `code` rules into [NetArchTest](https://github.com/BenMorris/NetArchTest) + NUnit tests |
| [`ad-plugin-fscheck`](https://github.com/phi42/ad-plugin-fscheck)         | Any       | Executes `file` rules directly against the filesystem                                            |

Install a plugin from a GitHub release:

```bash
ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo
```

Pin a specific release tag with `@version`:

```bash
ade plugin install archgo --repo github.com/phi42/ad-plugin-archgo@v0.1.1
```

Or register a locally built binary:

```bash
ade plugin install fscheck --path ./path/to/fscheck
```

The plugin protocol (stdin/stdout contract, `--info` JSON, logging conventions) is documented in [docs/plugin-development.md](docs/plugin-development.md). Starter templates for Go, C#, and Java plugins are in [extras/templates/](extras/templates/).

## Configuration

To avoid passing `-p` and `-i` on every invocation, set defaults:

```bash
ade config set defaults.compile.plugin archgo
ade config set defaults.compile.input  ./docs/adr
ade config set defaults.verify.plugin  fscheck
```

Defaults are persisted to `.ade.yaml` in the project directory or the global config (with `--global`). The full configuration model, including plugin-specific keys under `plugin_configs.<prefix>.<key>`, is documented in [docs/enforcement.md](docs/enforcement.md).

## VS Code extension

A syntax-highlighting extension for `.rule` files lives in [extras/vscode](extras/vscode/). Install from the [latest release](https://github.com/phi42/ad-enforcement-tool/releases):

```bash
code --install-extension ade-syntax.vsix
```

## Embedding the `cmd` package

The [`cmd`](cmd/) package exposes the full enforcement command tree as a single `*cobra.Command`. Other tools can register it directly:

```go
import adecmd "github.com/phi42/ad-enforcement-tool/cmd"

// Register as a subcommand (e.g. adg enforce ...)
rootCmd.AddCommand(adecmd.NewEnforceCommand())
```

This is how [ad-guidance-tool](https://github.com/adr/ad-guidance-tool) integrates the enforcement commands under `adg enforce`.

The protobuf types (`rule.Spec`, `rule.Rule`, ...) used for plugin communication are exposed as a public Go package:

```go
import "github.com/phi42/ad-enforcement-tool/rule"
```

The DSL reference and a `Validate` shorthand are exposed by the [`dsl`](dsl/) package, which embeds [dsl/dsl.md](dsl/dsl.md):

```go
import adedsl "github.com/phi42/ad-enforcement-tool/dsl"

fmt.Println(adedsl.Reference)              // language reference (markdown)
err := adedsl.Validate(ruleFileContents)   // syntax + semantic check
```

## Documentation

- [dsl/dsl.md](dsl/dsl.md): DSL reference (every keyword, every verb phrase). Embedded in the `dsl` package as `dsl.Reference`.
- [docs/enforcement.md](docs/enforcement.md): full CLI reference and configuration model.
- [docs/implementation.md](docs/implementation.md): architecture and package layout for contributors.
- [docs/plugin-development.md](docs/plugin-development.md): plugin protocol, protobuf schema, and how to author your own plugin.

## License

Licensed under the [Apache License, Version 2.0](./LICENSE).
