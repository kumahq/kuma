GINKGO_VERSION := v2.1.3
GOLANGCI_LINT_VERSION := v1.43.0
GOLANG_PROTOBUF_VERSION := v1.5.2
HELM_DOCS_VERSION := 1.5.0
KUSTOMIZE_VERSION := v4.4.1
PROTOC_PGV_VERSION := v0.4.1
PROTOC_VERSION := 3.14.0
UDPA_LATEST_VERSION := main
GOOGLEAPIS_LATEST_VERSION := master
KUMADOC_VERSION := v0.1.7
DATAPLANE_API_LATEST_VERSION := main
SHELLCHECK_VERSION := v0.8.0

CI_KUBEBUILDER_VERSION ?= 2.3.2
CI_MINIKUBE_VERSION ?= v1.24.0
CI_KUBECTL_VERSION ?= v1.18.14

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
export PATH := $(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KUBECONFIG_DIR := $(HOME)/.kube

PROTOC_PATH := $(CI_TOOLS_DIR)/protoc
PROTOBUF_WKT_DIR := $(CI_TOOLS_DIR)/protobuf.d
KUBEBUILDER_DIR := $(CI_TOOLS_DIR)/kubebuilder.d
KUBEBUILDER_PATH := $(CI_TOOLS_DIR)/kubebuilder
KUSTOMIZE_PATH := $(CI_TOOLS_DIR)/kustomize
MINIKUBE_PATH := $(CI_TOOLS_DIR)/minikube
KUBECTL_PATH := $(CI_TOOLS_DIR)/kubectl
KUBE_APISERVER_PATH := $(CI_TOOLS_DIR)/kube-apiserver
ETCD_PATH := $(CI_TOOLS_DIR)/etcd
GOLANGCI_LINT_DIR := $(CI_TOOLS_DIR)
HELM_DOCS_PATH := $(CI_TOOLS_DIR)/helm-docs
SHELLCHECK_PATH := $(CI_TOOLS_DIR)/shellcheck

TOOLS_DIR ?= $(shell pwd)/tools

PROTOC_OS=unknown
PROTOC_ARCH=$(shell uname -m)

UNAME_S := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
ifeq ($(UNAME_S), Linux)
	PROTOC_OS=linux
	SHELLCHECK_OS=linux
else
	ifeq ($(UNAME_S), Darwin)
		PROTOC_OS=osx
		SHELLCHECK_OS=darwin
	endif
endif

HELM_DOCS_ARCH := $(shell uname -m)
ifeq ($(UNAME_ARCH), aarch64)
	PROTOC_ARCH=aarch_64
	HELM_DOCS_ARCH=arm64
endif

CURL_PATH ?= curl
CURL_DOWNLOAD := $(CURL_PATH) --location --fail --progress-bar

.PHONY: dev/tools
dev/tools: dev/tools/all ## Bootstrap: Install all development tools

.PHONY: dev/tools/all
dev/tools/all: dev/install/protoc dev/install/protobuf-wellknown-types \
	dev/install/protoc-gen-go dev/install/protoc-gen-validate \
	dev/install/protoc-gen-kumadoc \
	dev/install/ginkgo \
	dev/install/kubectl \
	dev/install/kubebuilder \
	dev/install/kustomize \
	dev/install/kind \
	dev/install/k3d \
	dev/install/minikube \
	dev/install/golangci-lint \
	dev/install/helm3 \
	dev/install/helm-docs \
	dev/install/data-plane-api \
	dev/install/shellcheck

.PHONY: dev/install/protoc-gen-kumadoc
dev/install/protoc-gen-kumadoc:
	go install github.com/kumahq/protoc-gen-kumadoc@$(KUMADOC_VERSION)

.PHONY: dev/install/data-plane-api
dev/install/data-plane-api:
	go get github.com/envoyproxy/data-plane-api@$(DATAPLANE_API_LATEST_VERSION)
	go get github.com/cncf/udpa@$(UDPA_LATEST_VERSION)
	go get github.com/googleapis/googleapis@$(GOOGLEAPIS_LATEST_VERSION)

.PHONY: dev/install/protoc
dev/install/protoc: ## Bootstrap: Install Protoc (protobuf compiler)
	@if [ -e $(PROTOC_PATH) ]; then echo "Protoc $$( $(PROTOC_PATH) --version ) is already installed at $(PROTOC_PATH)" ; fi
	@if [ ! -e $(PROTOC_PATH) ]; then \
		echo "Installing Protoc $(PROTOC_VERSION) ..." \
		&& set -x \
		&& mkdir -p /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& $(CURL_DOWNLOAD) -o /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip bin/protoc -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& cp /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH)/bin/protoc $(PROTOC_PATH) \
		&& rm -rf /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& rm /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& set +x \
		&& echo "Protoc $(PROTOC_VERSION) has been installed at $(PROTOC_PATH)" ; fi

.PHONY: dev/install/protobuf-wellknown-types
dev/install/protobuf-wellknown-types:: ## Bootstrap: Install Protobuf well-known types
	@if [ -e $(PROTOBUF_WKT_DIR) ]; then echo "Protobuf well-known types are already installed at $(PROTOBUF_WKT_DIR)" ; fi
	@if [ ! -e $(PROTOBUF_WKT_DIR) ]; then \
		echo "Installing Protobuf well-known types $(PROTOC_VERSION) ..." \
		&& set -x \
		&& mkdir -p /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& $(CURL_DOWNLOAD) -o /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip 'include/*' -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(PROTOBUF_WKT_DIR) \
		&& cp -r /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH)/include $(PROTOBUF_WKT_DIR) \
		&& rm -rf /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& rm /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& set +x \
		&& echo "Protobuf well-known types $(PROTOC_VERSION) have been installed at $(PROTOBUF_WKT_DIR)" ; fi

.PHONY: dev/install/protoc-gen-go
dev/install/protoc-gen-go: ## Bootstrap: Install Protoc Go Plugin (protobuf Go generator)
	go install github.com/golang/protobuf/protoc-gen-go@$(GOLANG_PROTOBUF_VERSION)

.PHONY: dev/install/protoc-gen-validate
dev/install/protoc-gen-validate: ## Bootstrap: Install Protoc Gen Validate Plugin (protobuf validation code generator)
	go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION)

.PHONY: dev/install/ginkgo
dev/install/ginkgo: ## Bootstrap: Install Ginkgo (BDD testing framework)
	# see https://github.com/onsi/ginkgo#set-me-up
	echo "Installing Ginkgo ..."
	go install github.com/onsi/ginkgo/v2/ginkgo@$(GINKGO_VERSION)  # installs the ginkgo CLI
	echo "Ginkgo has been installed at $(GOPATH_BIN_DIR)/ginkgo"

.PHONY: dev/install/kubebuilder
dev/install/kubebuilder: ## Bootstrap: Install Kubebuilder (including etcd and kube-apiserver)
	# see https://book.kubebuilder.io/quick-start.html#installation
	@if [ -e $(KUBEBUILDER_PATH) ]; then echo "Kubebuilder $$( $(KUBEBUILDER_PATH) version ) is already installed at $(KUBEBUILDER_PATH)" ; fi
	@if [ ! -e $(KUBEBUILDER_PATH) -a -d $(KUBEBUILDER_DIR) ]; then echo "Can not install Kubebuilder since directory $(KUBEBUILDER_DIR) already exists. Please remove/rename it and try again" ; false ; fi
	@if [ ! -e $(KUBEBUILDER_PATH) ]; then \
		echo "Installing Kubebuilder $(CI_KUBEBUILDER_VERSION) ..." \
		&& set -x \
		&& $(CURL_DOWNLOAD) https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(CI_KUBEBUILDER_VERSION)/kubebuilder_$(CI_KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH).tar.gz | tar -xz -C /tmp/ \
		&& mkdir -p $(KUBEBUILDER_DIR) \
		&& cp -r /tmp/kubebuilder_$(CI_KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH)/* $(KUBEBUILDER_DIR) \
		&& rm -rf /tmp/kubebuilder_$(CI_KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) \
		&& for tool in $$( ls $(KUBEBUILDER_DIR)/bin ) ; do if [ ! -e $(CI_TOOLS_DIR)/$${tool} ]; then ln -s $(KUBEBUILDER_DIR)/bin/$${tool} $(CI_TOOLS_DIR)/$${tool} ; echo "Installed $(CI_TOOLS_DIR)/$${tool}" ; else echo "$(CI_TOOLS_DIR)/$${tool} already exists" ; fi; done \
		&& set +x \
		&& echo "Kubebuilder $(CI_KUBEBUILDER_VERSION) has been installed at $(KUBEBUILDER_PATH)" ; fi

.PHONY: dev/install/kustomize
dev/install/kustomize: ## Bootstrap: Install Kustomize
	# see https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/
	@if [ -e $(KUSTOMIZE_PATH) ]; then echo "Kustomize $$( $(KUSTOMIZE_PATH) version ) is already installed at $(KUSTOMIZE_PATH)" ; fi
	@if [ ! -e $(KUSTOMIZE_PATH) ]; then \
		echo "Installing Kustomize $(KUSTOMIZE_VERSION) ..." \
		&& set -x \
		&& mkdir -p $(KUBEBUILDER_DIR)/bin \
		&& $(CURL_DOWNLOAD) https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2F$(KUSTOMIZE_VERSION)/kustomize_$(KUSTOMIZE_VERSION)_$(GOOS)_$(GOARCH).tar.gz | tar -xz -C $(KUBEBUILDER_DIR)/bin \
		&& chmod +x $(KUBEBUILDER_DIR)/bin/kustomize \
		&& ln -s $(KUBEBUILDER_DIR)/bin/kustomize $(KUSTOMIZE_PATH) \
		&& set +x \
		&& echo "Kustomize latest has been installed at $(KUSTOMIZE_PATH)" ; fi

.PHONY: dev/install/kubectl
dev/install/kubectl: ## Bootstrap: Install kubectl
	# see https://kubernetes.io/docs/tasks/tools/install-kubectl/#install-kubectl-binary-with-curl-on-linux
	@if [ -e $(KUBECTL_PATH) ]; then echo "Kubectl $$( $(KUBECTL_PATH) version ) is already installed at $(KUBECTL_PATH)" ; fi
	@if [ ! -e $(KUBECTL_PATH) ]; then \
		echo "Installing Kubectl $(CI_KUBECTL_VERSION) ..." \
		&& set -x \
		&& $(CURL_DOWNLOAD) -O https://storage.googleapis.com/kubernetes-release/release/$(CI_KUBECTL_VERSION)/bin/$(GOOS)/$(GOARCH)/kubectl \
		&& chmod +x kubectl \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv kubectl $(KUBECTL_PATH) \
		&& set +x \
		&& echo "Kubectl $(CI_KUBECTL_VERSION) has been installed at $(KUBECTL_PATH)" ; fi

.PHONY: dev/install/kind
dev/install/kind: ## Bootstrap: Install KIND (Kubernetes in Docker)
	# see https://kind.sigs.k8s.io/docs/user/quick-start/#installation
	@if [ -e $(KIND_PATH) ]; then echo "Kind $$( $(KIND_PATH) version ) is already installed at $(KIND_PATH)" ; fi
	@if [ ! -e $(KIND_PATH) ]; then \
		echo "Installing Kind $(CI_KIND_VERSION) ..." \
		&& set -x \
		&& $(CURL_DOWNLOAD) -o kind https://github.com/kubernetes-sigs/kind/releases/download/$(CI_KIND_VERSION)/kind-$(GOOS)-$(GOARCH) \
		&& chmod +x kind \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv kind $(KIND_PATH) \
		&& set +x \
		&& echo "Kind $(CI_KIND_VERSION) has been installed at $(KIND_PATH)" ; fi

.PHONY: dev/install/k3d
dev/install/k3d: ## Bootstrap: Install K3D (K3s in Docker)
	# see https://raw.githubusercontent.com/rancher/k3d/main/install.sh
	@if [ ! -e $(CI_TOOLS_DIR)/k3d ] || [ `$(CI_TOOLS_DIR)/k3d version | head -1 | awk '{ print $$3 }'` != "$(CI_K3D_VERSION)" ]; then \
		echo "Installing K3d $(CI_K3D_VERSION) ..." \
		&& set -x \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& $(CURL_DOWNLOAD) https://raw.githubusercontent.com/rancher/k3d/main/install.sh | \
		        TAG=$(CI_K3D_VERSION) USE_SUDO="false" K3D_INSTALL_DIR="$(CI_TOOLS_DIR)" bash \
		&& set +x \
		&& echo "K3d $(CI_K3D_VERSION) has been installed at $(CI_TOOLS_DIR)/k3d" ; \
	else echo "K3d version: \"$$( $(CI_TOOLS_DIR)/k3d version )\" is already installed at $(CI_TOOLS_DIR)/k3d"; fi


.PHONY: dev/install/minikube
dev/install/minikube: ## Bootstrap: Install Minikube
	# see https://kubernetes.io/docs/tasks/tools/install-minikube/#linux
	@if [ -e $(MINIKUBE_PATH) ]; then echo "Minikube $$( $(MINIKUBE_PATH) version ) is already installed at $(MINIKUBE_PATH)" ; fi
	@if [ ! -e $(MINIKUBE_PATH) ]; then \
		echo "Installing Minikube $(CI_MINIKUBE_VERSION) ..." \
		&& set -x \
		&& $(CURL_DOWNLOAD) -o minikube https://github.com/kubernetes/minikube/releases/download/$(CI_MINIKUBE_VERSION)/minikube-$(GOOS)-$(GOARCH) \
		&& chmod +x minikube \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv minikube $(MINIKUBE_PATH) \
		&& set +x \
		&& echo "Minikube $(CI_MINIKUBE_VERSION) has been installed at $(MINIKUBE_PATH)" ; fi

.PHONY: dev/install/golangci-lint
dev/install/golangci-lint: ## Bootstrap: Install golangci-lint
	$(CURL_DOWNLOAD) https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOLANGCI_LINT_DIR) $(GOLANGCI_LINT_VERSION)

SHELLCHECK_ARCHIVE := "shellcheck-$(SHELLCHECK_VERSION).$(SHELLCHECK_OS).$(UNAME_ARCH).tar.xz"

.PHONY: dev/install/shellcheck
dev/install/shellcheck:
	@if [ -e $(SHELLCHECK_PATH) ]; then echo "Shellcheck $$( $(SHELLCHECK_PATH) --version ) is already installed at $(SHELLCHECK_PATH)" ; fi
	@if [ ! -e $(SHELLCHECK_PATH) ]; then \
		echo "Installing shellcheck $(SHELLCHECK_VERSION) ..." \
		&& set -x \
		&& $(CURL_DOWNLOAD) -o shellcheck.tar.xz https://github.com/koalaman/shellcheck/releases/download/$(SHELLCHECK_VERSION)/$(SHELLCHECK_ARCHIVE) \
		&& tar -xf shellcheck.tar.xz shellcheck-$(SHELLCHECK_VERSION)/shellcheck \
		&& rm shellcheck.tar.xz \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv shellcheck-$(SHELLCHECK_VERSION)/shellcheck $(SHELLCHECK_PATH) \
		&& chmod +x $(SHELLCHECK_PATH) \
		&& rmdir shellcheck-$(SHELLCHECK_VERSION) \
		&& set +x \
		&& echo "Shellcheck $(SHELLCHECK_VERSION) has been installed at $(SHELLCHECK_PATH)" ; fi

.PHONY: dev/install/helm3
dev/install/helm3: ## Bootstrap: Install Helm 3
	$(CURL_DOWNLOAD) https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | \
		env HELM_INSTALL_DIR=$(CI_TOOLS_DIR) USE_SUDO=false bash

.PHONY: dev/install/helm-docs
dev/install/helm-docs: ## Bootstrap: Install helm-docs
	@if [ -e $(HELM_DOCS_PATH) ]; then echo "Helm Docs $$( $(HELM_DOCS_PATH) --version ) is already installed at $(HELM_DOCS_PATH)" ; fi
	@if [ ! -e $(HELM_DOCS_PATH) ]; then \
		echo "Installing helm-docs ...." \
		&& set -x \
		&& $(CURL_DOWNLOAD) -o helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(HELM_DOCS_ARCH).tar.gz https://github.com/norwoodj/helm-docs/releases/download/v$(HELM_DOCS_VERSION)/helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(HELM_DOCS_ARCH).tar.gz \
		&& tar -xf helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(HELM_DOCS_ARCH).tar.gz helm-docs \
		&& rm helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(HELM_DOCS_ARCH).tar.gz \
		&& chmod +x helm-docs \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv helm-docs $(HELM_DOCS_PATH) \
		&& set +x \
		&& echo "helm-docs $(HELM_DOCS_VERSION) has been installed at $(HELM_DOCS_PATH)" ; fi

GEN_CHANGELOG_START_TAG ?= 0.7.1
GEN_CHANGELOG_BRANCH ?= $(shell git branch --show-current)
GEN_CHANGELOG_MD ?= $(TOP)/changelog.generated.md
GEN_CHANGELOG_REPO ?= https://github.com/kumahq/kuma.git
.PHONY: changelog
changelog:
	@cd $(TOOLS_DIR)/releases/changelog/ && \
		go run ./... \
		--repo $(GEN_CHANGELOG_REPO) \
		--start refs/heads/$(GEN_CHANGELOG_START_TAG) \
		--branch refs/heads/$(GEN_CHANGELOG_BRANCH) > $(GEN_CHANGELOG_MD)
	@echo "The generated changelog is in $(GEN_CHANGELOG_MD)"

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
	@echo 'export KUBEBUILDER_ASSETS=$${CI_TOOLS_DIR}' >> .envrc
	@direnv allow
