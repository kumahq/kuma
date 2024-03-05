#!/bin/bash

set -e

POLICIES_FILE="pkg/config/plugins/policies/policies.go"

policies=$(for i in "${@:1}"; do
  if [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]]; then
    policy_dir="pkg/plugins/policies/${i}"
    policy_crd_dir="${policy_dir}/k8s/crd"
    policy_crd_file="$(find "${policy_crd_dir}" -type f)"
    plural=$(yq e '.spec.names.plural' "$policy_crd_file")
    echo "\"$plural\","
  fi
done)

if [[ $policies == "" ]]; then
  rm -f "${POLICIES_FILE}"
  exit 0
fi

echo "package policies

import \"golang.org/x/exp/slices\"

var DefaultPluginPoliciesEnabled = []string{
  $policies
}

func DefaultPoliciesConfig() *Config {
	slices.Sort(DefaultPluginPoliciesEnabled)
	return &Config{
		PluginPoliciesEnabled: DefaultPluginPoliciesEnabled,
	}
}
" > "${POLICIES_FILE}"

gofmt -w "${POLICIES_FILE}"
