#!/bin/bash

set -e
GO_MODULE=$1

IMPORTS_FILE="pkg/plugins/policies/imports.go"

imports=$(for i in "${@:2}"; do [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]] && echo "\"${GO_MODULE}/pkg/plugins/policies/${i}\""; done)
if [[ $imports == "" ]]; then
  rm -f "${IMPORTS_FILE}"
  exit 0
fi

mappings=$(for i in "${@:2}"; do [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]] && echo "\"${i}\": ${i}.InitPlugin,"; done)

echo "package policies

import (
$imports
)

var nameToModule = map[string]func(bool){
  $mappings
}

func initAllPolicies() {
	for _, initializer := range nameToModule {
		initializer(true)
	}
}

func init() {
	// we read this manually otherwise we would have to wire in enabling plugins in every test
	rawEnabledPluginPolicies := os.Getenv(\"KUMA_PLUGIN_POLICIES_ENABLED\")
	if rawEnabledPluginPolicies == \"\" {
		initAllPolicies()
	} else {
		enabledPluginPolicies := strings.Split(rawEnabledPluginPolicies, \",\")
		for _, policy := range enabledPluginPolicies {
			initializer, ok := nameToModule[policy]
			if ok {
				initializer(true)
			}
		}
	}
}
" > "${IMPORTS_FILE}"

gofmt -w "${IMPORTS_FILE}"
