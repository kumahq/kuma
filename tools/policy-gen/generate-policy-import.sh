#!/bin/bash

set -e
GO_MODULE=$1

IMPORTS_FILE="pkg/plugins/policies/imports.go"

imports=$(for i in "${@:2}"; do [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]] && echo "\"${GO_MODULE}/pkg/plugins/policies/${i}\""; done)
if [[ $imports == "" ]]; then
  rm -f "${IMPORTS_FILE}"
  exit 0
fi

mappings=$(for i in "${@:2}"; do
  if [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]]; then
    policy_dir="pkg/plugins/policies/${i}"
    policy_crd_dir="${policy_dir}/k8s/crd"
    policy_crd_file="$(find "${policy_crd_dir}" -type f)"
    plural=$(yq e '.spec.names.plural' "$policy_crd_file")
    echo "\"$plural\": {initFn: ${i}.InitPlugin, initialized: false},"
  fi
done)

echo "package policies

import (
  $imports
)

type PluginInitializer struct {
	InitFn      func()
	Initialized bool
}

var NameToModule = map[string]*pluginInitializer{
  $mappings
}

func InitAllPolicies() {
	for _, initializer := range NameToModule {
		if !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		}
	}
}

func InitPolicies(enabledPluginPolicies []string) {
	for _, policy := range enabledPluginPolicies {
		initializer, ok := NameToModule[policy]
		if ok && !initializer.Initialized {
			initializer.InitFn()
			initializer.Initialized = true
		} else {
			panic(\"policy \" + policy + \" not found\")
		}
	}
}
" > "${IMPORTS_FILE}"

gofmt -w "${IMPORTS_FILE}"
