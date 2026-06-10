# Pipe the pre-generated sample Spec into the plugin.
# Run from the go\ directory:  .\test-pipe.ps1
$ErrorActionPreference = 'Stop'
$dir = $PSScriptRoot

go build -o "$dir\plugin.exe" $dir
cmd /c "`"$dir\plugin.exe`" < `"$dir\testdata\sample.bin`""
