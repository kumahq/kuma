#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
set -e

POLICY=$1
VERSION=$2
TOOLS_LOCATION=$3
POLICIES_DIR=$4

POLICIES_API_DIR="${POLICIES_DIR}/${POLICY}/api/${VERSION}"
POLICIES_CRD_DIR="${POLICIES_DIR}/${POLICY}/k8s/crd"

SCHEMA_TEMPLATE="${TOOLS_LOCATION}/openapi/templates/schema.yaml"

# 1. Copy file ${SCHEMA_TEMPLATE} to ${POLICIES_API_DIR}/schema.yaml. It contains
#    information about fields that are equal for all resources 'type', 'mesh' and 'name'.
#
# 2. Using yq extract item from the list '.spec.version[]' that has ${VERSION} and
#    take '.schema.openAPIV3Schema.properties.spec'.
#
# 3. Delete 'type' and 'description' for the extracted item, because these are 'type'
#    and 'description' for the 'spec' field.
#
# 4. Using yq eval-all with ireduce merge the file from Step 1 and output from Step 3,
#    placing the result into the file from Step 1

echo "Generating schema for ${POLICY}/${VERSION} based on CRD"

function cleanupOnError() {
    rm "${POLICIES_API_DIR}"/schema.yaml
    echo "Script failed, schema.yaml wasn't generated"
}
trap cleanupOnError ERR

cp "${SCHEMA_TEMPLATE}" "${POLICIES_API_DIR}"/schema.yaml

if [ "$(find "${POLICIES_CRD_DIR}" -type f | wc -l | xargs echo)" != 1 ]; then
  echo "Exactly 1 file is expected in ${POLICIES_CRD_DIR}"
  exit 1
fi

CRD_FILE=$(find "${POLICIES_CRD_DIR}" -type f)

yq e -i ".properties *= (load(\"${CRD_FILE}\") | ((.spec.versions[] | select (.name == \"${VERSION}\") | .schema.openAPIV3Schema.properties | del(.apiVersion) | del(.metadata) | del(.kind)) * {\"type\": {\"enum\": [.spec.names.kind]}}))" "${POLICIES_API_DIR}"/schema.yaml
