# Pipe the pre-generated sample Spec into the plugin.
# Run from the csharp\ directory:  .\test-pipe.ps1
$ErrorActionPreference = 'Stop'
$dir = $PSScriptRoot

dotnet clean -c Release --nologo -q
dotnet publish -c Release -o "$dir\publish" --nologo
cmd /c "`"$dir\publish\plugin.exe`" < `"$dir\testdata\sample.bin`""
