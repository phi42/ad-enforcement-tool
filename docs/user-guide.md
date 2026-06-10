# ADE User Guide

This guide walks through using ADE end-to-end: installing the `ade` binary, installing plugins, writing and validating a rule, compiling or verifying it, and running everything in CI on every push.

If you are looking for a short reference of every flag, see the [CLI reference](cli-reference.md). For the DSL grammar, see the [DSL reference](../dsl/dsl-reference.md). To author your own plugin, see the [plugin developer guide](plugin-developer-guide.md).

## Step 1: Install ade

You can install `ade` either through the Go toolchain or from a release archive.

**Via `go install`** (requires Go 1.22+):

```bash
go install github.com/phi42/ad-enforcement-tool/ade@latest
```

The `ade` binary is placed in `$GOPATH/bin` (typically `~/go/bin` on Linux/macOS or `%USERPROFILE%\go\bin` on Windows). Make sure that directory is in your `PATH`, then verify:

```bash
ade --help
```

**Via release archive.** Precompiled binaries for Linux, macOS, and Windows on both `amd64` and `arm64` are available on the [releases page](https://github.com/phi42/ad-enforcement-tool/releases). Download the asset for your OS and architecture, make it executable, and place it on your `PATH`.

### Optional: install the VS Code extension

A syntax-highlighting extension for `.rule` files is shipped with every release. Install it from the latest release:

```bash
code --install-extension ade.vsix
```

## Step 2: Install one or more plugins

ADE itself only parses and validates rule files. The actual enforcement is done by plugins, separate executables tailored to a specific language or check. Install the plugins you need before writing rules that depend on them.

| Plugin                                                                    | Target    | Description                                                                                       |
| ------------------------------------------------------------------------- | --------- | ------------------------------------------------------------------------------------------------- |
| [`ad-plugin-archgo`](https://github.com/phi42/ad-plugin-archgo)           | Go        | Compiles `code` rules into [arch-go](https://github.com/arch-go/arch-go) tests.                   |
| [`ad-plugin-netarchtest`](https://github.com/phi42/ad-plugin-netarchtest) | .NET / C# | Compiles `code` rules into [NetArchTest](https://github.com/BenMorris/NetArchTest) + NUnit tests. |
| [`ad-plugin-fscheck`](https://github.com/phi42/ad-plugin-fscheck)         | Any       | Executes `file` rules directly against the filesystem.                                            |

Install a plugin from its GitHub release:

```bash
ade plugin install fscheck --repo github.com/phi42/ad-plugin-fscheck
```

Pin a specific tag for reproducible installs (recommended in CI):

```bash
ade plugin install fscheck --repo github.com/phi42/ad-plugin-fscheck@v0.1.1
```

If you built a plugin locally, register the binary directly:

```bash
ade plugin install my-plugin --path ./bin/my-plugin
```

Confirm the install:

```bash
ade plugin list
```

```
PLUGIN     PATH                                                  STATUS   VERSION   SOURCE
fscheck    /home/user/.local/share/ade/plugins/fscheck           ok       v0.1.1    github.com/phi42/ad-plugin-fscheck
```

Plugins installed in remote mode can later be re-fetched with `ade plugin update <name>` (or `ade plugin update --all`). Locally installed plugins must be replaced with another `ade plugin install ... --path`.

### Optional: configure plugin-specific settings

Most plugins accept settings (e.g. an output directory, a project file, a root directory) through `plugin_configs.<prefix>.<key>` in the active config file. The `<prefix>` is the `config_prefix` the plugin advertises in its `--info` response, typically the plugin's short name.

For example, to point the `archgo` plugin at a generated-tests directory and the `fscheck` plugin at the source root:

```bash
ade config set plugin_configs.archgo.output-dir ./internal/archtests
ade config set plugin_configs.fscheck.root-dir  ./src
```

These values are written to `.ade.yaml` in the current directory by default; pass `--global` to write them to the user-level config instead. See [Step 4](#step-4-set-defaults-optional) and the [CLI reference](cli-reference.md#plugin-configuration) for the full configuration model.

## Step 3: Write and validate a rule

ADE rules live in `.rule` files. Each file encodes one or more checks derived from a single ADR. Create one alongside your ADRs (for example, in `docs/adr/`):

`docs/adr/0001-clean-architecture.rule`:

```dsl
adr "0001" "Use Clean Architecture"

component "Domain"         = "MyApp.Domain"
component "Application"    = "MyApp.Application"
component "Infrastructure" = "MyApp.Infrastructure"

code "domain_isolated" {
  Domain must not depend on Application, Infrastructure
  severity error
}

code "application_inward" {
  Application must only depend on Domain
  severity error
}

file "adr_exists" {
  path "docs/adr/0001-*.md" must exist
  severity error
}
```

Validate the syntax and semantics of the file without running any plugin:

```bash
ade validate -i docs/adr/0001-clean-architecture.rule
```

Pass a directory to validate every `.rule` file under it:

```bash
ade validate -i docs/adr/
```

For the full DSL grammar (selectors, exclusions, naming, visibility, annotations, custom blocks, and so on) see the [DSL reference](../dsl/dsl-reference.md).

## Step 4: Set defaults (optional)

To avoid passing `-p` and `-i` on every invocation, save them as defaults in `.ade.yaml`:

```bash
ade config set defaults.compile.plugin archgo
ade config set defaults.compile.input  ./docs/adr
ade config set defaults.verify.plugin  fscheck
ade config set defaults.verify.input   ./docs/adr
```

Project-level values land in `.ade.yaml` next to your code. Add `--global` to write to the user-level config instead. Use `ade config list` to see every configured key, its effective value, and where it came from.

The full configuration model and a worked example are in the [CLI reference](cli-reference.md#ade-config).

## Step 5: Enforce the rules locally

There are two enforcement modes, and you typically use both depending on what each rule checks.

### Compile mode

`ade compile` asks a plugin to translate `code` rules into executable test code (for example, an arch-go test or a NetArchTest test class) and writes that code into your project. You then run the generated tests with your normal test runner.

```bash
ade compile -i docs/adr -p archgo
go test ./internal/archtests/...
```

The output directory is read from `plugin_configs.<prefix>.output-dir` (or whatever key the plugin documents). Generated test files are typically named after the ADR ID so that re-running `ade compile` overwrites them deterministically.

Use compile mode when:

- the plugin targets a language with mature architecture-testing libraries (Go, .NET, Java);
- you want the rules to surface in your existing test reports;
- you want IDE navigation and refactoring to keep the tests in sync with the production code.

### Verify mode

`ade verify` asks a plugin to evaluate the rules directly and report violations on the spot. Nothing is written to disk.

```bash
ade verify -i docs/adr -p fscheck
```

Verify mode is the right fit for:

- file-system checks (existence, content, regex matches) that do not need a compiler;
- custom plugins that perform a specialised check;
- quick local feedback while iterating on a rule.

`ade verify` exits non-zero when any `error`-severity rule is violated, which makes it suitable for use in pre-commit hooks and CI without further plumbing.

## Step 6: Run ade in CI

Once the rule file is committed and runs locally, wire `ade` into your pipeline so that every push and pull request is checked.

### GitHub Actions

A ready-to-use workflow is available at [`extras/ci-templates/github-actions/enforce.yml`](../extras/ci-templates/github-actions/enforce.yml). Copy it to `.github/workflows/enforce.yml` in your repository and adapt the placeholders marked with `TODO`.

### Considerations

- **Install `ade` with `go install`.** The Go toolchain handles platform differences automatically (no per-OS curl URL, no permissions juggling), and GitHub Actions runners already have `actions/setup-go` available. Pin a release tag (`@vX.Y.Z`) rather than `@latest` so the build is reproducible.
- **Pin plugin versions.** Pin every `--repo ...@vX.Y.Z` so that the same revision of every component runs every time. An unpinned plugin can introduce or remove violations without any change to your rules.
- **Provide `GITHUB_TOKEN`.** ADE forwards `GITHUB_TOKEN` as a `Bearer` token on every GitHub API and download request. Without it, downloads are throttled to 60 per hour. With it, the limit is 5 000 per hour. GitHub Actions injects the variable automatically through `secrets.GITHUB_TOKEN`; expose it via `env:` as shown above.
- **Run `verify` and `compile` together when appropriate.** `verify` returns its own exit code, so it can fail the build directly. `compile` only writes files; the build fails only when the subsequent test runner picks up the generated tests and reports a failure. If you forget to run the test step, compile-mode rules are not enforced.
- **Commit the rule files, not the generated tests.** `compile` rewrites the generated files on every run, so checking them in adds noise. If you want the generated tests in the repository (for IDE support or for builds that do not invoke `ade`), regenerate them in a separate workflow and commit the result, or set up a git hook to refuse stale generated files.
- **Use a project-level `.ade.yaml` for defaults that travel with the repo.** Plugin paths, however, should never be committed: `ade plugin install` writes them with absolute filesystem paths to the developer's data directory. Commit `defaults.*` and `plugin_configs.*` keys; do not commit `plugin_locations.*`, `plugin_sources.*`, or `plugin_versions.*`.
- **Cache the plugin binaries** with `actions/cache` keyed on the plugin tags if pull-request volume makes the install step a noticeable share of the build time. The binaries land under `$XDG_DATA_HOME/ade/plugins/` (or `~/.local/share/ade/plugins/` on the default Ubuntu runner).
- **Use `--config <path>` for matrix builds.** If you have multiple projects in one repository, each with its own rules and plugin config, point each matrix job at its own config file with `ade --config ./module-a/.ade.yaml ...` rather than depending on the working directory.

### Other CI providers

ADE is a single static binary with no runtime dependencies, so any CI provider that can download a binary and run shell commands can host it. The steps are always the same: download the binary, install plugins with `ade plugin install --repo`, then call `ade verify` or `ade compile` followed by your test runner.

## More documentation

- [CLI reference](cli-reference.md): every flag and subcommand, plus the full configuration model.
- [DSL reference](../dsl/dsl-reference.md): every keyword and verb phrase the rule parser understands.
- [Plugin developer guide](plugin-developer-guide.md): the plugin protocol, the protobuf schema, and how to author your own plugin.
