#!/bin/bash
OUTPUT_BIN_DIR=$1/bin
OUTPUT_PROTO_DIR=$1/protos

# Use go list -m for deps that are also on go.mod this way we use dependabot a live on the same version

PGV=github.com/envoyproxy/protoc-gen-validate@$(go list -m github.com/envoyproxy/protoc-gen-validate | cut -d ' ' -f2)
PGKUMADOC=github.com/kumahq/protoc-gen-kumadoc@$(go list -m github.com/kumahq/protoc-gen-kumadoc | cut -d ' ' -f2)
for i in \
    google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1 \
    google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(go list -m google.golang.org/grpc/cmd/protoc-gen-go-grpc | cut -d ' ' -f2) \
    ${PGV} \
    ${PGKUMADOC} \
    github.com/onsi/ginkgo/v2/ginkgo@$(go list -m github.com/onsi/ginkgo/v2 | cut -d ' ' -f2) \
    github.com/golangci/golangci-lint/cmd/golangci-lint@v1.49.0 \
    github.com/mikefarah/yq/v4@v4.27.2 \
    github.com/norwoodj/helm-docs/cmd/helm-docs@v1.11.0 \
    golang.stackrox.io/kube-linter/cmd/kube-linter@v0.0.0-20220513142942-846f273ed465; do
  echo "install go dep: ${i}"
  GOBIN=${OUTPUT_BIN_DIR} go install "${i}"
done

set +x
# Get the protos from some go dependencies
#
ROOT=$(go env GOPATH)/pkg/mod

function cpOnlyProto() {
  pushd "${1}" || exit
  # shellcheck disable=SC2044
  for i in $(find . -name '*.proto'); do
    local base_path
    base_path=${2}/$(dirname "${i}")
    mkdir -p "${base_path}" && install "${i}" "${base_path}"
  done
  popd || exit
}

rm -fr "${OUTPUT_PROTO_DIR}"/udpa "${OUTPUT_PROTO_DIR}"/xds
mkdir -p "${OUTPUT_PROTO_DIR}"/{udpa,xds}
go mod download github.com/cncf/udpa@master
VERSION=$(find "${ROOT}"/github.com/cncf/udpa@* -maxdepth 0 | sort -r | head -1)
cpOnlyProto "${VERSION}"/udpa "${OUTPUT_PROTO_DIR}"/udpa
cpOnlyProto "${VERSION}"/xds "${OUTPUT_PROTO_DIR}"/xds

rm -fr "${OUTPUT_PROTO_DIR}"/envoy
mkdir -p "${OUTPUT_PROTO_DIR}"
go mod download github.com/envoyproxy/data-plane-api@main
VERSION=$(find "${ROOT}"/github.com/envoyproxy/data-plane-api@* -maxdepth 0 | sort -r | head -1)
cpOnlyProto "${VERSION}"/envoy "${OUTPUT_PROTO_DIR}"/envoy

rm -fr "${OUTPUT_PROTO_DIR}"/validate
mkdir -p "${OUTPUT_PROTO_DIR}"/validate
go mod download "${PGV}"
cpOnlyProto "${ROOT}"/"${PGV}"/validate "${OUTPUT_PROTO_DIR}"/validate

rm -fr "${OUTPUT_PROTO_DIR}"/kuma-doc
mkdir -p "${OUTPUT_PROTO_DIR}"/kuma-doc
go mod download "${PGKUMADOC}"
cpOnlyProto "${ROOT}"/"${PGKUMADOC}"/proto "${OUTPUT_PROTO_DIR}"/kuma-doc

rm -rf "${OUTPUT_PROTO_DIR}"/google/{api,rpc}
mkdir -p "${OUTPUT_PROTO_DIR}"/google/{api,rpc}

go mod download github.com/googleapis/googleapis@master
VERSION=$(find "${ROOT}"/github.com/googleapis/googleapis@* -maxdepth 0 | sort -r | head -1)
cpOnlyProto "${VERSION}"/google/api "${OUTPUT_PROTO_DIR}"/google/api
cpOnlyProto "${VERSION}"/google/rpc "${OUTPUT_PROTO_DIR}"/google/rpc
