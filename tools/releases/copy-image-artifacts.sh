#!/usr/bin/env bash

# this script copies artifacts from a .dockeringore file to a build folder

if [[ $# -ne 3 ]]; then
  usage
  exit 1
fi

function usage() {
  echo "Usage: $0 <component> <os> <arch>"
}

DOCKERIGNORE_FILE="tools/releases/dockerfiles/Dockerfile.$1.dockerignore"
GOOS="$2"
GOARCH="$3"
ARTIFACTS_DIR="build/artifacts-$GOOS-$GOARCH"
DESTINATION_DIR="$ARTIFACTS_DIR/$1"

function safe_cp() {
  FILE="$1"
  DIR_1=$(dirname "$FILE")
  DEST="$2"
  FINAL_DEST="$DEST/$DIR_1/"
  # shellcheck disable=SC2086
  mkdir -p "$FINAL_DEST" && cp -R $FILE "$FINAL_DEST"
}

export -f safe_cp

tail -n +2 "$DOCKERIGNORE_FILE" | cut -d '!' -f 2 | while read -r file
do
  safe_cp "$file" "$DESTINATION_DIR"
done

find build | xargs -I {} touch -mt 201212211111 {}
