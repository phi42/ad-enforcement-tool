# Java plugin template

Minimal ADE enforcement plugin skeleton written in Java (Gradle build).

## Prerequisites

- JDK 17 or later
- Gradle 8 or later (or use `./gradlew` after running `gradle wrapper` once)

## Build and run

```sh
gradle fatJar
java -jar build/libs/plugin.jar --info
# → {"modes":["compile","verify"],"config_prefix":"my-plugin"}
```

The `src/main/proto/rule.proto` file is a copy of the canonical proto from
`ad-enforcement-tool/rule/rule.proto`.  The `com.google.protobuf` Gradle plugin
generates Java classes automatically on `gradle build`.  If the proto changes,
replace the file and rebuild.

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

The scripts use `cmd /c` with stdin redirection (`<`) to preserve binary bytes --
PowerShell's pipeline would corrupt them.

## Integration test with `ade`

ADE expects a native executable.  Wrap the JAR in a shell script on Unix, or
use GraalVM `native-image` for a self-contained binary:

```sh
# Unix wrapper example
printf '#!/bin/sh\nexec java -jar "$(dirname "$0")/plugin.jar" "$@"\n' > plugin
chmod +x plugin
ade plugin install my-plugin --path ./plugin
ade verify -i testdata/sample.rule -p my-plugin
```

## Next steps

1. Update `INFO` in `Main.java`: set the `configPrefix` to your plugin's prefix
   and remove any `modes` your plugin does not implement.
2. Replace the `printSpec` call with your actual compile / verify logic.
3. Update the package to match your project structure:
   - Change `java_package` in `src/main/proto/rule.proto`.
   - Rename the `com/example/plugin/` source directory to match.
   - Update `Main-Class` and `mainClass` in `build.gradle.kts`.
   - Update the `package` declaration at the top of `Main.java`.
