KUMA_DIR ?= .
TOOLS_DIR = $(KUMA_DIR)/tools
# Important to use `:=` to only run the script once per make invocation!
BUILD_INFO := $(shell $(TOOLS_DIR)/releases/version.sh)
BUILD_INFO_VERSION = $(word 1, $(BUILD_INFO))
GIT_TAG = $(word 2, $(BUILD_INFO))
GIT_COMMIT = $(word 3, $(BUILD_INFO))
BUILD_DATE = $(word 4, $(BUILD_INFO))
CI_TOOLS_VERSION = $(word 5, $(BUILD_INFO))
# renovate: datasource=github-tags depName=envoy packageName=kumahq/envoy-builds versioning=semver
ENVOY_VERSION ?= 1.39.1
KUMA_CHARTS_URL ?= https://kumahq.github.io/charts
CHART_REPO_NAME ?= kuma
PROJECT_NAME ?= kuma

MISE := $(shell which mise)
ifeq (,$(MISE))
$(error "mise - https://github.com/jdx/mise - not found. Please install it.")
endif

CI_TOOLS_DIR ?= ${HOME}/.local/share/mise/${PROJECT_NAME}
ifdef XDG_DATA_HOME
	CI_TOOLS_DIR := ${XDG_DATA_HOME}/.local/share/mise/${PROJECT_NAME}
endif
CI_TOOLS_BIN_DIR=$(CI_TOOLS_DIR)/bin

# Change here and `make check` ensures these are used for CI
# Note: These are _docker image tags_
# If changing min version, update mk/kind.mk as well
K8S_MIN_VERSION=v1.34.9-k3s1
K8S_MAX_VERSION=v1.35.3-k3s1
# This should have the same minor version as K8S_MAX_VERSION
KUBEBUILDER_ASSETS_VERSION=1.33

GO := $(shell $(MISE) which go)
JQ := $(shell $(MISE) which jq)
export GO_VERSION := $(shell $(GO) mod edit -json | $(JQ) -r .Go)
GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)

# A helper to protect calls that push things upstreams (.e.g docker push or github artifact publish)
# $(1) - the actual command to run, if ALLOW_PUSH is not set we'll prefix this with '#' to prevent execution
define GATE_PUSH
$(if $(filter $(ALLOW_PUSH),true),$(1), # $(1))
endef

# KUBECONFIG_DIR is defined in mk/k8s.mk (included via Makefile).
# Do not redefine it here.

# _once(VAR,value): on first expansion replace VAR with the simply-expanded
# value, so every make invocation does not pay for lookups it never uses.
_once = $(eval $(1) := $(2))$($(1))

PROTOS_DEPS_PATH = $(call _once,PROTOS_DEPS_PATH,$(shell $(MISE) where protoc)/include)

BUF = $(call _once,BUF,$(shell $(MISE) which buf))

# Proto dependencies via Buf, exported by dev/protos/deps (network, so not at parse time)
BUF_CACHE_DIR := $(CI_TOOLS_DIR)/buf/cache
PROTO_GOOGLE_APIS := $(BUF_CACHE_DIR)/googleapis
PROTO_PGV := $(BUF_CACHE_DIR)/pgv
# Envoy does not publish a BSR proto snapshot for every patch release, but always does
# for a minor, so follow ENVOY_VERSION's minor. Override to pick up a later patch's API.
ENVOY_PROTO_VERSION ?= v$(basename $(ENVOY_VERSION)).0
PROTO_ENVOY := $(BUF_CACHE_DIR)/envoy
PROTO_XDS := $(BUF_CACHE_DIR)/xds
YQ = $(call _once,YQ,$(shell $(MISE) which yq))
HELM = $(call _once,HELM,$(shell $(MISE) which helm))
K3D=$(MISE) exec -- k3d
KIND = $(call _once,KIND,$(shell $(MISE) which kind))
SETUP_ENVTEST = $(call _once,SETUP_ENVTEST,$(shell $(MISE) which setup-envtest))
KUBEBUILDER_ASSETS = $(call _once,KUBEBUILDER_ASSETS,$(shell $(SETUP_ENVTEST) use $(KUBEBUILDER_ASSETS_VERSION) --bin-dir $(CI_TOOLS_BIN_DIR) -p path))
CONTROLLER_GEN = $(call _once,CONTROLLER_GEN,$(shell $(MISE) which controller-gen))
KUBECTL = $(call _once,KUBECTL,$(shell $(MISE) which kubectl))
PROTOC_BIN = $(call _once,PROTOC_BIN,$(shell $(MISE) which protoc))
SHELLCHECK = $(call _once,SHELLCHECK,$(shell $(MISE) which shellcheck))
ACTIONLINT = $(call _once,ACTIONLINT,$(shell $(MISE) which actionlint))
CONTAINER_STRUCTURE_TEST = $(call _once,CONTAINER_STRUCTURE_TEST,$(shell $(MISE) which container-structure-test))
PROTOC_GEN_GO = $(call _once,PROTOC_GEN_GO,$(shell $(MISE) which protoc-gen-go))
PROTOC_GEN_GO_GRPC = $(call _once,PROTOC_GEN_GO_GRPC,$(shell $(MISE) which protoc-gen-go-grpc))
PROTOC_GEN_VALIDATE=$(MISE) exec -- protoc-gen-validate
PROTOC_GEN_KUMADOC=$(MISE) exec -- protoc-gen-kumadoc
PROTOC_GEN_JSONSCHEMA = $(call _once,PROTOC_GEN_JSONSCHEMA,$(shell $(MISE) which protoc-gen-jsonschema))
GINKGO = $(call _once,GINKGO,$(shell $(MISE) which ginkgo))
GOLANGCI_LINT = $(call _once,GOLANGCI_LINT,$(shell $(MISE) which golangci-lint))
HELM_DOCS = $(call _once,HELM_DOCS,$(shell $(MISE) which helm-docs))
KUBE_LINTER = $(call _once,KUBE_LINTER,$(shell $(MISE) which kube-linter))
HADOLINT = $(call _once,HADOLINT,$(shell $(MISE) which hadolint))
DASHBOARD_LINTER = $(call _once,DASHBOARD_LINTER,$(shell $(MISE) which dashboard-linter))
# oapi-codegen: mise go: backend installs to CI_TOOLS_BIN_DIR, mise which doesn't find it
OAPI_CODEGEN = $(call _once,OAPI_CODEGEN,$(shell test -f $(CI_TOOLS_BIN_DIR)/oapi-codegen && echo $(CI_TOOLS_BIN_DIR)/oapi-codegen || command -v oapi-codegen))

TOOLS_DEPS_DIRS=$(KUMA_DIR)/mk/dependencies
TOOLS_DEPS_LOCK_FILE=mk/dependencies/deps.lock
TOOLS_MAKEFILE=$(KUMA_DIR)/mk/dev.mk

LATEST_RELEASE_BRANCH = $(call _once,LATEST_RELEASE_BRANCH,$(shell $(YQ) e '.[] | .branch' versions.yml | grep -v dev | sort -V | tail -n 1))

.PHONY: dev/protos/deps
dev/protos/deps: ## Dev: Export third-party proto dependencies with buf
	$(BUF) export buf.build/googleapis/googleapis --output $(PROTO_GOOGLE_APIS)
	$(BUF) export buf.build/envoyproxy/protoc-gen-validate --output $(PROTO_PGV)
	$(BUF) export buf.build/envoyproxy/envoy:$(ENVOY_PROTO_VERSION) --output $(PROTO_ENVOY)
	$(BUF) export buf.build/cncf/xds --output $(PROTO_XDS)

.PHONY: cmd/check/%
cmd/check/%:
	@command -v "$*" >/dev/null 2>&1 || { \
		echo >&2 "Error: required command '$*' is not in PATH"; \
		exit 1; \
	}

# Install all dependencies on tools and protobuf files
.PHONY: install
install: cmd/check/curl cmd/check/git cmd/check/unzip cmd/check/make cmd/check/go
	$(MISE) install
	$(BUF) dep update

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
	@for c in \
		$(KUBECONFIG_DIR)/kind-kuma.yaml \
		$(patsubst %,$(KUBECONFIG_DIR)/kind-%.yaml,$(K8SCLUSTERS)) \
		$(KUBECONFIG_DIR)/k3d-kuma.yaml \
		$(patsubst %,$(KUBECONFIG_DIR)/k3d-%.yaml,$(K8SCLUSTERS)) \
		$(KUBECONFIG_DIR)/kind-kuma-current ; do \
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
