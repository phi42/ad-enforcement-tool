# Go plugin template

Minimal ADE enforcement plugin skeleton written in Go.

## Prerequisites

- Go 1.22 or later

## Build and run

```sh
go mod tidy
go build -o plugin .
./plugin --info
# → {"modes":["compile","verify"],"config_prefix":"my-plugin"}
```

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

The scripts use `cmd /c` with stdin redirection (`<`) to preserve binary bytes —
PowerShell's pipeline would corrupt them.

## Integration test with `ade`

```sh
go build -o plugin .
ade plugin install my-plugin --path ./plugin
ade verify -i testdata/sample.rule -p my-plugin
```

## Next steps

1. Update the `info` variable in `main.go`: set `ConfigPrefix` to your plugin's prefix
   and remove any `Modes` your plugin does not implement.
2. Replace the `printSpec` call with your actual compile / verify logic.
3. Rename the module in `go.mod` to match your repository.
4. Once `ad-enforcement-tool` has a published release, remove the `replace` directive
   in `go.mod` and pin the real version.
