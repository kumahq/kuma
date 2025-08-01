KUMA_DIR ?= .
TOOLS_DIR = $(KUMA_DIR)/tools
# Important to use `:=` to only run the script once per make invocation!
BUILD_INFO := $(shell $(TOOLS_DIR)/releases/version.sh)
BUILD_INFO_VERSION = $(word 1, $(BUILD_INFO))
GIT_TAG = $(word 2, $(BUILD_INFO))
GIT_COMMIT = $(word 3, $(BUILD_INFO))
BUILD_DATE = $(word 4, $(BUILD_INFO))
CI_TOOLS_VERSION = $(word 5, $(BUILD_INFO))
# renovate: datasource=github-releases depName=kumahq/envoy-builds versioning=semver
ENVOY_VERSION ?= 1.35.0
KUMA_CHARTS_URL ?= https://kumahq.github.io/charts
CHART_REPO_NAME ?= kuma
PROJECT_NAME ?= kuma

ifeq (,$(shell which mise))
$(error "mise - https://github.com/jdx/mise - not found. Please install it.")
endif
MISE := $(shell which mise)

CI_TOOLS_DIR ?= ${HOME}/.local/share/mise/${PROJECT_NAME}
ifdef XDG_DATA_HOME
	CI_TOOLS_DIR := ${XDG_DATA_HOME}/.local/share/mise/${PROJECT_NAME}
endif
CI_TOOLS_BIN_DIR=$(CI_TOOLS_DIR)/bin

# Change here and `make check` ensures these are used for CI
# Note: These are _docker image tags_
# If changing min version, update mk/kind.mk as well
K8S_MIN_VERSION = v1.27.16-k3s1
K8S_MAX_VERSION=v1.32.2-k3s1
# This should have the same minor version as K8S_MAX_VERSION
KUBEBUILDER_ASSETS_VERSION=1.32

export GO_VERSION=$(shell go mod edit -json | jq -r .Go)
# This needs to rely on version from mise
export GOLANGCI_LINT_VERSION=v2.2.2
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

# A helper to protect calls that push things upstreams (.e.g docker push or github artifact publish)
# $(1) - the actual command to run, if ALLOW_PUSH is not set we'll prefix this with '#' to prevent execution
define GATE_PUSH
$(if $(filter $(ALLOW_PUSH),true),$(1), # $(1))
endef

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KUBECONFIG_DIR := $(HOME)/.kube

PROTOS_DEPS_PATH=$(shell $(MISE) where protoc)/include
# TODO find better way of managing proto deps: https://github.com/kumahq/kuma/issues/13880
XDS_VERSION=$(shell go list -f '{{ .Version }}' -m github.com/cncf/xds/go)
PROTO_XDS=$(shell go mod download github.com/cncf/xds@$(XDS_VERSION) && go list -f '{{ .Dir }}' -m github.com/cncf/xds@$(XDS_VERSION))
PGV_VERSION=$(shell go list -f '{{.Version}}' -m github.com/envoyproxy/protoc-gen-validate)
PROTO_PGV=$(shell go mod download github.com/envoyproxy/protoc-gen-validate@$(PGV_VERSION) && go list -f '{{ .Dir }}' -m github.com/envoyproxy/protoc-gen-validate@$(PGV_VERSION))
PROTO_GOOGLE_APIS=$(shell go mod download github.com/googleapis/googleapis@master && go list -f '{{ .Dir }}' -m github.com/googleapis/googleapis@master)
PROTO_ENVOY=$(shell go mod download github.com/envoyproxy/data-plane-api@main && go list -f '{{ .Dir }}' -m github.com/envoyproxy/data-plane-api@main)

CLANG_FORMAT=$(shell $(MISE) which clang-format)
YQ=$(shell $(MISE) which yq)
HELM=$(shell $(MISE) which helm)
K3D_BIN=$(shell $(MISE) which k3d)
KIND=$(shell $(MISE) which kind)
SETUP_ENVTEST=$(shell $(MISE) which setup-envtest)
KUBEBUILDER_ASSETS=$(shell $(SETUP_ENVTEST) use $(KUBEBUILDER_ASSETS_VERSION) --bin-dir $(CI_TOOLS_BIN_DIR) -p path)
CONTROLLER_GEN=$(shell $(MISE) which controller-gen)
KUBECTL=$(shell $(MISE) which kubectl)
PROTOC_BIN=$(shell $(MISE) which protoc)
SHELLCHECK=$(shell $(MISE) which shellcheck)
CONTAINER_STRUCTURE_TEST=$(shell $(MISE) which container-structure-test)
PROTOC_GEN_GO=$(shell $(MISE) which protoc-gen-go)
PROTOC_GEN_GO_GRPC=$(shell $(MISE) which protoc-gen-go-grpc)
PROTOC_GEN_VALIDATE=$(shell $(MISE) which protoc-gen-validate)
PROTOC_GEN_KUMADOC=$(shell $(MISE) which protoc-gen-kumadoc)
PROTOC_GEN_JSONSCHEMA=$(shell $(MISE) which protoc-gen-jsonschema)
GINKGO=$(shell $(MISE) which ginkgo)
GOLANGCI_LINT=$(shell $(MISE) which golangci-lint)
HELM_DOCS=$(shell $(MISE) which helm-docs)
KUBE_LINTER=$(shell $(MISE) which kube-linter)
HADOLINT=$(shell $(MISE) which hadolint)
OAPI_CODEGEN=$(shell $(MISE) which oapi-codegen)

TOOLS_DEPS_DIRS=$(KUMA_DIR)/mk/dependencies
TOOLS_DEPS_LOCK_FILE=mk/dependencies/deps.lock
TOOLS_MAKEFILE=$(KUMA_DIR)/mk/dev.mk

LATEST_RELEASE_BRANCH := $(shell $(YQ) e '.[] | .branch' versions.yml | grep -v dev | sort -V | tail -n 1)

.PHONY: cmd/check/%
cmd/check/%:
	@command -v "$*" >/dev/null 2>&1 || { \
		echo >&2 "Error: required command '$*' is not in PATH"; \
		exit 1; \
	}

# Install all dependencies on tools and protobuf files
.PHONY: install
install: cmd/check/curl cmd/check/git cmd/check/unzip cmd/check/make cmd/check/go cmd/check/ninja
	$(MISE) install

$(KUBECONFIG_DIR):
	@mkdir -p $(KUBECONFIG_DIR)

# kubectl always writes the current context into the first config file. When
# debugging, it's common to switch contexts and we don't want to edit the Kind
# config files (because then the integration tests have the wrong current
# context). So we create this as a place for kubectl to write the interactive
# current context.
$(KUBECONFIG_DIR)/kind-kuma-current: $(KUBECONFIG_DIR)
	@touch $@

.PHONY: dev/print-latest-release-branch
dev/print-latest-release-branch:
	@echo $(LATEST_RELEASE_BRANCH)

TAKE_FILES_FROM_MASTER = app/kuma-ui/pkg/resources go.mod go.sum deployments/charts/*/Chart.yaml

.PHONY: dev/merge-release
dev/merge-release:
	git merge origin/$(LATEST_RELEASE_BRANCH) --no-commit || true
	git rm -rf $(TAKE_FILES_FROM_MASTER)
	git checkout HEAD -- $(TAKE_FILES_FROM_MASTER)
	@if git diff --name-status --diff-filter=U --exit-code; then\
		echo "Run \`git commit\` to finish merge!";\
	else\
	    echo "Fix above conflicts and run \`git commit\` to finish merge!";\
	fi

# Generate a .envrc that prepends e2e test suite configs to whatever
# KUBECONFIG currently has, and stores CI tooling in .tools.
.PHONY: dev/enrc
dev/envrc: $(KUBECONFIG_DIR)/kind-kuma-current ## Generate .envrc
	@echo 'export CI_TOOLS_DIR=$$(expand_path .tools)' > .envrc
	@for c in $(patsubst %,$(KUBECONFIG_DIR)/kind-%-config,kuma $(K8SCLUSTERS)) $(KUBECONFIG_DIR)/kind-kuma-current ; do \
		echo "path_add KUBECONFIG $$c" ; \
	done >> .envrc
	@echo 'export KUBECONFIG' >> .envrc
	@for prog in $(BUILD_RELEASE_BINARIES) $(BUILD_TEST_BINARIES) ; do \
		echo "PATH_add $(BUILD_ARTIFACTS_DIR)/$$prog" ; \
	done >> .envrc
	@echo 'export KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS)' >> .envrc
	@direnv allow

.PHONY: dev/sync-demo
dev/sync-demo:
	rm app/kumactl/data/install/k8s/demo/*.yaml
	curl -s --fail https://raw.githubusercontent.com/kumahq/kuma-counter-demo/master/demo.yaml | \
		sed 's/"local"/"{{ .Zone }}"/g' | \
		sed 's/\([^/]\)kuma-demo/\1{{ .Namespace }}/g' | \
		sed 's/\([^/]\)kuma-system/\1{{ .SystemNamespace }}/g' \
		> app/kumactl/data/install/k8s/demo/demo.yaml
	curl -s --fail https://raw.githubusercontent.com/kumahq/kuma-counter-demo/master/gateway.yaml | \
		sed 's/\([^/]\)kuma-demo/\1{{ .Namespace }}/g' | \
		sed 's/\([^/]\)kuma-system/\1{{ .SystemNamespace }}/g' \
		> app/kumactl/data/install/k8s/demo/gateway.yaml

.PHONY: dev/set-kuma-helm-repo
dev/set-kuma-helm-repo:
	$(HELM) repo add ${CHART_REPO_NAME} ${KUMA_CHARTS_URL}

.PHONY: clean
clean: clean/build clean/generated clean/docs ## Dev: Clean

.PHONY: dev/fetch-demo
dev/fetch-demo: ## Dev: Fetch demo files
	mkdir -p $(BUILD_DIR)/k8s
	curl -s --fail https://raw.githubusercontent.com/kumahq/kuma-counter-demo/refs/heads/main/k8s/001-with-mtls.yaml > $(BUILD_DIR)/k8s/001-with-mtls.yaml
