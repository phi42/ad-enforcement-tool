# Pipe the pre-generated sample Spec into the plugin.
# Run from the java\ directory:  .\test-pipe.ps1
$ErrorActionPreference = 'Stop'
$dir = $PSScriptRoot

.\gradlew.bat fatJar -q
cmd /c "java -jar `"$dir\build\libs\plugin.jar`" < `"$dir\testdata\sample.bin`""
