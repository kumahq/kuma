KUMA_DIR ?= .
TOOLS_DIR = $(KUMA_DIR)/tools
# Important to use `:=` to only run the script once per make invocation!
BUILD_INFO := $(shell $(TOOLS_DIR)/releases/version.sh)
BUILD_INFO_VERSION = $(word 1, $(BUILD_INFO))
GIT_TAG = $(word 2, $(BUILD_INFO))
GIT_COMMIT = $(word 3, $(BUILD_INFO))
BUILD_DATE = $(word 4, $(BUILD_INFO))
CI_TOOLS_VERSION = $(word 5, $(BUILD_INFO))
ENVOY_VERSION ?= 1.30.6
KUMA_CHARTS_URL ?= https://kumahq.github.io/charts
CHART_REPO_NAME ?= kuma
PROJECT_NAME ?= kuma

CI_TOOLS_DIR ?= ${HOME}/.kuma-dev/${PROJECT_NAME}-${CI_TOOLS_VERSION}
ifdef XDG_DATA_HOME
	CI_TOOLS_DIR := ${XDG_DATA_HOME}/kuma-dev/${PROJECT_NAME}-${CI_TOOLS_VERSION}
endif
CI_TOOLS_BIN_DIR=$(CI_TOOLS_DIR)/bin

# Change here and `make check` ensures these are used for CI
# Note: These are _docker image tags_
# If changing min version, update mk/kind.mk as well
K8S_MIN_VERSION = v1.25.16-k3s1
K8S_MAX_VERSION = v1.31.1-k3s1
export GO_VERSION=$(shell go mod edit -json | jq -r .Go)
export GOLANGCI_LINT_VERSION=v1.60.3
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

PROTOS_DEPS_PATH=$(CI_TOOLS_DIR)/protos

CLANG_FORMAT=$(CI_TOOLS_BIN_DIR)/clang-format
YQ=$(CI_TOOLS_BIN_DIR)/yq
HELM=$(CI_TOOLS_BIN_DIR)/helm
K3D_BIN=$(CI_TOOLS_BIN_DIR)/k3d
KIND=$(CI_TOOLS_BIN_DIR)/kind
KUBEBUILDER=$(CI_TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_ASSETS=$(CI_TOOLS_BIN_DIR)
CONTROLLER_GEN=$(CI_TOOLS_BIN_DIR)/controller-gen
KUBECTL=$(CI_TOOLS_BIN_DIR)/kubectl
PROTOC_BIN=$(CI_TOOLS_BIN_DIR)/protoc
SHELLCHECK=$(CI_TOOLS_BIN_DIR)/shellcheck
CONTAINER_STRUCTURE_TEST=$(CI_TOOLS_BIN_DIR)/container-structure-test
# from go-deps
PROTOC_GEN_GO=$(CI_TOOLS_BIN_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC=$(CI_TOOLS_BIN_DIR)/protoc-gen-go-grpc
PROTOC_GEN_VALIDATE=$(CI_TOOLS_BIN_DIR)/protoc-gen-validate
PROTOC_GEN_KUMADOC=$(CI_TOOLS_BIN_DIR)/protoc-gen-kumadoc
PROTOC_GEN_JSONSCHEMA=$(CI_TOOLS_BIN_DIR)/protoc-gen-jsonschema
GINKGO=$(CI_TOOLS_BIN_DIR)/ginkgo
GOLANGCI_LINT=$(CI_TOOLS_BIN_DIR)/golangci-lint
HELM_DOCS=$(CI_TOOLS_BIN_DIR)/helm-docs
KUBE_LINTER=$(CI_TOOLS_BIN_DIR)/kube-linter
HADOLINT=$(CI_TOOLS_BIN_DIR)/hadolint

TOOLS_DEPS_DIRS=$(KUMA_DIR)/mk/dependencies
TOOLS_DEPS_LOCK_FILE=mk/dependencies/deps.lock
TOOLS_MAKEFILE=$(KUMA_DIR)/mk/dev.mk

LATEST_RELEASE_BRANCH := $(shell $(YQ) e '.[] | .branch' versions.yml | grep -v dev | sort -V | tail -n 1)

# Install all dependencies on tools and protobuf files
# We add one script per tool in the `mk/dependencies` folder. Add a VARIABLE for each binary and use this everywhere in Makefiles
# ideally the tool should be idempotent to make things quick to rerun.
# it's important that everything lands in $(CI_TOOLS_DIR) to be able to cache this folder in CI and speed up the build.
.PHONY: dev/tools
dev/tools: ## Bootstrap: Install all development tools
	$(TOOLS_DIR)/dev/install-dev-tools.sh $(CI_TOOLS_BIN_DIR) $(CI_TOOLS_DIR) "$(TOOLS_DEPS_DIRS)" $(TOOLS_DEPS_LOCK_FILE) $(GOOS) $(GOARCH) $(TOOLS_MAKEFILE)

.PHONY: dev/tools/clean
dev/tools/clean: ## Bootstrap: Remove all development tools
	rm -rf $(CI_TOOLS_DIR)

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
		sed 's/\([^/]\)kuma-demo/\1{{ .Namespace }}/g' \
		> app/kumactl/data/install/k8s/demo/demo.yaml
	curl -s --fail https://raw.githubusercontent.com/kumahq/kuma-counter-demo/master/gateway.yaml | \
		sed 's/\([^/]\)kuma-demo/\1{{ .Namespace }}/g' \
		> app/kumactl/data/install/k8s/demo/gateway.yaml

.PHONY: dev/set-kuma-helm-repo
dev/set-kuma-helm-repo:
	${CI_TOOLS_BIN_DIR}/helm repo add ${CHART_REPO_NAME} ${KUMA_CHARTS_URL}

.PHONY: clean
clean: clean/build clean/generated clean/docs ## Dev: Clean
