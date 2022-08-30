#!/bin/bash

set -o pipefail
set -o nounset
set -x

HELM_VALUES_FILE="deployments/charts/kuma/values.yaml"
HELM_CRD_DIR="deployments/charts/kuma/crds/"

policies=""

for policy in "$@"; do

  policy_dir="pkg/plugins/policies/${policy}"
  policy_crd_dir="${policy_dir}/k8s/crd"

  if [ "$(find "${policy_crd_dir}" -type f | wc -l | xargs echo)" != 1 ]; then
    echo "More than 1 file in crd directory"
    exit 1
  fi

  policy_crd_file="$(find "${policy_crd_dir}" -type f)"
  rm -f "${HELM_CRD_DIR}/$(basename "${policy_crd_file}")"

  if [ ! -f "${policy_dir}/zz_generated.plugin.go" ]; then
    echo "Policy ${policy} has skip registration, not updating helm"
    continue
  fi

  cp "${policy_crd_file}" "${HELM_CRD_DIR}"

  plural=$(yq e '.spec.names.plural' "${policy_crd_file}")

  policies="${policies}${policies:+, }\"${plural}\""

done

# yq_patch preserves indentation and blank lines of the original file
function yq_patch() {
  yq '.' "$2" > "$2.noblank"
  yq eval "$1" "$2" | diff -B "$2.noblank" - | patch -f --no-backup-if-mismatch "$2" -
  rm "$2.noblank"
}

yq_patch '.plugins.policies = []' "${HELM_VALUES_FILE}"
yq_patch '.plugins.policies = [ '"${policies}"' ]' "${HELM_VALUES_FILE}"
