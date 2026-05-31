# ADE

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)

ADE (Architectural Decision Enforcement) is a command-line tool written in Go for enforcing architectural decisions. Rules are written in a domain-specific language (DSL) and stored in `.rule` files. The tool compiles them into executable architecture tests or verifies them directly against a codebase using language-specific plugins.

ADE is part of the [ADG](https://github.com/adr/ad-guidance-tool) ecosystem. The `adg enforce` command is powered by this module.

## Getting started

### Installing via Go

```bash
go install github.com/phi42/ad-enforcement-tool/ade@latest
```

This places the `ade` binary in your `$GOPATH/bin` directory (typically `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows). Make sure this directory is in your `PATH`.

### Downloading a release

Precompiled executables for major operating systems are available on the [releases page](https://github.com/phi42/ad-enforcement-tool/releases).

## Commands

```
ade validate    Validate rule file syntax
ade compile     Compile rules into architecture tests using a plugin
ade verify      Verify rules directly against the target using a plugin
ade plugin      Manage enforcement plugins (install, uninstall, list, update)
ade config      Manage configuration defaults
```

Run `ade --help` or `ade <command> --help` for details on any command.

## Writing a rule file

Rules are written in a small DSL and saved as `.rule` files alongside your ADRs:

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

See [docs/dsl.md](docs/dsl.md) for the full DSL reference.

## Validating rule files

Check rule files for syntax errors without running any plugin:

```bash
ade validate -i my-adr.rule
ade validate -i rules/     # validate all .rule files in a directory
```

## Compiling into architecture tests

Compile rule files into executable tests using a language-specific plugin:

```bash
ade compile -i my-adr.rule -p arch-go -o ./internal      # Go (arch-go)
ade compile -i my-adr.rule -p netarch -o ./src/Tests      # .NET (NetArchTest)
```

The generated test file is compiled and run as part of your normal test pipeline (`go test`, `dotnet test`).

## Verifying directly

`file` rules can be verified immediately against the filesystem without generating any test code:

```bash
ade verify -i my-adr.rule -p filecheck
```

## Plugins

Enforcement relies on plugins, which are separate executables that receive the parsed rule IR and generate tests or perform checks.

| Plugin | Target | Description |
| ------ | ------ | ----------- |
| [`adplugin-arch-go`](https://github.com/phi42/adplugin-arch-go) | Go | Compiles `code` rules into [arch-go](https://github.com/arch-go/arch-go) tests |
| [`adplugin-netarchtest`](https://github.com/phi42/adplugin-netarchtest) | .NET / C# | Compiles `code` rules into [NetArchTest](https://github.com/BenMorris/NetArchTest) + NUnit tests |
| [`adplugin-fscheck`](https://github.com/phi42/adplugin-fscheck) | Any | Executes `file` rules directly against the filesystem |

Install a plugin from a GitHub release:

```bash
ade plugin install arch-go --repo github.com/phi42/adplugin-arch-go
```

Or register a locally built binary:

```bash
ade plugin install filecheck --path ./path/to/filecheck
```

## Configuration defaults

To avoid passing `-p` and `-o` on every run:

```bash
ade config set defaults.compile.plugin arch-go
ade config set defaults.compile.output ./internal
ade config set defaults.verify.plugin  filecheck
```

Defaults are stored in `.ade.yaml` in the project directory or in the global config. See [docs/enforcement.md](docs/enforcement.md) for the full command reference and configuration details.

## VS Code extension

A syntax-highlighting extension for `.rule` files is available in [editor/vscode](editor/vscode/). Install from the [latest release](https://github.com/phi42/ad-enforcement-tool/releases):

```bash
code --install-extension ade-syntax.vsix
```

## Importing the `enforce` package

The `enforce` package exposes the full enforcement command tree as a single `*cobra.Command`. Other tools can embed it directly:

```go
import adecmd "github.com/phi42/ad-enforcement-tool/enforce"

// Add as a subcommand (e.g., adg enforce ...)
rootCmd.AddCommand(adecmd.NewEnforceCommand())
```

This is how `adg enforce` is implemented — ADG imports this module and registers the command at `adg enforce`.

## Plugin development

See [docs/plugin-development.md](docs/plugin-development.md) for the plugin protocol, the protobuf schema, logging conventions, and instructions for regenerating the parser and protobuf code.

The protobuf types (`SpecIR`, `RuleIR`) used for plugin communication are available as a public Go package:

```go
import "github.com/phi42/ad-enforcement-tool/rule"
```

## License

ADE is released under the [Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0).
