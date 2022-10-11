#!/bin/bash

set -x

PB_GO_FILE=$1

function os_agnostic_sed() {
  if [[ "$(go env GOOS)" == "darwin" ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

# add 'nullable' marker for all fields with type list
os_agnostic_sed '/.*\[\].*omitempty/ i\
	// +nullable
' "${PB_GO_FILE}"

# remove 'omitempty' tag for all fields with type list
os_agnostic_sed '/.*\[\]/s/,omitempty//g' "${PB_GO_FILE}"
