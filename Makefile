.PHONY: help clean clean/build clean/proto \
		dev/tools dev/tools/all \
		dev/install/protoc dev/install/protoc-gen-gogofast dev/install/protoc-gen-validate \
		dev/install/ginkgo \
		dev/install/kubebuilder dev/install/kustomize \
		dev/install/kubectl dev/install/kind dev/install/minikube \
		start/k8s start/kind start/control-plane/k8s \
		deploy/example-app/k8s deploy/control-plane/k8s \
		kind/load/control-plane kind/load/kuma-dp kind/load/kuma-injector \
		generate protoc/pkg/config/app/kumactl/v1alpha1 generate/kumactl/install/control-plane \
		fmt fmt/go fmt/proto vet check test integration build run/k8s run/universal/memory run/universal/postgres \
		images image/kuma-cp image/kuma-dp image/kumactl image/kuma-injector image/kuma-tcp-echo \
		docker/build docker/build/kuma-cp docker/build/kuma-dp docker/build/kumactl docker/build/kuma-injector docker/build/kuma-tcp-echo \
		docker/save docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-injector docker/save/kuma-tcp-echo \
		docker/load docker/load/kuma-cp docker/load/kuma-dp docker/load/kumactl docker/load/kuma-injector docker/load/kuma-tcp-echo \
		build/kuma-cp build/kuma-dp build/kumactl build/kuma-injector build/kuma-tcp-echo \
		build/kuma-cp/linux-amd64 build/kuma-dp/linux-amd64 build/kumactl/linux-amd64 build/kuma-injector/linux-amd64 build/kuma-tcp-echo/linux-amd64 \
		docs _docs_ docs/kumactl \
		run/example/envoy config_dump/example/envoy \
		run/example/docker-compose wait/example/docker-compose curl/example/docker-compose stats/example/docker-compose \
		verify/example/docker-compose/inbound verify/example/docker-compose/outbound verify/example/docker-compose \
		build/example/minikube load/example/minikube deploy/example/minikube wait/example/minikube curl/example/minikube stats/example/minikube \
		verify/example/minikube/inbound verify/example/minikube/outbound verify/example/minikube \
		print/kubebuilder/test_assets \
		generate/test/cert/kuma-injector run/kuma-injector \
		run/kuma-dp

PKG_LIST := ./... ./api/... ./pkg/plugins/resources/k8s/native/...

BUILD_INFO_GIT_TAG ?= $(shell git describe --tags 2>/dev/null || echo unknown)
BUILD_INFO_GIT_COMMIT ?= $(shell git rev-parse HEAD 2>/dev/null || echo unknown)
BUILD_INFO_BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" || echo unknown)
BUILD_INFO_VERSION ?= $(shell prefix=$$(echo $(BUILD_INFO_GIT_TAG) | cut -c 1); if [ "$${prefix}" = "v" ]; then echo $(BUILD_INFO_GIT_TAG) | cut -c 2- ; else echo $(BUILD_INFO_GIT_TAG) ; fi)

build_info_fields := \
	version=$(BUILD_INFO_VERSION) \
	gitTag=$(BUILD_INFO_GIT_TAG) \
	gitCommit=$(BUILD_INFO_GIT_COMMIT) \
	buildDate=$(BUILD_INFO_BUILD_DATE)
build_info_ld_flags := $(foreach entry,$(build_info_fields), -X github.com/Kong/kuma/pkg/version.$(entry))

LD_FLAGS := -ldflags="-s -w $(build_info_ld_flags)"
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
GO_BUILD := GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build -v $(LD_FLAGS)
GO_RUN := CGO_ENABLED=0 go run $(LD_FLAGS)
GO_TEST := go test $(LD_FLAGS)

BUILD_DIR ?= build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts-${GOOS}-${GOARCH}
BUILD_DOCKER_IMAGES_DIR ?= $(BUILD_DIR)/docker-images

GO_TEST_OPTS ?=

BUILD_COVERAGE_DIR ?= $(BUILD_DIR)/coverage

COVERAGE_PROFILE := $(BUILD_COVERAGE_DIR)/coverage.out
COVERAGE_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage.html

COVERAGE_INTEGRATION_PROFILE := $(BUILD_COVERAGE_DIR)/coverage-integration.out
COVERAGE_INTEGRATION_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage-integration.html

CP_BIND_HOST ?= localhost
CP_GRPC_PORT ?= 5678
SDS_GRPC_PORT ?= 5677
CP_K8S_ADMISSION_PORT ?= 5443

LOCAL_IP ?= $(shell ifconfig en0 | grep 'inet ' | awk '{print $$2}')

ENVOY_BINARY ?= envoy
EXAMPLE_DATAPLANE_MESH ?= default
EXAMPLE_DATAPLANE_NAME ?= example
EXAMPLE_ENVOY_ID ?= $(EXAMPLE_DATAPLANE_MESH).$(EXAMPLE_DATAPLANE_NAME)
EXAMPLE_ENVOY_IP ?= $(LOCAL_IP)
EXAMPLE_ENVOY_PORT ?= 8080
ENVOY_ADMIN_PORT ?= 9901

EXAMPLE_NAMESPACE ?= kuma-demo

KIND_KUBECONFIG = $(shell kind get kubeconfig-path --name=kuma)

define KIND_EXAMPLE_DATAPLANE_MESH
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_MESH)
endef
define KIND_EXAMPLE_DATAPLANE_NAME
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_NAME)
endef

SIMPLE_DISCOVERY_REQUEST ?= '{"node": {"id": "$(EXAMPLE_ENVOY_ID)", "metadata": {"IPS": "$(EXAMPLE_ENVOY_IP)", "PORTS": "$(EXAMPLE_ENVOY_PORT)"}}}'

KUMA_VERSION ?= master

BINTRAY_REGISTRY ?= kong-docker-konvoy-docker.bintray.io
BINTRAY_USERNAME ?=
BINTRAY_API_KEY ?=

KUMA_CP_DOCKER_IMAGE_NAME ?= kuma/kuma-cp
KUMA_DP_DOCKER_IMAGE_NAME ?= kuma/kuma-dp
KUMACTL_DOCKER_IMAGE_NAME ?= kuma/kumactl
KUMA_INJECTOR_DOCKER_IMAGE_NAME ?= kuma/kuma-injector
KUMA_TCP_ECHO_DOCKER_IMAGE_NAME ?= kuma/kuma-tcp-echo

KUMA_CP_DOCKER_IMAGE ?= $(KUMA_CP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
KUMA_DP_DOCKER_IMAGE ?= $(KUMA_DP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
KUMACTL_DOCKER_IMAGE ?= $(KUMACTL_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
KUMA_INJECTOR_DOCKER_IMAGE ?= $(KUMA_INJECTOR_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
KUMA_TCP_ECHO_DOCKER_IMAGE ?= $(KUMA_TCP_ECHO_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)

KUMACTL_INSTALL_USE_LOCAL_IMAGES ?= yes
ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),yes)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-image=$(KUMA_CP_DOCKER_IMAGE_NAME) --dataplane-image=$(KUMA_DP_DOCKER_IMAGE_NAME) --injector-image=$(KUMA_INJECTOR_DOCKER_IMAGE_NAME)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif

PROTOC_VERSION := 3.6.1
PROTOC_PGV_VERSION := v0.1.0
GOGO_PROTOBUF_VERSION := v1.2.1

CI_KUBEBUILDER_VERSION ?= 2.0.0
CI_KIND_VERSION ?= v0.5.1
CI_MINIKUBE_VERSION ?= v1.4.0
CI_KUBERNETES_VERSION ?= v1.15.3
CI_KUBECTL_VERSION ?= v1.14.0
CI_TOOLS_IMAGE ?= circleci/golang:1.12.9

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
BUILD_KUMACTL_DIR := ${BUILD_ARTIFACTS_DIR}/kumactl
export PATH := $(BUILD_KUMACTL_DIR):$(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

PROTOC_PATH := $(CI_TOOLS_DIR)/protoc
KUBEBUILDER_DIR := $(CI_TOOLS_DIR)/kubebuilder.d
KUBEBUILDER_PATH := $(CI_TOOLS_DIR)/kubebuilder
KUSTOMIZE_PATH := $(CI_TOOLS_DIR)/kustomize
KIND_PATH := $(CI_TOOLS_DIR)/kind
MINIKUBE_PATH := $(CI_TOOLS_DIR)/minikube
KUBECTL_PATH := $(CI_TOOLS_DIR)/kubectl
KUBE_APISERVER_PATH := $(CI_TOOLS_DIR)/kube-apiserver
ETCD_PATH := $(CI_TOOLS_DIR)/etcd

PROTO_DIR := ./pkg/config

protoc_search_go_packages := \
	github.com/gogo/protobuf@$(GOGO_PROTOBUF_VERSION)/protobuf \
	github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION) \

protoc_search_go_paths := $(foreach go_package,$(protoc_search_go_packages),--proto_path=$(GOPATH_DIR)/pkg/mod/$(go_package))

# Protobuf-specifc configuration
PROTOC_GO := protoc \
	--proto_path=. \
	$(protoc_search_go_paths) \
	--gogofast_out=plugins=grpc:. \
	--validate_out=lang=gogo:.

PROTOC_OS=unknown
PROTOC_ARCH=$(shell uname -m)

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S), Linux)
	PROTOC_OS=linux
else
	ifeq ($(UNAME_S), Darwin)
		PROTOC_OS=osx
	endif
endif

# tools we expect to be pre-installed
CLANG_FORMAT_PATH ?= clang-format

export TEST_ASSET_KUBE_APISERVER=$(KUBE_APISERVER_PATH)
export TEST_ASSET_ETCD=$(ETCD_PATH)
export TEST_ASSET_KUBECTL=$(KUBECTL_PATH)

DOCKER_COMPOSE_OPTIONS ?=

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev/tools: dev/tools/all ## Bootstrap: Install all development tools

dev/tools/all: dev/install/protoc dev/install/protoc-gen-gogofast dev/install/protoc-gen-validate \
	dev/install/ginkgo \
	dev/install/kubebuilder dev/install/kustomize \
	dev/install/kubectl dev/install/kind dev/install/minikube

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

dev/install/protoc-gen-gogofast: ## Bootstrap: Install Protoc Go Plugin (protobuf Go generator)
	go get -u github.com/gogo/protobuf/protoc-gen-gogofast@$(GOGO_PROTOBUF_VERSION)

dev/install/protoc-gen-validate: ## Bootstrap: Install Protoc Gen Validate Plugin (protobuf validation code generator)
	go get -u github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION)

dev/install/ginkgo: ## Bootstrap: Install Ginkgo (BDD testing framework)
	# see https://github.com/onsi/ginkgo#set-me-up
	echo "Installing Ginkgo ..."
	go get -u github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
	echo "Ginkgo has been installed at $(GOPATH_BIN_DIR)/ginkgo"
	echo "Installing Gomega ..."
	go get -u github.com/onsi/gomega/... # fetches the matcher library
	echo "Gomega has been installed"

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

start/k8s: start/kind deploy/example-app/k8s ## Bootstrap: Start Kubernetes locally (KIND) and deploy sample app

start/kind:
	kind create cluster --name kuma --image=kindest/node:$(CI_KUBERNETES_VERSION) 2>/dev/null || true
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="$$(kind get kubeconfig-path --name=kuma)"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

deploy/example-app/k8s:
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(EXAMPLE_NAMESPACE) || true
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl label namespace $(EXAMPLE_NAMESPACE) kuma.io/sidecar-injection=enabled --overwrite
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -n $(EXAMPLE_NAMESPACE) -f examples/local/demo-app.yaml
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=120s --for=condition=Available -n $(EXAMPLE_NAMESPACE) deployment/demo-app
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(EXAMPLE_NAMESPACE) pods -l app=demo-app

kind/load/control-plane: image/kuma-cp
	kind load docker-image $(KUMA_CP_DOCKER_IMAGE) --name=kuma

kind/load/kuma-dp: image/kuma-dp
	kind load docker-image $(KUMA_DP_DOCKER_IMAGE) --name=kuma

kind/load/kuma-injector: image/kuma-injector
	kind load docker-image $(KUMA_INJECTOR_DOCKER_IMAGE) --name=kuma

deploy/control-plane/k8s: build/kumactl
	kumactl install control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG)  kubectl apply -f -
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete -n kuma-system pod -l app=kuma-injector
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n kuma-system deployment/kuma-injector
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n kuma-system pods -l app=kuma-injector
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete -n $(EXAMPLE_NAMESPACE) pod -l app=demo-app
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(EXAMPLE_NAMESPACE) pods -l app=demo-app

start/control-plane/k8s: kind/load/control-plane kind/load/kuma-dp kind/load/kuma-injector deploy/control-plane/k8s ## Bootstrap: Deploy Control Plane on Kubernetes (KIND)

clean: clean/build ## Dev: Clean

clean/build: ## Dev: Remove build/ dir
	rm -rf "$(BUILD_DIR)"

clean/proto: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIR) -name '*.pb.go' -delete
	find $(PROTO_DIR) -name '*.pb.validate.go' -delete

generate: clean/proto protoc/pkg/config/app/kumactl/v1alpha1 ## Dev: Run code generators

protoc/pkg/config/app/kumactl/v1alpha1:
	$(PROTOC_GO) pkg/config/app/kumactl/v1alpha1/*.proto

# Notice that this command is not include into `make generate` by intention (since generated code differes between dev host and ci server)
generate/kumactl/install/control-plane:
	go generate ./app/kumactl/pkg/install/k8s/control-plane/...
	go generate ./app/kumactl/pkg/install/universal/control-plane/postgres/...

fmt: fmt/go fmt/proto ## Dev: Run various format tools

fmt/go: ## Dev: Run go fmt
	go fmt ./...
	@# apparently, it's not possible to simply use `go fmt ./pkg/plugins/resources/k8s/native/...`
	make fmt -C pkg/plugins/resources/k8s/native

fmt/proto: ## Dev: Run clang-format on .proto files
	which $(CLANG_FORMAT_PATH) && find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT_PATH) -i || true

vet: ## Dev: Run go vet
	go vet ./...
	@# for consistency with `fmt`
	make vet -C pkg/plugins/resources/k8s/native

check: generate fmt vet docs ## Dev: Run code checks (go fmt, go vet, ...)
	make generate manifests -C pkg/plugins/resources/k8s/native
	git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

test: ## Dev: Run tests
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" $(PKG_LIST)
	go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

test/kuma-cp: PKG_LIST=./app/kuma-cp/... ./pkg/config/app/kuma-cp/...
test/kuma-cp: test ## Dev: Run `kuma-cp` tests only

test/kuma-dp: PKG_LIST=./app/kuma-dp/... ./pkg/config/app/kuma-dp/...
test/kuma-dp: test ## Dev: Run `kuma-dp` tests only

test/kumactl: PKG_LIST=./app/kumactl/... ./pkg/config/app/kumactl/...
test/kumactl: test ## Dev: Run `kumactl` tests only

test/kuma-injector: PKG_LIST=./app/kuma-injector/... ./pkg/config/app/kuma-injector/...
test/kuma-injector: test ## Dev: Run 'kuma injector' tests only

integration: ## Dev: Run integration tests
	mkdir -p "$(shell dirname "$(COVERAGE_INTEGRATION_PROFILE)")"
	tools/test/run-integration-tests.sh '$(GO_TEST) -race -covermode=atomic -tags=integration -count=1 -coverpkg=./... -coverprofile=$(COVERAGE_INTEGRATION_PROFILE) $(PKG_LIST)'
	go tool cover -html="$(COVERAGE_INTEGRATION_PROFILE)" -o "$(COVERAGE_INTEGRATION_REPORT_HTML)"

build: build/kuma-cp build/kuma-dp build/kumactl build/kuma-injector build/kuma-tcp-echo ## Dev: Build all binaries

build/kuma-cp: ## Dev: Build `Control Plane` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-cp/kuma-cp ./app/kuma-cp

build/kuma-dp: ## Dev: Build `kuma-dp` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp ./app/kuma-dp

build/kumactl: ## Dev: Build `kumactl` binary
	$(GO_BUILD) -o $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl ./app/kumactl

build/kuma-injector: ## Dev: Build `kuma-injector` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-injector/kuma-injector ./app/kuma-injector

build/kuma-tcp-echo: ## Dev: Build `kuma-tcp-echo` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-tcp-echo/kuma-tcp-echo ./app/kuma-tcp-echo/main.go

run/k8s: fmt vet ## Dev: Run Control Plane locally in Kubernetes mode
	KUBECONFIG=$(KIND_KUBECONFIG) make crd/upgrade -C pkg/plugins/resources/k8s/native
	KUBECONFIG=$(KIND_KUBECONFIG) \
	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=kubernetes \
	KUMA_STORE_TYPE=kubernetes \
	KUMA_SDS_SERVER_TLS_CERT_FILE=app/kuma-injector/cmd/testdata/tls.crt \
	KUMA_SDS_SERVER_TLS_KEY_FILE=app/kuma-injector/cmd/testdata/tls.key \
	KUMA_KUBERNETES_ADMISSION_SERVER_PORT=$(CP_K8S_ADMISSION_PORT) \
	KUMA_KUBERNETES_ADMISSION_SERVER_CERT_DIR=app/kuma-injector/cmd/testdata \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

run/universal/memory: fmt vet ## Dev: Run Control Plane locally in universal mode with in-memory store
	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=memory \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

start/postgres: ## Boostrap: start Postgres for Control Plane with initial schema
	docker-compose -f tools/postgres/docker-compose.yaml up $(DOCKER_COMPOSE_OPTIONS) -d
	tools/postgres/wait-for-postgres.sh 15432

run/universal/postgres: fmt vet ## Dev: Run Control Plane locally in universal mode with Postgres store
	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=postgres \
	KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

run/example/envoy/k8s: EXAMPLE_DATAPLANE_MESH=$(KIND_EXAMPLE_DATAPLANE_MESH)
run/example/envoy/k8s: EXAMPLE_DATAPLANE_NAME=$(KIND_EXAMPLE_DATAPLANE_NAME)
run/example/envoy/k8s: run/example/envoy

run/example/envoy/universal: run/example/envoy

run/example/envoy: build/kuma-dp ## Dev: Run Envoy configured against local Control Plane
	KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_URL=http://localhost:5682 \
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp run --log-level=debug

config_dump/example/envoy: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

images: image/kuma-cp image/kuma-dp image/kumactl image/kuma-injector image/kuma-tcp-echo ## Dev: Build all Docker images

build/kuma-cp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-cp

build/kuma-dp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-dp

build/kumactl/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kumactl

build/kuma-injector/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-injector

build/kuma-tcp-echo/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-tcp-echo

docker/build: docker/build/kuma-cp docker/build/kuma-dp docker/build/kumactl docker/build/kuma-injector docker/build/kuma-tcp-echo

docker/build/kuma-cp: build/artifacts-linux-amd64/kuma-cp/kuma-cp
	docker build -t $(KUMA_CP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-cp .

docker/build/kuma-dp: build/artifacts-linux-amd64/kuma-dp/kuma-dp
	docker build -t $(KUMA_DP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-dp .

docker/build/kumactl: build/artifacts-linux-amd64/kumactl/kumactl
	docker build -t $(KUMACTL_DOCKER_IMAGE) -f tools/ci/dockerfiles/Dockerfile.kumactl .

docker/build/kuma-injector: build/artifacts-linux-amd64/kuma-injector/kuma-injector
	docker build -t $(KUMA_INJECTOR_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-injector .

docker/build/kuma-tcp-echo: build/artifacts-linux-amd64/kuma-tcp-echo/kuma-tcp-echo
	docker build -t $(KUMA_TCP_ECHO_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-tcp-echo .

image/kuma-cp: build/kuma-cp/linux-amd64 docker/build/kuma-cp ## Dev: Build `kuma-cp` Docker image

image/kuma-dp: build/kuma-dp/linux-amd64 docker/build/kuma-dp ## Dev: Build `kuma-dp` Docker image

image/kumactl: build/kumactl/linux-amd64 docker/build/kumactl ## Dev: Build `kumactl` Docker image

image/kuma-injector: build/kuma-injector/linux-amd64 docker/build/kuma-injector ## Dev: Build `kuma-injector` Docker image

image/kuma-tcp-echo: build/kuma-tcp-echo/linux-amd64 docker/build/kuma-tcp-echo ## Dev: Build `kuma-tcp-echo` Docker image

${BUILD_DOCKER_IMAGES_DIR}:
	mkdir -p ${BUILD_DOCKER_IMAGES_DIR}

docker/save: docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-injector docker/save/kuma-tcp-echo

docker/save/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar $(KUMA_CP_DOCKER_IMAGE)

docker/save/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar $(KUMA_DP_DOCKER_IMAGE)

docker/save/kumactl: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar $(KUMACTL_DOCKER_IMAGE)

docker/save/kuma-injector: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-injector.tar $(KUMA_INJECTOR_DOCKER_IMAGE)

docker/save/kuma-tcp-echo: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-tcp-echo.tar $(KUMA_TCP_ECHO_DOCKER_IMAGE)

docker/load: docker/load/kuma-cp docker/load/kuma-dp docker/load/kumactl docker/load/kuma-injector docker/load/kuma-tcp-echo

docker/load/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar

docker/load/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar

docker/load/kumactl: ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar

docker/load/kuma-injector: ${BUILD_DOCKER_IMAGES_DIR}/kuma-injector.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-injector.tar

docker/load/kuma-tcp-echo: ${BUILD_DOCKER_IMAGES_DIR}/kuma-tcp-echo.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-tcp-echo.tar

image/kuma-cp/push: image/kuma-cp
	docker login -u $(BINTRAY_USERNAME) -p $(BINTRAY_API_KEY) $(BINTRAY_REGISTRY)
	docker tag $(KUMA_CP_DOCKER_IMAGE) $(BINTRAY_REGISTRY)/kuma-cp:$(KUMA_VERSION)
	docker push $(BINTRAY_REGISTRY)/kuma-cp:$(KUMA_VERSION)
	docker logout $(BINTRAY_REGISTRY)

image/kuma-dp/push: image/kuma-dp
	docker login -u $(BINTRAY_USERNAME) -p $(BINTRAY_API_KEY) $(BINTRAY_REGISTRY)
	docker tag $(KUMA_DP_DOCKER_IMAGE) $(BINTRAY_REGISTRY)/kuma-dp:$(KUMA_VERSION)
	docker push $(BINTRAY_REGISTRY)/kuma-dp:$(KUMA_VERSION)
	docker logout $(BINTRAY_REGISTRY)

image/kumactl/push: image/kumactl
	docker login -u $(BINTRAY_USERNAME) -p $(BINTRAY_API_KEY) $(BINTRAY_REGISTRY)
	docker tag $(KUMACTL_DOCKER_IMAGE) $(BINTRAY_REGISTRY)/kumactl:$(KUMA_VERSION)
	docker push $(BINTRAY_REGISTRY)/kumactl:$(KUMA_VERSION)
	docker logout $(BINTRAY_REGISTRY)

image/kuma-tcp-echo/push: image/kuma-tcp-echo
	docker login -u $(BINTRAY_USERNAME) -p $(BINTRAY_API_KEY) $(BINTRAY_REGISTRY)
	docker tag $(KUMA_TCP_ECHO_DOCKER_IMAGE) $(BINTRAY_REGISTRY)/kuma-tcp-echo:$(KUMA_VERSION)
	docker push $(BINTRAY_REGISTRY)/kuma-tcp-echo:$(KUMA_VERSION)
	docker logout $(BINTRAY_REGISTRY)

images/push: image/kuma-cp/push image/kuma-dp/push image/kumactl/push image/kuma-tcp-echo/push

docs: ## Dev: Generate all docs
	# re-build `kumactl` binary with a predictable `version`
	$(MAKE) _docs_ BUILD_INFO_VERSION=latest

_docs_: docs/kumactl

docs/kumactl: build/kumactl ## Dev: Generate `kumactl` docs
	tools/docs/kumactl/gen_help.sh ${BUILD_KUMACTL_DIR}/kumactl >docs/cmd/kumactl/HELP.md

print/kubebuilder/test_assets: ## Dev: Print Kubebuilder Environment variables
	@echo export TEST_ASSET_KUBE_APISERVER=$(TEST_ASSET_KUBE_APISERVER)
	@echo export TEST_ASSET_ETCD=$(TEST_ASSET_ETCD)
	@echo export TEST_ASSET_KUBECTL=$(TEST_ASSET_KUBECTL)

run/example/docker-compose: ## Docker Compose: Run demo setup
	docker-compose -f examples/docker-compose/docker-compose.yaml pull
	docker-compose -f examples/docker-compose/docker-compose.yaml up --build --no-start
	docker-compose -f examples/docker-compose/docker-compose.yaml up $(DOCKER_COMPOSE_OPTIONS)

wait/example/docker-compose: ## Docker Compose: Wait for demo setup to get ready
	docker run --network docker-compose_envoymesh --rm -ti $(CI_TOOLS_IMAGE) dockerize -wait http://demo-app:8080 -timeout 1m

curl/example/docker-compose: ## Docker Compose: Make sample requests to demo setup
	docker run --network docker-compose_envoymesh --rm -ti $(CI_TOOLS_IMAGE) sh -c 'set -e ; for i in `seq 1 10`; do test $$(curl -s http://demo-app:8080 | jq -r .url) = "http://mockbin.org/request" && echo "request #$$i successful" ; sleep 1 ; done'

stats/example/docker-compose: ## Docker Compose: Observe Envoy metrics from demo setup
	docker-compose -f examples/docker-compose/docker-compose.yaml exec envoy curl -s localhost:9901/stats/prometheus | grep upstream_rq_total

verify/example/docker-compose/inbound:
	@echo "Checking number of Inbound requests via Envoy ..."
	test $$( docker-compose --file examples/docker-compose/docker-compose.yaml exec envoy curl -s localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8080"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 10
	@echo "Check passed!"

verify/example/docker-compose/outbound:
	@echo "Checking number of Outbound requests via Envoy ..."
	test $$( docker-compose --file examples/docker-compose/docker-compose.yaml exec envoy curl -s localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 10
	@echo "Check passed!"

verify/example/docker-compose: verify/example/docker-compose/inbound verify/example/docker-compose/outbound ## Docker Compose: Verify Envoy stats (after sample requests)

build/example/minikube: ## Minikube: build Docker images inside Minikube
	eval $$(minikube docker-env) && $(MAKE) images

load/example/minikube: ## Minikube: load Docker images into Minikube
	eval $$(minikube docker-env) && $(MAKE) docker/load

deploy/example/minikube: ## Minikube: deploy demo setup
	docker run --rm $(KUMACTL_DOCKER_IMAGE) kumactl install control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | kubectl apply -f -
	kubectl wait --timeout=60s --for=condition=Available -n kuma-system deployment/kuma-injector
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-system pods -l app=kuma-injector
	kubectl apply -f examples/minikube/kuma-demo/
	kubectl wait --timeout=60s --for=condition=Available -n kuma-demo deployment/demo-app
	kubectl wait --timeout=60s --for=condition=Ready -n kuma-demo pods -l app=demo-app

wait/example/minikube: ## Minikube: Wait for demo setup to get ready
	kubectl -n default run wait --rm -ti --restart=Never --image=$(CI_TOOLS_IMAGE) -- dockerize -wait http://demo-app.kuma-demo:8000/request -timeout 1m

curl/example/minikube: ## Minikube: Make sample requests to demo setup
	kubectl -n default run curl --rm -ti --restart=Never --image=$(CI_TOOLS_IMAGE) -- sh -c 'set -e ; for i in `seq 1 10`; do test $$(curl -s http://demo-app.kuma-demo:8000/request | jq -r .url) = "http://mockbin.org/request" && echo "request #$$i successful" ; sleep 1 ; done'

stats/example/minikube: ## Minikube: Observe Envoy metrics from demo setup
	kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep upstream_rq_total

verify/example/minikube/inbound:
	@echo "Checking number of Inbound requests via Envoy ..."
	test $$( kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8000"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 10
	@echo "Check passed!"

verify/example/minikube/outbound:
	@echo "Checking number of Outbound requests via Envoy ..."
	test $$( kubectl -n kuma-demo exec $$(kubectl -n kuma-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 1
	@echo "Check passed!"

verify/example/minikube: verify/example/minikube/inbound verify/example/minikube/outbound ## Minikube: Verify Envoy stats (after sample requests)

kumactl/example/minikube:
	cat examples/minikube/kumactl_workflow.sh | docker run -i --rm --user $$(id -u):$$(id -g) --network host -v $$HOME/.kube:/tmp/.kube -v $$HOME/.minikube:$$HOME/.minikube -e HOME=/tmp -w /tmp $(KUMACTL_DOCKER_IMAGE)

generate/test/cert/kuma-injector:  ## Dev: Generate TLS cert for Kuma Injector (for use in development and unit tests)
	OUTPUT_DIR=$(shell pwd)/app/kuma-injector/cmd/testdata && \
	TMP_DIR=$(shell mktemp -d) && \
	cd $$TMP_DIR && \
	go run $(shell go env GOROOT)/src/crypto/tls/generate_cert.go --host=*.kuma-system.svc,*.kuma-system,localhost --duration=87660h && \
	mv cert.pem $$OUTPUT_DIR/tls.crt && \
	mv key.pem $$OUTPUT_DIR/tls.key && \
	rm -rf $$TMP_DIR

run/kuma-injector: ## Dev: Run Kuma Injector locally
	KUBECONFIG=$(KIND_KUBECONFIG) \
	KUMA_INJECTOR_WEBHOOK_SERVER_CERT_DIR=$(shell pwd)/app/kuma-injector/cmd/testdata \
	$(GO_RUN) ./app/kuma-injector/main.go run --log-level=debug

run/kuma-dp: ## Dev: Run `kuma-dp` locally
	KUMA_CONTROL_PLANE_BOOTSTRAP_SERVER_URL=http://localhost:5682 \
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	$(GO_RUN) ./app/kuma-dp/main.go run --log-level=debug
