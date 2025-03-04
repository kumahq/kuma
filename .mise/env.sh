#!/usr/bin/env bash

set -o errexit  # Exit immediately if any command has a non-zero exit status (error)
set -o errtrace # Ensure that traps are inherited by functions, command substitutions, and subshells
set -o nounset  # Treat any unset variables as an error and exit immediately
set -o pipefail # If any command in a pipeline fails, the entire pipeline fails

export __GOPATH_MOD="${GOPATH_MOD:-$__GOPATH/pkg/mod}"
export __GOPATH_MOD_GITHUB="$__GOPATH_MOD/github.com"

# Project directory setup
export __KUMA_DIR="${KUMA_DIR:-.}"

export __TASKS_SCRIPTS_PATH="${TASKS_SCRIPTS_PATH:-$(realpath "$__KUMA_DIR")/.mise/scripts}"

# Kubernetes version range for CI
export __K8S_MIN_VERSION="v1.25.16-k3s4"
export __K8S_MAX_VERSION="v1.31.1-k3s1"

# GitHub Action prefix (can be set externally)
export __ACTION_PREFIX=""

# Derive the project name from the current directory if not explicitly set
export __PROJECT_NAME="${PROJECT_NAME:-$(basename "${PWD:-kuma}")}"

export __BUILD_DIR="${BUILD_DIR:-$(realpath "$__KUMA_DIR")/build}"
export __TOOLS_DIR="${TOOLS_DIR:-$__BUILD_DIR/tools}"
export __TOOLS_DIR_BIN="${__TOOLS_DIR}/bin"
export __TOOLS_DIR_PROTOS="${__TOOLS_DIR}/protos"

# Paths for tools and imports
export __GENERATE_ENVOY_IMPORTS="./pkg/xds/envoy/imports.go"
export __GENERATE_TOOLS_DIR="${__TOOLS_DIR_BIN}/${__GOOS}-${__GOARCH}"
export __GENERATE_TOOLS_POLICY_GEN_SOURCE="${__KUMA_DIR}/tools/policy-gen"
export __GENERATE_TOOLS_POLICY_GEN_BIN="${__GENERATE_TOOLS_DIR}/policy-gen"
export __GENERATE_TOOLS_RESOURCE_GEN_SOURCE="${__KUMA_DIR}/tools/resource-gen"
export __GENERATE_TOOLS_RESOURCE_GEN_BIN="${__GENERATE_TOOLS_DIR}/resource-gen"

# Protobuf directories and Go module
export __GENERATE_PROTO_DIRS="${GENERATE_PROTO_DIRS:-./api ./pkg/config ./pkg/plugins ./test/server/grpc/api}"
export __GENERATE_GO_MODULE="${GENERATE_GO_MODULE:-github.com/kumahq/kuma}"

# Helm file paths and settings
export __GENERATE_HELM_CHARTS_DIR="${HELM_CHARTS_DIR:-deployments/charts}"
export __GENERATE_HELM_VALUES_FILE="${HELM_VALUES_FILE:-deployments/charts/kuma/values.yaml}"
export __GENERATE_HELM_CRDS_DIR="${HELM_CRDS_DIR:-deployments/charts/kuma/crds}"
export __GENERATE_HELM_VALUES_FILE_POLICY_PATH="${HELM_VALUES_FILE_POLICY_PATH:-.plugins.policies}"

# Generation prerequisites and additional dependencies
export __GENERATE_OAS_PREREQUISITES="${GENERATE_OAS_PREREQUISITES:-}"
export __GENERATE_EXTRA_DEPS_TARGETS="${EXTRA_DEPS_TARGETS:-generate::envoy-imports}"

export __GENERATE_POLICIES_DIR="${POLICIES_DIR:-pkg/plugins/policies}"
export __GENERATE_RESOURCES_DIR="${RESOURCES_DIR:-pkg/core/resources/apis}"
