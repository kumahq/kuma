#!/usr/bin/env bash

# this script copies all artifacts using copy-image-artifacts.sh

GOOS=$1
GOARCH=$2

CURRENT_DIR="${BASH_SOURCE%/*}"

# shellcheck disable=SC2231
for dockerignore in $CURRENT_DIR/dockerfiles/*.dockerignore
do
  BASENAME=$(basename "$dockerignore")
  IFS='.' read -ra parts <<< "$BASENAME"
  COMPONENT="${parts[1]}"
  "$CURRENT_DIR"/copy-image-artifacts.sh "$CURRENT_DIR"/dockerfiles/Dockerfile."$COMPONENT".dockerignore build/artifacts-"$GOOS"-"$GOARCH"/"$COMPONENT"
done
