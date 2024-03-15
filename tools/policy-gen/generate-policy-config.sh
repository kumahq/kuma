#!/usr/bin/env bash

set -e

POLICIES_DIR=${1#"pkg/"} # plugins/policies or core/resources/apis
CONFIG_POLICIES_FILE="pkg/config/${POLICIES_DIR}/zz_generated.policies.go"

policies=$(for i in "${@:2}"; do
  if [[ -f "pkg/${POLICIES_DIR}/${i}/zz_generated.plugin.go" ]]; then
    policy_dir="pkg/${POLICIES_DIR}/${i}"
    policy_crd_dir="${policy_dir}/k8s/crd"
    policy_crd_file="$(find "${policy_crd_dir}" -type f)"
    plural=$(yq e '.spec.names.plural' "$policy_crd_file")
    echo "\"$plural\","
  fi
done)

IFS="/" read -ra policies_dir_components <<< "${POLICIES_DIR}"
len=${#policies_dir_components[@]}
package_name=${policies_dir_components[len-1]}

echo "// Generated by tools/policy-gen
// Run \"make generate\" to update this file.

package ${package_name}

var DefaultEnabled = []string{
  $policies
}

func Default() *Config {
	return &Config{
		Enabled: DefaultEnabled,
	}
}
" > "${CONFIG_POLICIES_FILE}"

gofmt -w "${CONFIG_POLICIES_FILE}"
