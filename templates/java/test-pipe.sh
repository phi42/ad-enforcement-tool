#!/usr/bin/env bash
# Pipe the pre-generated sample Spec into the plugin.
# Run from the java/ directory:  bash test-pipe.sh
set -euo pipefail
DIR="$(cd "$(dirname "$0")" && pwd)"
./gradlew fatJar -q
java -jar "$DIR/build/libs/plugin.jar" < "$DIR/testdata/sample.bin"
