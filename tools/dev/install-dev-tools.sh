#!/bin/bash

set -e

CI_TOOLS_BIN_DIR="$1"
CI_TOOLS_DIR="$2"
TOOLS_DEPS_DIRS="$3"
TOOLS_DEPS_LOCK_FILE="$4"
GOOS="$5"
GOARCH="$6"
# We cannot provide list of arguments not as a string, so we join them with a comma
TOOLS_DEPS_DIRS=${TOOLS_DEPS_DIRS//,/ }

mkdir -p "$CI_TOOLS_BIN_DIR" "$CI_TOOLS_DIR"/protos

# Also compute a hash to use for caching
FILES=$(find $TOOLS_DEPS_DIRS -name '*.sh' | sort)
for i in ${FILES}; do OS="$GOOS" ARCH="$GOARCH" "$i" "${CI_TOOLS_DIR}"; done
for i in ${FILES}; do echo "---${i}"; cat "${i}"; done | git hash-object --stdin > "$TOOLS_DEPS_LOCK_FILE"

echo "All non code dependencies installed, if you use these tools outside of make add $CI_TOOLS_BIN_DIR to your PATH"
