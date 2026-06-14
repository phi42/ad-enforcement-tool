# ADE

[![License: Apache 2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](./LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/phi42/ad-enforcement-tool.svg)](https://pkg.go.dev/github.com/phi42/ad-enforcement-tool)
[![Go Version](https://img.shields.io/github/go-mod/go-version/phi42/ad-enforcement-tool)](./go.mod)
[![Go Report Card](https://goreportcard.com/badge/github.com/phi42/ad-enforcement-tool)](https://goreportcard.com/report/github.com/phi42/ad-enforcement-tool)
[![Latest Release](https://img.shields.io/github/v/release/phi42/ad-enforcement-tool?sort=semver)](https://github.com/phi42/ad-enforcement-tool/releases)

ADE (Architectural Decision Enforcement) is a command-line tool for enforcing architectural decisions. Rules are written in a small DSL and stored in `.rule` files. ADE either compiles each rule into an executable architecture test or verifies it directly against the codebase, delegating language-specific work to plugins.

ADE is part of the [ADG](https://github.com/adr/ad-guidance-tool) ecosystem. The `adg enforce` command tree is powered by this module.

## Quick start

```bash
go install github.com/phi42/ad-enforcement-tool/ade@latest
ade plugin install fscheck --repo github.com/phi42/ad-plugin-fscheck
ade validate -i my-adr.rule
ade verify   -i my-adr.rule -p fscheck
```

A typical rule file:

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

For a step-by-step walkthrough (install, plugins, writing a rule, compiling, running in CI) see the [user guide](docs/user-guide.md).

## Plugins

Enforcement is delegated to plugins, separate executables that receive the parsed rule IR (`rule.Spec` protobuf) and either generate tests (`compile` mode) or perform checks directly (`verify` mode).

| Plugin                                                                    | Target    | Description                                                                                       |
| ------------------------------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------- |
| [`ad-plugin-archgo`](https://github.com/phi42/ad-plugin-archgo)           | Go        | Compiles `code` rules into [arch-go](https://github.com/arch-go/arch-go) tests.                   |
| [`ad-plugin-netarchtest`](https://github.com/phi42/ad-plugin-netarchtest) | .NET / C# | Compiles `code` rules into [NetArchTest](https://github.com/BenMorris/NetArchTest) + NUnit tests. |
| [`ad-plugin-fscheck`](https://github.com/phi42/ad-plugin-fscheck)         | Any       | Executes `file` rules directly against the filesystem.                                            |

To author your own plugin, copy one of the starter templates in [`extras/plugin-templates/`](extras/plugin-templates/) (Go, C#, or Java) and follow the [plugin developer guide](docs/plugin-developer-guide.md).

> If you have developed a plugin that you'd like to share, feel free to open a PR adding it to the list above.

## Import ADE as a go module

The [`cmd`](cmd/) package exposes the full enforcement command tree as a single `*cobra.Command`. Other tools can register it directly:

```go
import adecmd "github.com/phi42/ad-enforcement-tool/cmd"

rootCmd.AddCommand(adecmd.NewEnforceCommand())
```

This is how [ad-guidance-tool](https://github.com/adr/ad-guidance-tool) integrates the enforcement commands under `adg enforce`. The protobuf types used for plugin communication are also exposed as a public Go package at [`rule`](rule/), and the DSL reference is exposed by the [`dsl`](dsl/) package.

## Documentation

- [docs/user-guide.md](docs/user-guide.md): end-to-end walkthrough; install, plugins, writing a rule, compile/verify, and running in CI.
- [docs/cli-reference.md](docs/cli-reference.md): every `ade` command, flag, and configuration key.
- [dsl/dsl-reference.md](dsl/dsl-reference.md): DSL reference; every keyword, every verb phrase.
- [docs/plugin-developer-guide.md](docs/plugin-developer-guide.md): plugin protocol, protobuf schema, and how to author your own plugin.
- [docs/implementation.md](docs/implementation.md): internal architecture and package layout for contributors.
- [extras/ci-templates/](extras/ci-templates/): ready-to-use CI workflow templates.
- [extras/vscode/](extras/vscode/): VS Code syntax-highlighting extension for `.rule` files.

## License

Licensed under the [Apache License, Version 2.0](./LICENSE).
