#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

POLICY=$1
VERSION=${2:-"v1alpha1"}

POLICIES_DIR=pkg/plugins/policies
POLICIES_CRD_DIR="${POLICIES_DIR}/${POLICY}/k8s/crd"

if [ $(ls "${POLICIES_CRD_DIR}" | wc -l) != 1 ]; then
  echo "More than 1 file in crd directory"
  exit 1
fi

CRD_FILE="${POLICIES_CRD_DIR}/$(ls ${POLICIES_CRD_DIR})"

HELM_CRD_DIR=deployments/charts/kuma/crds

cp "${CRD_FILE}" "${HELM_CRD_DIR}"
