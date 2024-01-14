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

func DefaultPoliciesConfig() *Config {
    return &Config{
        PluginPoliciesEnabled: []string{
            $policies
        },
    }
}
" > "${POLICIES_FILE}"

gofmt -w "${POLICIES_FILE}"

KUMA_DEFAULTS_FILE="pkg/config/app/kuma-cp/kuma-cp.defaults.yaml"

policies_no_newlines=$(echo -n "$policies" | tr -d '\n')
# shellcheck disable=SC2001
policies_no_last_char=$(echo "$policies_no_newlines" | sed 's/.$//')
echo "$policies_no_last_char"

yq e -i ".policies.pluginPoliciesEnabled = [$policies_no_last_char]" $KUMA_DEFAULTS_FILE
