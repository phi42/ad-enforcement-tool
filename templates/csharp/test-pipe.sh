#!/usr/bin/env bash
# Pipe the pre-generated sample Spec into the plugin.
# Run from the csharp/ directory:  bash test-pipe.sh
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
dotnet clean -c Release --nologo -q
dotnet publish -c Release -o "$DIR/publish" --nologo
"$DIR/publish/plugin" < "$DIR/testdata/sample.bin"
