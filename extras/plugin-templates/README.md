# ADE plugin templates

Ready-to-use starting points for writing an ADE enforcement plugin in Go, C#, or Java.

## What is an ADE plugin?

Plugins extend the `ade compile` and `ade verify` commands with
language-specific behaviour.  They are plain executables that speak a small
protocol:

1. When called with `--info` they print a JSON object listing the modes they
   support (`compile`, `verify`, or both) and a `config_prefix` string that
   identifies their section in the `.rule` file's `plugin_config` map.
2. Otherwise they read a serialised `rule.Spec` protobuf message from stdin,
   do their work, and exit zero on success.

See [`docs/plugin-developer-guide.md`](../../docs/plugin-developer-guide.md) for
the full protocol specification.

## Templates

Each template lives in its own self-contained directory -- copy just the folder
you need, it already includes its own `testdata/` with a ready-to-use binary
fixture for standalone testing.

| Directory            | Language     | Build tool       | Entry point                 |
| -------------------- | ------------ | ---------------- | --------------------------- |
| [`go/`](go/)         | Go 1.22+     | `go build`       | `main.go`                   |
| [`csharp/`](csharp/) | C# / .NET 8+ | `dotnet publish` | `Program.cs`                |
| [`java/`](java/)     | Java 17+     | `gradle fatJar`  | `src/main/java/…/Main.java` |

Each template:

- Handles `--info` using a typed struct / class that marshals to JSON, returning
  `{"modes":["compile","verify"],"config_prefix":"my-plugin"}`.
- Reads and deserialises a `rule.Spec` protobuf from stdin.
- Prints a one-line summary to confirm the message was received.
- Includes `test-pipe.sh` (Unix/macOS) and `test-pipe.ps1` (Windows) scripts
  for standalone testing without running `ade`.

## Running the pipe test

Each template's test scripts pipe `testdata/sample.bin` into the plugin using
shell stdin redirection -- the only portable way to deliver raw binary bytes.
**Do not use PowerShell's `|` pipeline for binary data**; it corrupts the bytes.

**Unix / macOS** -- run from inside the template directory:
```sh
bash test-pipe.sh
```

**Windows** -- run from inside the template directory:
```powershell
.\test-pipe.ps1
```

Expected output (all languages):
```
received Spec: ADR [0001] "Use Layered Architecture" -- 2 selector(s), 2 rule(s), mode=MODE_VERIFY
```

## Regenerating sample.bin

`testdata/sample.bin` in each template was generated from `testdata/main.go`.  If
`rule.proto` changes, regenerate it and copy the result into each template:

```sh
cd testdata
go run main.go > sample.bin
# then copy sample.bin to each template's testdata/
```

## Next steps after choosing a template

1. Copy the template directory to your own repository.
2. Set `configPrefix` to the prefix your plugin reads from
   `plugin_config`, and remove any `modes` your plugin does not implement.
3. Replace the `printSpec` / `PrintSpec` call with your actual logic.
4. Rename the module / project / package to match your repository.
5. Run the pipe test script to confirm the plugin deserialises correctly.
6. Run `ade plugin install <your-plugin> --path <path-to-your-plugin>` and then `ade verify -i testdata/sample.rule -p <your-plugin>` for a full integration test.
