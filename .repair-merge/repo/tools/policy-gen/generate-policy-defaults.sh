#!/usr/bin/env bash

set -e

policies=$(for i in "${@:3}"; do
  if [[ -f pkg/plugins/policies/${i}/zz_generated.plugin.go ]]; then
    policy_dir="pkg/plugins/policies/${i}"
    policy_crd_dir="${policy_dir}/k8s/crd"
    policy_crd_file="$(find "${policy_crd_dir}" -type f)"
    plural=$(yq e '.spec.names.plural' "$policy_crd_file")
    echo "\"$plural\","
  fi
done | sort | uniq)

KUMA_DEFAULTS_FILE="pkg/config/app/kuma-cp/kuma-cp.defaults.yaml"

policies_no_newlines=$(echo -n "$policies" | tr -d '\n')
# shellcheck disable=SC2001
policies_no_last_char=$(echo "$policies_no_newlines" | sed 's/.$//')
echo "$policies_no_last_char"

yq e -i ".policies.pluginPoliciesEnabled = [$policies_no_last_char]" $KUMA_DEFAULTS_FILE
