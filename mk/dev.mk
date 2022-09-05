KUMA_DIR ?= .
ENVOY_VERSION = $(shell ${KUMA_DIR}/tools/envoy/version.sh)

CI_TOOLS_DIR ?= ${HOME}/.kuma-dev
ifdef XDG_DATA_HOME
	CI_TOOLS_DIR := ${XDG_DATA_HOME}/kuma-dev
endif
CI_TOOLS_BIN_DIR=$(CI_TOOLS_DIR)/bin

GOARCH := $(shell go env GOARCH)

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KUBECONFIG_DIR := $(HOME)/.kube

PROTOS_DEPS_PATH=$(CI_TOOLS_DIR)/protos

TOOLS_DIR ?= $(shell pwd)/tools
CLANG_FORMAT=$(CI_TOOLS_BIN_DIR)/clang-format
HELM=$(CI_TOOLS_BIN_DIR)/helm
K3D_BIN=$(CI_TOOLS_BIN_DIR)/k3d
KIND=$(CI_TOOLS_BIN_DIR)/kind
KUBEBUILDER=$(CI_TOOLS_BIN_DIR)/kubebuilder
KUBEBUILDER_ASSETS=$(CI_TOOLS_BIN_DIR)
KUBECTL=$(CI_TOOLS_BIN_DIR)/kubectl
PROTOC_BIN=$(CI_TOOLS_BIN_DIR)/protoc
SHELLCHECK=$(CI_TOOLS_BIN_DIR)/shellcheck
# from go-deps
PROTOC_GEN_GO=$(CI_TOOLS_BIN_DIR)/protoc-gen-go
PROTOC_GEN_GO_GRPC=$(CI_TOOLS_BIN_DIR)/protoc-gen-go-grpc
PROTOC_GEN_VALIDATE=$(CI_TOOLS_BIN_DIR)/protoc-gen-validate
PROTOC_GEN_KUMADOC=$(CI_TOOLS_BIN_DIR)/protoc-gen-kumadoc
GINKGO=$(CI_TOOLS_BIN_DIR)/ginkgo
GOLANGCI_LINT=$(CI_TOOLS_BIN_DIR)/golangci-lint
HELM_DOCS=$(CI_TOOLS_BIN_DIR)/helm-docs
KUBE_LINTER=$(CI_TOOLS_BIN_DIR)/kube-linter

# Install all dependencies on tools and protobuf files
# We add one script per tool in the `mk/dependencies` folder. Add a VARIABLE for each binary and use this everywhere in Makefiles
# ideally the tool should be idempotent to make things quick to rerun.
# it's important that everything lands in $(CI_TOOLS_DIR) to be able to cache this folder in CI and speed up the build.
.PHONY: dev/tools
dev/tools: ## Bootstrap: Install all development tools
	@mkdir -p $(CI_TOOLS_BIN_DIR) $(CI_TOOLS_DIR)/protos
	@for i in mk/dependencies/*.sh; do OS=$(GOOS) ARCH=$(GOARCH) $$i $(CI_TOOLS_DIR); done
	# Compute a hash to use for caching
	@for i in mk/dependencies/*.sh; do echo "---$${i}"; cat $${i}; done | git hash-object --stdin > mk/dependencies/deps.lock
	@echo "All non code dependencies installed, if you use these tools outside of make add $(CI_TOOLS_BIN_DIR) to your PATH"

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
