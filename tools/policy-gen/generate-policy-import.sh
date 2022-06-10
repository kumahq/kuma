#!/bin/bash

IMPORTS=""
for i in "$@"; do
  if [[ -f pkg/plugins/policies/${i}/plugin.go ]]; then
    IMPORTS="${IMPORTS}\t_ \"github.com/kumahq/kuma/pkg/plugins/policies/${i}\"\n"
  fi
done

rm -f pkg/plugins/policies/imports.go
if [[ ${IMPORTS} == "" ]]; then
  exit 0
fi
echo "package policies

import (
${IMPORTS}
)
" > pkg/plugins/policies/imports.go
