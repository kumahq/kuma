#!/bin/bash

set -o errexit
set -o pipefail
set -o nounset

POLICY=$1
VERSION=${2:-"v1alpha1"}

POLICIES_DIR=pkg/plugins/policies
POLICIES_API_DIR="${POLICIES_DIR}/${POLICY}/api/${VERSION}"
POLICIES_CRD_DIR="${POLICIES_DIR}/${POLICY}/k8s/crd"

SCHEMA_TEMPLATE=tools/policy-gen/templates/schema.yaml

echo "Generating schema for ${POLICY}/${VERSION} based on CRD"

function cleanupOnError() {
    rm "${POLICIES_API_DIR}"/schema.yaml
    echo "Script failed, schema.yaml wasn't generated"
}
trap cleanupOnError ERR

cp "${SCHEMA_TEMPLATE}" "${POLICIES_API_DIR}"/schema.yaml

if [ $(ls "${POLICIES_CRD_DIR}" | wc -l) != 1 ]; then
  echo "More than 1 file in crd directory"
  exit 1
fi

CRD_FILE=$(ls "${POLICIES_CRD_DIR}")

yq e '.spec.versions[] | select (.name == "'"${VERSION}"'") | .schema.openAPIV3Schema.properties.spec | del(.type) | del(.description)' \
"${POLICIES_CRD_DIR}/${CRD_FILE}" | yq eval-all -i '. as $item ireduce ({}; . * $item )' \
"${POLICIES_API_DIR}"/schema.yaml -
