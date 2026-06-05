#!/usr/bin/env bash
# Pipe the pre-generated sample Spec into the plugin.
# Run from the go/ directory:  bash test-pipe.sh
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
go build -o "$DIR/plugin" "$DIR"
"$DIR/plugin" < "$DIR/testdata/sample.bin"
