#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

IMPORTS_FILE="pkg/plugins/policies/imports.go"

imports=$(for i in "$@"; do [[ -f pkg/plugins/policies/${i}/plugin.go ]] && echo "_ \"github.com/kumahq/kuma/pkg/plugins/policies/${i}\""; done)
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
