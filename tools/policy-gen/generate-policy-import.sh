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

type pluginInitializer struct {
	initFn      func()
	initialized bool
}

var nameToModule = map[string]*pluginInitializer{
  $mappings
}

func InitAllPolicies() {
	for _, initializer := range nameToModule {
		if !initializer.initialized {
			initializer.initFn()
			initializer.initialized = true
		}
	}
}

func InitPolicies(enabledPluginPolicies []string) {
	for _, policy := range enabledPluginPolicies {
		initializer, ok := nameToModule[policy]
		if ok && !initializer.initialized {
			initializer.initFn()
			initializer.initialized = true
		} else {
			panic(\"policy \" + policy + \" not found\")
		}
	}
}
" > "${IMPORTS_FILE}"

gofmt -w "${IMPORTS_FILE}"
