#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

POLICY=$1

POLICIES_DIR=pkg/plugins/policies
POLICIES_CRD_DIR="${POLICIES_DIR}/${POLICY}/k8s/crd"

if [ "$(find "${POLICIES_CRD_DIR}" -type f | wc -l | xargs echo)" != 1 ]; then
  echo "More than 1 file in crd directory"
  exit 1
fi

CRD_FILE="$(find "${POLICIES_CRD_DIR}" -type f)"

HELM_CRD_DIR=deployments/charts/kuma/crds

cp "${CRD_FILE}" "${HELM_CRD_DIR}"
