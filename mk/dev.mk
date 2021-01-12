PROTOC_VERSION := 3.14.0
PROTOC_PGV_VERSION := v0.4.1
GOLANG_PROTOBUF_VERSION := v1.4.3
GOLANGCI_LINT_VERSION := v1.31.0
GINKGO_VERSION := v1.12.0
HELM_DOCS_VERSION := 1.4.0

CI_KUBEBUILDER_VERSION ?= 2.3.1
CI_MINIKUBE_VERSION ?= v1.13.1
CI_KUBECTL_VERSION ?= v1.18.9
CI_TOOLS_IMAGE ?= circleci/golang:1.15.5

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
export PATH := $(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

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

TOOLS_DIR ?= $(shell pwd)/tools

PROTOC_OS=unknown
PROTOC_ARCH=$(shell uname -m)

UNAME_S := $(shell uname -s)
UNAME_ARCH := $(shell uname -m)
ifeq ($(UNAME_S), Linux)
	PROTOC_OS=linux
else
	ifeq ($(UNAME_S), Darwin)
		PROTOC_OS=osx
	endif
endif

.PHONY: dev/tools
dev/tools: dev/tools/all ## Bootstrap: Install all development tools

.PHONY: dev/tools/all
dev/tools/all: dev/install/protoc dev/install/protobuf-wellknown-types \
	dev/install/protoc-gen-go dev/install/protoc-gen-validate \
	dev/install/ginkgo \
	dev/install/kubebuilder dev/install/kustomize \
	dev/install/kubectl dev/install/kind dev/install/minikube \
	dev/install/golangci-lint \
	dev/install/goimports \
	dev/install/helm3 \
	dev/install/helm-docs

.PHONY: dev/install/protoc
dev/install/protoc: ## Bootstrap: Install Protoc (protobuf compiler)
	@if [ -e $(PROTOC_PATH) ]; then echo "Protoc $$( $(PROTOC_PATH) --version ) is already installed at $(PROTOC_PATH)" ; fi
	@if [ ! -e $(PROTOC_PATH) ]; then \
		echo "Installing Protoc $(PROTOC_VERSION) ..." \
		&& set -x \
		&& curl -Lo /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
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
		&& curl -Lo /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip 'include/*' -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(PROTOBUF_WKT_DIR) \
		&& cp -r /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH)/include $(PROTOBUF_WKT_DIR) \
		&& rm -rf /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& rm /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& set +x \
		&& echo "Protobuf well-known types $(PROTOC_VERSION) have been installed at $(PROTOBUF_WKT_DIR)" ; fi

.PHONY: dev/install/protoc-gen-go
dev/install/protoc-gen-go: ## Bootstrap: Install Protoc Go Plugin (protobuf Go generator)
	go get github.com/golang/protobuf/protoc-gen-go@$(GOLANG_PROTOBUF_VERSION)

.PHONY: dev/install/protoc-gen-validate
dev/install/protoc-gen-validate: ## Bootstrap: Install Protoc Gen Validate Plugin (protobuf validation code generator)
	go get github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION)

.PHONY: dev/install/ginkgo
dev/install/ginkgo: ## Bootstrap: Install Ginkgo (BDD testing framework)
	# see https://github.com/onsi/ginkgo#set-me-up
	echo "Installing Ginkgo ..."
	go get github.com/onsi/ginkgo/ginkgo@$(GINKGO_VERSION)  # installs the ginkgo CLI
	echo "Ginkgo has been installed at $(GOPATH_BIN_DIR)/ginkgo"

.PHONY: dev/install/kubebuilder
dev/install/kubebuilder: ## Bootstrap: Install Kubebuilder (including etcd and kube-apiserver)
	# see https://book.kubebuilder.io/quick-start.html#installation
	@if [ -e $(KUBEBUILDER_PATH) ]; then echo "Kubebuilder $$( $(KUBEBUILDER_PATH) version ) is already installed at $(KUBEBUILDER_PATH)" ; fi
	@if [ ! -e $(KUBEBUILDER_PATH) -a -d $(KUBEBUILDER_DIR) ]; then echo "Can not install Kubebuilder since directory $(KUBEBUILDER_DIR) already exists. Please remove/rename it and try again" ; false ; fi
	@if [ ! -e $(KUBEBUILDER_PATH) ]; then \
		echo "Installing Kubebuilder $(CI_KUBEBUILDER_VERSION) ..." \
		&& set -x \
		&& curl -L https://go.kubebuilder.io/dl/$(CI_KUBEBUILDER_VERSION)/$(GOOS)/$(GOARCH) | tar -xz -C /tmp/ \
		&& mkdir -p $(KUBEBUILDER_DIR) \
		&& cp -r /tmp/kubebuilder_$(CI_KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH)/* $(KUBEBUILDER_DIR) \
		&& rm -rf /tmp/kubebuilder_$(CI_KUBEBUILDER_VERSION)_$(GOOS)_$(GOARCH) \
        && for tool in $$( ls $(KUBEBUILDER_DIR)/bin ) ; do if [ ! -e $(CI_TOOLS_DIR)/$${tool} ]; then ln -s $(KUBEBUILDER_DIR)/bin/$${tool} $(CI_TOOLS_DIR)/$${tool} ; echo "Installed $(CI_TOOLS_DIR)/$${tool}" ; else echo "$(CI_TOOLS_DIR)/$${tool} already exists" ; fi; done \
		&& set +x \
		&& echo "Kubebuilder $(CI_KUBEBUILDER_VERSION) has been installed at $(KUBEBUILDER_PATH)" ; fi

.PHONY: dev/install/kustomize
dev/install/kustomize: ## Bootstrap: Install Kustomize
	# see https://book.kubebuilder.io/quick-start.html#installation
	@if [ -e $(KUSTOMIZE_PATH) ]; then echo "Kustomize $$( $(KUSTOMIZE_PATH) version ) is already installed at $(KUSTOMIZE_PATH)" ; fi
	@if [ ! -e $(KUSTOMIZE_PATH) ]; then \
		echo "Installing Kustomize latest ..." \
		&& set -x \
		&& curl -Lo kustomize https://go.kubebuilder.io/kustomize/$(GOOS)/$(GOARCH) \
		&& chmod +x kustomize \
		&& mkdir -p $(KUBEBUILDER_DIR)/bin \
		&& mv kustomize $(KUBEBUILDER_DIR)/bin/ \
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
		&& curl -LO https://storage.googleapis.com/kubernetes-release/release/$(CI_KUBECTL_VERSION)/bin/$(GOOS)/$(GOARCH)/kubectl \
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
		&& curl -Lo kind https://github.com/kubernetes-sigs/kind/releases/download/$(CI_KIND_VERSION)/kind-$(GOOS)-$(GOARCH) \
		&& chmod +x kind \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv kind $(KIND_PATH) \
		&& set +x \
		&& echo "Kind $(CI_KIND_VERSION) has been installed at $(KIND_PATH)" ; fi

.PHONY: dev/install/minikube
dev/install/minikube: ## Bootstrap: Install Minikube
	# see https://kubernetes.io/docs/tasks/tools/install-minikube/#linux
	@if [ -e $(MINIKUBE_PATH) ]; then echo "Minikube $$( $(MINIKUBE_PATH) version ) is already installed at $(MINIKUBE_PATH)" ; fi
	@if [ ! -e $(MINIKUBE_PATH) ]; then \
		echo "Installing Minikube $(CI_MINIKUBE_VERSION) ..." \
		&& set -x \
		&& curl -Lo minikube https://storage.googleapis.com/minikube/releases/$(CI_MINIKUBE_VERSION)/minikube-$(GOOS)-$(GOARCH) \
		&& chmod +x minikube \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv minikube $(MINIKUBE_PATH) \
		&& set +x \
		&& echo "Minikube $(CI_MINIKUBE_VERSION) has been installed at $(MINIKUBE_PATH)" ; fi

.PHONY: dev/install/golangci-lint
dev/install/golangci-lint: ## Bootstrap: Install golangci-lint
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b $(GOLANGCI_LINT_DIR) $(GOLANGCI_LINT_VERSION)

.PHONY: dev/install/goimports
dev/install/goimports: ## Bootstrap: Install goimports
	go get golang.org/x/tools/cmd/goimports

.PHONY: dev/install/helm3
dev/install/helm3: ## Bootstrap: Install Helm 3
	curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

.PHONY: dev/install/helm-docs
dev/install/helm-docs: ## Bootstrap: Install helm-docs
	@if [ -e $(HELM_DOCS_PATH) ]; then echo "Helm Docs $$( $(HELM_DOCS_PATH) --version ) is already installed at $(HELM_DOCS_PATH)" ; fi
	@if [ ! -e $(HELM_DOCS_PATH) ]; then \
		echo "Installing helm-docs ...." \
		&& set -x \
		&& curl -Lo helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(UNAME_ARCH).tar.gz https://github.com/norwoodj/helm-docs/releases/download/v$(HELM_DOCS_VERSION)/helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(UNAME_ARCH).tar.gz \
		&& tar -xf helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(UNAME_ARCH).tar.gz helm-docs \
		&& rm helm-docs_$(HELM_DOCS_VERSION)_$(UNAME_S)_$(UNAME_ARCH).tar.gz \
		&& chmod +x helm-docs \
		&& mkdir -p $(CI_TOOLS_DIR) \
		&& mv helm-docs $(HELM_DOCS_PATH) \
		&& set +x \
		&& echo "helm-docs $(HELM_DOCS_VERSION) has been installed at $(HELM_DOCS_PATH)" ; fi

GEN_CHANGELOG_START_TAG ?= 0.7.2
GEN_CHANGELOG_BRANCH ?= master
GEN_CHANGELOG_MD ?= changelog.generated.md
.PHONY: changelog
changelog:
	@cd tools/releases/changelog/ && \
	go run ./... --branch refs/heads/$(GEN_CHANGELOG_BRANCH) --start refs/heads/$(GEN_CHANGELOG_START_TAG) > ../../../$(GEN_CHANGELOG_MD)
	@echo "The generated changelog is in $(GEN_CHANGELOG_MD)"
