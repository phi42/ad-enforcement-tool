# Plugin Development

This guide covers everything needed to write, test, and release an ADE enforcement plugin. Plugins are standalone executables that live in their own repositories and can be written in any language.

## Quick start

The fastest way to start is to copy one of the ready-made templates from [`templates/`](../templates/):

- [`templates/go/`](../templates/go/) for a Go plugin using the `rule` package directly
- [`templates/csharp/`](../templates/csharp/) for a .NET 8 C# plugin using `Google.Protobuf`
- [`templates/java/`](../templates/java/) for a Java 21 plugin using `protobuf-java` and Gradle

Each template is self-contained: copy the directory of your chosen language into a new repository and follow the template's `README.md` to run the first smoke test.

Each template also ships a pre-configured GitHub Actions release workflow in `.github/workflows/release.yml`. Push a `v*` tag and the workflow produces native executables for all major platforms and uploads them to a GitHub release. See [Releasing your plugin](#releasing-your-plugin) for details.

## Protocol

Any executable that responds correctly to two interactions qualifies as an ADE plugin.

### Info query

When invoked with `--info`, the plugin must print a JSON object to `stdout` and exit zero:

```json
{"modes": ["compile", "verify"], "config_prefix": "my-plugin"}
```

- `modes` lists the invocation modes the plugin supports. Declare only what you implement.
- `config_prefix` is the key prefix your plugin reads from `plugin_config` in the `Spec` (see [Config prefix](#config-prefix)). This field is optional but recommended.

ADE calls `--info` before each invocation to verify that the plugin supports the requested mode.

### Spec delivery

For every actual invocation (not `--info`), ADE:

1. Parses and validates the `.rule` file.
2. Builds a `Spec` protobuf message describing the rule file and invocation context.
3. Spawns the plugin as a child process and writes the serialised `Spec` bytes to the plugin's `stdin`.
4. Waits for the plugin to exit. A non-zero exit code signals failure.

The plugin reads `stdin`, deserialises the `Spec`, and acts on it.

### Spec structure

`Spec` is defined in [`rule/rule.proto`](../rule/rule.proto). The fields most useful to a plugin are:

| Field | Type | Description |
| --- | --- | --- |
| `adr` | `Adr` | The ADR header: `id` and `title` from the `.rule` file. |
| `selectors` | `[]Selector` | Named element or path references declared in the rule file. Each has a `name`, `pattern`, and `kind` (`component`, `class`, `interface`, or `path`). |
| `rules` | `[]Rule` | The enforcement rules. See the `Rule` message below. |
| `mode` | `InvocationMode` | `MODE_COMPILE` or `MODE_VERIFY`. |
| `plugin_config` | `map<string,string>` | Key/value pairs from `plugin_config` blocks in the rule file. Keys are prefixed with the plugin's config prefix. |
| `output_dir` | `string` | Target directory for generated files (compile mode only). |

Each `Rule` carries:

| Field | Type | Description |
| --- | --- | --- |
| `name` | `string` | Rule name from the DSL. |
| `kind` | `RuleKind` | Constraint type: `depend_only`, `not_depend`, `annotate`, `implement`, `in`, `match`, `visibility`, and others. |
| `from` | `TargetRef` | The subject side of the constraint (a selector name or inline pattern). |
| `targets` | `[]TargetRef` | The target side of the constraint. |
| `severity` | `Severity` | `WARNING` or `ERROR`. |
| `excludes` | `[]Exclusion` | Elements explicitly excluded from the rule. |
| `is_file_rule` | `bool` | True for file-system rules (`checks` will be populated). |
| `is_custom_rule` | `bool` | True for `custom` blocks (see [Custom rule blocks](#custom-rule-blocks)). |
| `raw_body` | `string` | Raw body text for custom rules. |

### Protobuf bindings

Go plugins import the rule package directly:

```go
import "github.com/phi42/ad-enforcement-tool/rule"
```

For C# and Java, the templates already include a copy of `rule.proto` and a build step that generates the bindings. For other languages, regenerate from [`rule/rule.proto`](../rule/rule.proto) using your language's `protoc` plugin.

## Next steps: implementing your logic

After the template confirms that the `Spec` arrives correctly, replace the `printSpec` call with real logic. The two modes work differently.

### Verify mode

In verify mode, the plugin inspects the actual codebase against the rules in the `Spec` and reports any violations to `stdout` using the format:

```
LEVEL  [rule-name] description of the violation
```

A typical verify implementation:

1. Iterates over `spec.Rules`, skipping any rules the plugin does not understand (file rules, custom rules not meant for it, or rule kinds outside its scope).
2. For each supported rule, resolves the `from` and `targets` fields to actual code elements by scanning the project using the `selectors` as a guide.
3. Checks whether the actual code structure satisfies the rule's constraint (`kind`).
4. Prints a `WARN` or `ERROR` line for each violation and exits non-zero if any `ERROR`-severity violation was found.

The `netarchtest` plugin is a good reference: it maps each rule into a NetArchTest predicate, runs the tests against the compiled assembly, and writes one result line per rule.

### Compile mode

In compile mode, the plugin generates artefacts (test files, configuration, documentation, and so on) and writes them to `spec.OutputDir`. A typical compile implementation:

1. Iterates over `spec.Rules`, collecting the rules it can generate code for.
2. Builds a data model from the selectors and rules.
3. Renders one or more output files using a template engine and writes them to `spec.OutputDir`.
4. Prints an `INFO` line naming each generated file.

The `netarchtest` plugin generates a C# test class containing one test method per architecture rule and writes it as `<AdrId>ArchTests.cs` in `OutputDir`.

### Config prefix

Plugins can read per-rule configuration from the `plugin_config` map. A rule file author writes:

```dsl
plugin_config {
  my-plugin.timeout = "30s"
  my-plugin.strict  = "true"
}
```

The plugin reads `spec.PluginConfig["my-plugin.timeout"]` at runtime. Set `config_prefix` in the `--info` response to the prefix your plugin uses so ADE can validate the keys.

### Custom rule blocks

A `custom` rule block lets plugin authors define entirely new assertions without modifying the grammar. The host stores the raw body text in the `raw_body` field of `Rule` and forwards it to the plugin unchanged:

```dsl
custom "my_check" {
  any text the plugin understands
  can go here with whatever syntax
  the plugin author defines
}
```

The `is_custom_rule` boolean on `Rule` marks these entries. Custom rules are forwarded to the plugin for both compile and verify invocations.

## Releasing your plugin

### GitHub Actions workflows

Each template includes a `.github/workflows/release.yml` that triggers on `v*` tags and produces platform-native executables:

- Go template: uses [GoReleaser](https://goreleaser.com) with a matching `.goreleaser.yml`, cross-compiling for Linux, macOS, and Windows on both `amd64` and `arm64`.
- C# template: uses `dotnet publish --self-contained` in a matrix across five platform/architecture combinations.
- Java template: uses [GraalVM native-image](https://www.graalvm.org/native-image/) to compile the fat JAR into a native executable for each platform.

### Asset naming

`ade plugin install --repo` scans the latest GitHub release for an asset whose filename contains both the current OS and architecture. Name assets following this pattern:

```
<repo>-<goos>-<goarch>          # Unix
<repo>-<goos>-<goarch>.exe      # Windows
```

The included workflows name assets this way automatically. The `<goos>` and `<goarch>` values must match Go's `runtime.GOOS` and `runtime.GOARCH` strings (for example, `linux`, `darwin`, `windows`, and `amd64`, `arm64`).

### Installing a released plugin

```sh
ade plugin install my-plugin --repo github.com/<owner>/<repo>
```
