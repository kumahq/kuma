#!/bin/bash

imports=$(for i in "$@"; do [[ -f pkg/plugins/policies/${i}/plugin.go ]] && echo "_ \"github.com/kumahq/kuma/pkg/plugins/policies/${i}\""; done)
if [[ $imports == "" ]]; then
  rm -f pkg/plugins/policies/imports.go
  exit 0
fi

echo "package policies

import (
)
" > pkg/plugins/policies/imports.go
