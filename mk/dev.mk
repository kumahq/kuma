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
ENVOY_VERSION ?= 1.37.0
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
K8S_MIN_VERSION=v1.31.12-k3s1
K8S_MAX_VERSION=v1.34.1-k3s1
# This should have the same minor version as K8S_MAX_VERSION
KUBEBUILDER_ASSETS_VERSION=1.33

GO := $(shell $(MISE) which go)
export GO_VERSION := $(shell $(GO) mod edit -json | jq -r .Go)
GOOS := $(shell $(GO) env GOOS)
GOARCH := $(shell $(GO) env GOARCH)

# A helper to protect calls that push things upstreams (.e.g docker push or github artifact publish)
# $(1) - the actual command to run, if ALLOW_PUSH is not set we'll prefix this with '#' to prevent execution
define GATE_PUSH
$(if $(filter $(ALLOW_PUSH),true),$(1), # $(1))
endef

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KUBECONFIG_DIR := $(HOME)/.kube

PROTOS_DEPS_PATH=$(shell $(MISE) where protoc)/include

BUF=$(shell $(MISE) which buf)

# Proto dependencies via Buf
BUF_CACHE_DIR := $(CI_TOOLS_DIR)/buf/cache
PROTO_GOOGLE_APIS := $(shell $(BUF) export buf.build/googleapis/googleapis --output $(BUF_CACHE_DIR)/googleapis && echo $(BUF_CACHE_DIR)/googleapis)
PROTO_PGV := $(shell $(BUF) export buf.build/envoyproxy/protoc-gen-validate --output $(BUF_CACHE_DIR)/pgv && echo $(BUF_CACHE_DIR)/pgv)
PROTO_ENVOY := $(shell $(BUF) export buf.build/envoyproxy/envoy --output $(BUF_CACHE_DIR)/envoy && echo $(BUF_CACHE_DIR)/envoy)
PROTO_XDS := $(shell $(BUF) export buf.build/cncf/xds --output $(BUF_CACHE_DIR)/xds && echo $(BUF_CACHE_DIR)/xds)
YQ=$(shell $(MISE) which yq)
HELM=$(shell $(MISE) which helm)
K3D_BIN=$(MISE) exec -- k3d
KIND=$(shell $(MISE) which kind)
SETUP_ENVTEST=$(shell $(MISE) which setup-envtest)
KUBEBUILDER_ASSETS=$(shell $(SETUP_ENVTEST) use $(KUBEBUILDER_ASSETS_VERSION) --bin-dir $(CI_TOOLS_BIN_DIR) -p path)
CONTROLLER_GEN=$(shell $(MISE) which controller-gen)
KUBECTL=$(shell $(MISE) which kubectl)
PROTOC_BIN=$(shell $(MISE) which protoc)
SHELLCHECK=$(shell $(MISE) which shellcheck)
ACTIONLINT=$(shell $(MISE) which actionlint)
CONTAINER_STRUCTURE_TEST=$(shell $(MISE) which container-structure-test)
PROTOC_GEN_GO=$(shell $(MISE) which protoc-gen-go)
PROTOC_GEN_GO_GRPC=$(shell $(MISE) which protoc-gen-go-grpc)
PROTOC_GEN_VALIDATE=$(MISE) exec -- protoc-gen-validate
PROTOC_GEN_KUMADOC=$(MISE) exec -- protoc-gen-kumadoc
PROTOC_GEN_JSONSCHEMA=$(shell $(MISE) which protoc-gen-jsonschema)
GINKGO=$(shell $(MISE) which ginkgo)
GOLANGCI_LINT=$(shell $(MISE) which golangci-lint)
HELM_DOCS=$(shell $(MISE) which helm-docs)
KUBE_LINTER=$(shell $(MISE) which kube-linter)
HADOLINT=$(shell $(MISE) which hadolint)
# oapi-codegen: mise go: backend installs to CI_TOOLS_BIN_DIR, mise which doesn't find it
OAPI_CODEGEN=$(shell test -f $(CI_TOOLS_BIN_DIR)/oapi-codegen && echo $(CI_TOOLS_BIN_DIR)/oapi-codegen || command -v oapi-codegen)

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
