#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

POLICY=$1

POLICY_DIR=pkg/plugins/policies/${POLICY}
POLICY_CRD_DIR="${POLICY_DIR}/k8s/crd"

if [ ! -f "${POLICY_DIR}/plugin.go" ]; then
  echo "Policy has skip registration, not copying crd to helm"
  exit 0
fi

if [ "$(find "${POLICY_CRD_DIR}" -type f | wc -l | xargs echo)" != 1 ]; then
  echo "More than 1 file in crd directory"
  exit 1
fi

CRD_FILE="$(find "${POLICY_CRD_DIR}" -type f)"

HELM_CRD_DIR=deployments/charts/kuma/crds

cp "${CRD_FILE}" "${HELM_CRD_DIR}"
