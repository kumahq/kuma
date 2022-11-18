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
ARTIFACTS_DIR="build/artifacts-${GOOS}-${GOARCH}"
DESTINATION_DIR="${ARTIFACTS_DIR}/$1"

function safe_cp() {
  local files dest dir_1 final_dest

  files="$1"
	dest="$2"
	dir_1="$(dirname "${file}")"
	final_dest="${dest}/${dir_1}/"

	mkdir -p "${final_dest}" && for file in $files ; do cp -fR "${file}" "${final_dest}" ; done
}

export -f safe_cp

tail -n +2 "${DOCKERIGNORE_FILE}" | cut -d '!' -f 2 | while read -r file
do
	safe_cp "${file}" "${DESTINATION_DIR}"
done

find build -exec touch -mt 201212211111 {} \;
