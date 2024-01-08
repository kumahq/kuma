#!/bin/bash

set -o pipefail
set -o nounset
set -e

HELM_VALUES_FILE=$1
HELM_CRD_DIR=$2
VALUES_FILE_POLICY_PATH=$3

policies=""

for policy in "${@:4}"; do

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

  policies=${policies}$plural" "

done

# yq_patch preserves indentation and blank lines of the original file
cp "${HELM_VALUES_FILE}" "${HELM_VALUES_FILE}.noblank"
# shellcheck disable=SC2016
policies="${policies}" yq "${VALUES_FILE_POLICY_PATH}"' |= ((env(policies) | trim | split(" "))[] as $item ireduce ({}; .[$item] = {}))' "${HELM_VALUES_FILE}" | \
  diff --ignore-all-space --ignore-blank-lines "${HELM_VALUES_FILE}.noblank" - | \
  patch --force --no-backup-if-mismatch "${HELM_VALUES_FILE}" -
rm -f "${HELM_VALUES_FILE}.noblank"
