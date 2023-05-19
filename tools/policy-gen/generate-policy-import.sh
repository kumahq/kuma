#!/bin/bash

set -e
GO_MODULE=$1

IMPORTS_FILE="pkg/plugins/policies/imports.go"

imports=$(for i in "${@:2}"; do [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]] && echo "_ \"${GO_MODULE}/pkg/plugins/policies/${i}\""; done)
if [[ $imports == "" ]]; then
  rm -f "${IMPORTS_FILE}"
  exit 0
fi

echo "package policies

import (
$imports
)
" > "${IMPORTS_FILE}"

gofmt -w "${IMPORTS_FILE}"
