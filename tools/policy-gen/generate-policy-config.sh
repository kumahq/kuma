#!/bin/bash

set -e

POLICIES_FILE="pkg/config/plugins/policies/policies.go"

policies=$(for i in "${@:1}"; do [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]] && echo "\"${i}\","; done)
if [[ $policies == "" ]]; then
  rm -f "${POLICIES_FILE}"
  exit 0
fi

echo "package policies

func DefaultPoliciesConfig() *PoliciesConfig {
    return &PoliciesConfig{
        PluginPoliciesEnabled: []string{
            $policies
        },
    }
}
" > "${POLICIES_FILE}"

gofmt -w "${POLICIES_FILE}"
