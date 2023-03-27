#!/bin/bash

set -e

CI_TOOLS_BIN_DIR="$1"
CI_TOOLS_DIR="$2"
TOOLS_DEPS_DIRS="$3"
TOOLS_DEPS_LOCK_FILE="$4"
GOOS="$5"
GOARCH="$6"
TOOLS_MAKEFILE="$7"

mkdir -p "$CI_TOOLS_BIN_DIR" "$CI_TOOLS_DIR"/protos
# TOOLS_DEPS_DIRS has space separated directories
IFS=" " read -ra TOOLS_DEPS_DIRS <<< "${TOOLS_DEPS_DIRS[@]}"

# Also compute a hash to use for caching
FILES=$(find "${TOOLS_DEPS_DIRS[@]}" -name '*.sh' | sort)
for i in ${FILES}; do
  OS="$GOOS" ARCH="$GOARCH" "$i" "${CI_TOOLS_DIR}" &
done

wait
# use dev.mk to calculate the hash
FILES+=" "${TOOLS_MAKEFILE}
for i in ${FILES}; do cat "${i}"; done | git hash-object --stdin > "$TOOLS_DEPS_LOCK_FILE"

echo "All non code dependencies installed, if you use these tools outside of make add $CI_TOOLS_BIN_DIR to your PATH"
