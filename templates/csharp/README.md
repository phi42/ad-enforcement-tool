# C# plugin template

Minimal ADE enforcement plugin skeleton written in C# (.NET 8).

## Prerequisites

- .NET 8 SDK or later

## Build and run

```sh
dotnet build
dotnet run -- --info
# → {"modes":["compile","verify"],"config_prefix":"my-plugin"}
```

The `proto/rule.proto` file is a copy of the canonical proto from
`ad-enforcement-tool/rule/rule.proto`.  `Grpc.Tools` generates the C# classes
automatically on `dotnet build`.  If the proto changes, replace the file and
rebuild.

## Testing with the sample binary

`testdata/sample.bin` is a pre-generated protobuf `Spec` message matching
`testdata/sample.rule`.  Pipe it directly into the plugin to confirm the plugin
receives and deserialises it without running the full `ade` tool:

**Unix / macOS**
```sh
bash test-pipe.sh
```

**Windows (PowerShell)**
```powershell
.\test-pipe.ps1
```

The scripts use `cmd /c` with stdin redirection (`<`) to preserve binary bytes -- PowerShell's pipeline would corrupt them.

## Integration test with `ade`

```sh
ade plugin install my-plugin --path ./publish/plugin
ade verify -i testdata/sample.rule -p my-plugin
```

## Next steps

1. Update the `info` variable in `Program.cs`: set `ConfigPrefix` to your plugin's prefix
   and remove any `Modes` your plugin does not implement.
2. Replace the `PrintSpec` call with your actual compile / verify logic.
3. Update the namespace to match your project structure:
   - Change `csharp_namespace` in `proto/rule.proto`.
   - Update the `using` directive at the top of `Program.cs` to match.
   - Optionally update `RootNamespace` in `plugin.csproj`.
