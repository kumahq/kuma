.PHONY: help clean clean/build clean/proto \
		start/k8s start/kind start/control-plane/k8s \
		deploy/example-app/k8s deploy/control-plane/k8s \
		kind/load/control-plane kind/load/kuma-dp kind/load/kuma-init \
		generate protoc/plugins protoc/pkg/config/app/kumactl/v1alpha1 protoc/pkg/test/apis/sample/v1alpha1 generate/kumactl/install/k8s/control-plane generate/kumactl/install/k8s/metrics generate/kumactl/install/k8s/tracing generate/kuma-cp/migrations generate/gui \
		fmt fmt/go fmt/proto vet golangci-lint imports check test integration build run/k8s run/universal/memory run/universal/postgres \
		images image/kuma-cp image/kuma-dp image/kumactl image/kuma-init image/kuma-prometheus-sd \
		docker/build docker/build/kuma-cp docker/build/kuma-dp docker/build/kumactl docker/build/kuma-init docker/build/kuma-prometheus-sd \
		docker/save docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-init docker/save/kuma-prometheus-sd \
		docker/load docker/load/kuma-cp docker/load/kuma-dp docker/load/kumactl docker/load/kuma-init docker/load/kuma-prometheus-sd \
		build/kuma-cp build/kuma-dp build/kumactl build/kuma-init build/kuma-prometheus-sd \
		build/kuma-cp/linux-amd64 build/kuma-dp/linux-amd64 build/kumactl/linux-amd64 build/kuma-prometheus-sd/linux-amd64 \
		docs _docs_ docs/kumactl \
		run/example/envoy config_dump/example/envoy \
		print/kubebuilder/test_assets \
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

EXAMPLE_NAMESPACE ?= kuma-example

KIND_KUBECONFIG_DIR ?= $(HOME)/.kube
KIND_KUBECONFIG = $(KIND_KUBECONFIG_DIR)/kind-kuma-config

define KIND_EXAMPLE_DATAPLANE_MESH
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_MESH)
endef
define KIND_EXAMPLE_DATAPLANE_NAME
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_NAME)
endef

SIMPLE_DISCOVERY_REQUEST ?= '{"node": {"id": "$(EXAMPLE_ENVOY_ID)", "metadata": {"IPS": "$(EXAMPLE_ENVOY_IP)", "PORTS": "$(EXAMPLE_ENVOY_PORT)"}}}'

KUMA_VERSION ?= master

BINTRAY_REGISTRY ?= kong-docker-kuma-docker.bintray.io
BINTRAY_USERNAME ?=
BINTRAY_API_KEY ?=

KUMACTL_INSTALL_USE_LOCAL_IMAGES?=true
ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	DOCKER_REGISTRY ?= kuma
else
	DOCKER_REGISTRY ?= $(BINTRAY_REGISTRY)
endif

KUMA_CP_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-cp
KUMA_DP_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-dp
KUMACTL_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kumactl
KUMA_INIT_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-init
KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-prometheus-sd

export KUMA_CP_DOCKER_IMAGE ?= $(KUMA_CP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_DP_DOCKER_IMAGE ?= $(KUMA_DP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMACTL_DOCKER_IMAGE ?= $(KUMACTL_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_INIT_DOCKER_IMAGE ?= $(KUMA_INIT_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_PROMETHEUS_SD_DOCKER_IMAGE ?= $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)

ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-image=$(KUMA_CP_DOCKER_IMAGE_NAME) --dataplane-image=$(KUMA_DP_DOCKER_IMAGE_NAME) --dataplane-init-image=$(KUMA_INIT_DOCKER_IMAGE_NAME)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif
ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_METRICS_IMAGES := --kuma-prometheus-sd-image=$(KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME)
else
	KUMACTL_INSTALL_METRICS_IMAGES :=
endif

PROTOC_VERSION := 3.6.1
PROTOC_PGV_VERSION := v0.3.0-java
GOLANG_PROTOBUF_VERSION := v1.3.2
GOLANGCI_LINT_VERSION := v1.21.0

CI_KUBEBUILDER_VERSION ?= 2.0.0
CI_KIND_VERSION ?= v0.7.0
CI_MINIKUBE_VERSION ?= v1.4.0
CI_KUBERNETES_VERSION ?= v1.15.3
CI_KUBECTL_VERSION ?= v1.14.0
CI_TOOLS_IMAGE ?= circleci/golang:1.12.12

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
BUILD_KUMACTL_DIR := ${BUILD_ARTIFACTS_DIR}/kumactl
export PATH := $(BUILD_KUMACTL_DIR):$(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

PROTOC_PATH := $(CI_TOOLS_DIR)/protoc
PROTOBUF_WKT_DIR := $(CI_TOOLS_DIR)/protobuf.d
KUBEBUILDER_DIR := $(CI_TOOLS_DIR)/kubebuilder.d
KUBEBUILDER_PATH := $(CI_TOOLS_DIR)/kubebuilder
KUSTOMIZE_PATH := $(CI_TOOLS_DIR)/kustomize
KIND_PATH := $(CI_TOOLS_DIR)/kind
MINIKUBE_PATH := $(CI_TOOLS_DIR)/minikube
KUBECTL_PATH := $(CI_TOOLS_DIR)/kubectl
KUBE_APISERVER_PATH := $(CI_TOOLS_DIR)/kube-apiserver
ETCD_PATH := $(CI_TOOLS_DIR)/etcd
GOLANGCI_LINT_DIR := $(CI_TOOLS_DIR)

PROTO_DIR := ./pkg/config

protoc_search_go_packages := \
	github.com/golang/protobuf@$(GOLANG_PROTOBUF_VERSION) \
	github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION) \

protoc_search_go_paths := $(foreach go_package,$(protoc_search_go_packages),--proto_path=$(GOPATH_DIR)/pkg/mod/$(go_package))

# Protobuf-specifc configuration
PROTOC_GO := protoc \
	--proto_path=$(PROTOBUF_WKT_DIR)/include \
	--proto_path=./api \
	--proto_path=. \
	$(protoc_search_go_paths) \
	--go_out=plugins=grpc,Msystem/v1alpha1/datasource.proto=github.com/Kong/kuma/api/system/v1alpha1:. \
	--validate_out=lang=go:.

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

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

start/k8s: start/kind deploy/example-app/k8s ## Bootstrap: Start Kubernetes locally (KIND) and deploy sample app

${KIND_KUBECONFIG_DIR}:
	mkdir -p ${KIND_KUBECONFIG_DIR}

start/kind: ${KIND_KUBECONFIG_DIR}
	kind create cluster --name kuma --image=kindest/node:$(CI_KUBERNETES_VERSION) 2>/dev/null \
	&& kind get kubeconfig --name kuma > $(KIND_KUBECONFIG)
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="${KIND_KUBECONFIG}"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

deploy/example-app/k8s:
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(EXAMPLE_NAMESPACE) || true
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl label namespace $(EXAMPLE_NAMESPACE) kuma.io/sidecar-injection=enabled --overwrite
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/example-app/example-app.yaml
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=120s --for=condition=Available -n $(EXAMPLE_NAMESPACE) deployment/example-app
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(EXAMPLE_NAMESPACE) pods -l app=example-app

kind/load/control-plane: image/kuma-cp
	kind load docker-image $(KUMA_CP_DOCKER_IMAGE) --name=kuma

kind/load/kuma-dp: image/kuma-dp
	kind load docker-image $(KUMA_DP_DOCKER_IMAGE) --name=kuma

kind/load/kuma-init: image/kuma-init
	kind load docker-image $(KUMA_INIT_DOCKER_IMAGE) --name=kuma

kind/load/kuma-prometheus-sd: image/kuma-prometheus-sd
	kind load docker-image $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) --name=kuma

deploy/control-plane/k8s: build/kumactl
	kumactl install control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG)  kubectl apply -f -
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n kuma-system deployment/kuma-control-plane
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n kuma-system pods -l app=kuma-control-plane
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete -n $(EXAMPLE_NAMESPACE) pod -l app=example-app
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(EXAMPLE_NAMESPACE) pods -l app=example-app

start/control-plane/k8s: kind/load/control-plane kind/load/kuma-dp kind/load/kuma-init deploy/control-plane/k8s ## Bootstrap: Deploy Control Plane on Kubernetes (KIND)

clean: clean/build ## Dev: Clean

clean/build: ## Dev: Remove build/ dir
	rm -rf "$(BUILD_DIR)"

clean/proto: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIR) -name '*.pb.go' -delete
	find $(PROTO_DIR) -name '*.pb.validate.go' -delete

generate: clean/proto protoc/pkg/config/app/kumactl/v1alpha1 protoc/pkg/test/apis/sample/v1alpha1 protoc/plugins ## Dev: Run code generators

protoc/pkg/config/app/kumactl/v1alpha1:
	$(PROTOC_GO) pkg/config/app/kumactl/v1alpha1/*.proto

protoc/pkg/test/apis/sample/v1alpha1:
	$(PROTOC_GO) pkg/test/apis/sample/v1alpha1/*.proto

protoc/plugins:
	$(PROTOC_GO) pkg/plugins/ca/provided/config/*.proto

# Notice that this command is not include into `make generate` by intention (since generated code differes between dev host and ci server)
generate/kumactl/install/k8s/control-plane:
	go generate ./app/kumactl/pkg/install/k8s/control-plane/...

# Notice that this command is not include into `make generate` by intention (since generated code differes between dev host and ci server)
generate/kumactl/install/k8s/kuma-cni:
	go generate ./app/kumactl/pkg/install/k8s/kuma-cni/...

# Notice that this command is not include into `make generate` by intention (since generated code differes between dev host and ci server)
generate/kumactl/install/k8s/metrics:
	go generate ./app/kumactl/pkg/install/k8s/metrics/...

# Notice that this command is not include into `make generate` by intention (since generated code differes between dev host and ci server)
generate/kumactl/install/k8s/tracing:
	go generate ./app/kumactl/pkg/install/k8s/tracing/...

generate/kuma-cp/migrations:
	go generate ./pkg/plugins/resources/postgres/migrations/...

generate/gui: ## Generate go files with GUI static files to embed it into binary
	go generate ./app/kuma-ui/pkg/resources/...

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

golangci-lint: ## Dev: Runs golangci-lint linter
	$(GOLANGCI_LINT_DIR)/golangci-lint run -v

imports: ## Dev: Runs goimports in order to organize imports
	goimports -w -local github.com/Kong/kuma -d `find . -type f -name '*.go' -not -name '*.pb.go' -not -path './vendor/*'`

check: generate fmt vet docs golangci-lint imports ## Dev: Run code checks (go fmt, go vet, ...)
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

integration: ## Dev: Run integration tests
	mkdir -p "$(shell dirname "$(COVERAGE_INTEGRATION_PROFILE)")"
	tools/test/run-integration-tests.sh '$(GO_TEST) -race -covermode=atomic -tags=integration -count=1 -coverpkg=./... -coverprofile=$(COVERAGE_INTEGRATION_PROFILE) $(PKG_LIST)'
	go tool cover -html="$(COVERAGE_INTEGRATION_PROFILE)" -o "$(COVERAGE_INTEGRATION_REPORT_HTML)"

build: build/kuma-cp build/kuma-dp build/kumactl build/kuma-prometheus-sd ## Dev: Build all binaries

build/kuma-cp: ## Dev: Build `Control Plane` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-cp/kuma-cp ./app/kuma-cp

build/kuma-dp: ## Dev: Build `kuma-dp` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp ./app/kuma-dp

build/kumactl: ## Dev: Build `kumactl` binary
	$(GO_BUILD) -o $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl ./app/kumactl

build/kuma-prometheus-sd: ## Dev: Build `kuma-prometheus-sd` binary
	$(GO_BUILD) -o ${BUILD_ARTIFACTS_DIR}/kuma-prometheus-sd/kuma-prometheus-sd ./app/kuma-prometheus-sd

run/k8s: fmt vet ## Dev: Run Control Plane locally in Kubernetes mode
	KUBECONFIG=$(KIND_KUBECONFIG) make crd/upgrade -C pkg/plugins/resources/k8s/native
	KUBECONFIG=$(KIND_KUBECONFIG) \
	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=kubernetes \
	KUMA_STORE_TYPE=kubernetes \
	KUMA_SDS_SERVER_TLS_CERT_FILE=app/kuma-cp/cmd/testdata/tls.crt \
	KUMA_SDS_SERVER_TLS_KEY_FILE=app/kuma-cp/cmd/testdata/tls.key \
	KUMA_KUBERNETES_ADMISSION_SERVER_PORT=$(CP_K8S_ADMISSION_PORT) \
	KUMA_KUBERNETES_ADMISSION_SERVER_CERT_DIR=app/kuma-cp/cmd/testdata \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

run/universal/memory: ## Dev: Run Control Plane locally in universal mode with in-memory store
	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=memory \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

start/postgres: ## Boostrap: start Postgres for Control Plane with initial schema
	docker-compose -f tools/postgres/docker-compose.yaml up -d
	tools/postgres/wait-for-postgres.sh 15432

start/postgres/ssl: ## Boostrap: start Postgres for Control Plane with initial schema and SSL enabled
	docker-compose -f tools/postgres/ssl/docker-compose.yaml up -d
	tools/postgres/wait-for-postgres.sh 15432

POSTGRES_SSL_MODE ?= disable

run/universal/postgres/ssl: POSTGRES_SSL_MODE=verifyCa
run/universal/postgres/ssl: POSTGRES_SSL_CERT_PATH=$(shell pwd)/tools/postgres/ssl/certs/postgres.client.crt
run/universal/postgres/ssl: POSTGRES_SSL_KEY_PATH=$(shell pwd)/tools/postgres/ssl/certs/postgres.client.key
run/universal/postgres/ssl: POSTGRES_SSL_ROOT_CERT_PATH=$(shell pwd)/tools/postgres/ssl/certs/rootCA.crt
run/universal/postgres/ssl: run/universal/postgres ## Dev: Run Control Plane locally in universal mode with Postgres store and SSL enabled

run/universal/postgres: fmt vet ## Dev: Run Control Plane locally in universal mode with Postgres store
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=postgres \
	KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma \
	KUMA_STORE_POSTGRES_TLS_MODE=$(POSTGRES_SSL_MODE) \
	KUMA_STORE_POSTGRES_TLS_CERT_PATH=$(POSTGRES_SSL_CERT_PATH) \
	KUMA_STORE_POSTGRES_TLS_KEY_PATH=$(POSTGRES_SSL_KEY_PATH) \
	KUMA_STORE_POSTGRES_TLS_CA_PATH=$(POSTGRES_SSL_ROOT_CERT_PATH) \
	$(GO_RUN) ./app/kuma-cp/main.go migrate up --log-level=debug

	KUMA_SDS_SERVER_GRPC_PORT=$(SDS_GRPC_PORT) \
	KUMA_GRPC_PORT=$(CP_GRPC_PORT) \
	KUMA_ENVIRONMENT=universal \
	KUMA_STORE_TYPE=postgres \
	KUMA_STORE_POSTGRES_HOST=localhost \
	KUMA_STORE_POSTGRES_PORT=15432 \
	KUMA_STORE_POSTGRES_USER=kuma \
	KUMA_STORE_POSTGRES_PASSWORD=kuma \
	KUMA_STORE_POSTGRES_DB_NAME=kuma \
	KUMA_STORE_POSTGRES_TLS_MODE=$(POSTGRES_SSL_MODE) \
	KUMA_STORE_POSTGRES_TLS_CERT_PATH=$(POSTGRES_SSL_CERT_PATH) \
	KUMA_STORE_POSTGRES_TLS_KEY_PATH=$(POSTGRES_SSL_KEY_PATH) \
	KUMA_STORE_POSTGRES_TLS_CA_PATH=$(POSTGRES_SSL_ROOT_CERT_PATH) \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

run/example/envoy/k8s: EXAMPLE_DATAPLANE_MESH=$(KIND_EXAMPLE_DATAPLANE_MESH)
run/example/envoy/k8s: EXAMPLE_DATAPLANE_NAME=$(KIND_EXAMPLE_DATAPLANE_NAME)
run/example/envoy/k8s: run/example/envoy

run/example/envoy/universal: run/example/envoy

run/example/envoy: build/kuma-dp build/kumactl ## Dev: Run Envoy configured against local Control Plane
	${BUILD_ARTIFACTS_DIR}/kumactl/kumactl generate dataplane-token --dataplane=$(EXAMPLE_DATAPLANE_NAME) --mesh=$(EXAMPLE_DATAPLANE_MESH) > /tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	KUMA_DATAPLANE_RUNTIME_TOKEN_PATH=/tmp/kuma-dp-$(EXAMPLE_DATAPLANE_NAME)-$(EXAMPLE_DATAPLANE_MESH)-token \
	${BUILD_ARTIFACTS_DIR}/kuma-dp/kuma-dp run --log-level=debug

config_dump/example/envoy: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

images: image/kuma-cp image/kuma-dp image/kumactl image/kuma-init image/kuma-prometheus-sd ## Dev: Rebuild all Docker images

build/kuma-cp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-cp

build/kuma-dp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-dp

build/kumactl/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kumactl

build/kuma-prometheus-sd/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-prometheus-sd

docker/build: docker/build/kuma-cp docker/build/kuma-dp docker/build/kumactl docker/build/kuma-init docker/build/kuma-prometheus-sd ## Dev: Build all Docker images using existing artifacts from build

docker/build/kuma-cp: build/artifacts-linux-amd64/kuma-cp/kuma-cp ## Dev: Build `kuma-cp` Docker image using existing artifact
	docker build -t $(KUMA_CP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-cp .

docker/build/kuma-dp: build/artifacts-linux-amd64/kuma-dp/kuma-dp ## Dev: Build `kuma-dp` Docker image using existing artifact
	docker build -t $(KUMA_DP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-dp .

docker/build/kumactl: build/artifacts-linux-amd64/kumactl/kumactl ## Dev: Build `kumactl` Docker image using existing artifact
	docker build -t $(KUMACTL_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kumactl .

docker/build/kuma-init: ## Dev: Build `kuma-init` Docker image using existing artifact
	docker build -t $(KUMA_INIT_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-init .

docker/build/kuma-prometheus-sd: build/artifacts-linux-amd64/kuma-prometheus-sd/kuma-prometheus-sd ## Dev: Build `kuma-prometheus-sd` Docker image using existing artifact
	docker build -t $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-prometheus-sd .

image/kuma-cp: build/kuma-cp/linux-amd64 docker/build/kuma-cp ## Dev: Rebuild `kuma-cp` Docker image

image/kuma-dp: build/kuma-dp/linux-amd64 docker/build/kuma-dp ## Dev: Rebuild `kuma-dp` Docker image

image/kumactl: build/kumactl/linux-amd64 docker/build/kumactl ## Dev: Rebuild `kumactl` Docker image

image/kuma-init: docker/build/kuma-init ## Dev: Rebuild `kuma-init` Docker image

image/kuma-prometheus-sd: build/kuma-prometheus-sd/linux-amd64 docker/build/kuma-prometheus-sd ## Dev: Rebuild `kuma-prometheus-sd` Docker image

${BUILD_DOCKER_IMAGES_DIR}:
	mkdir -p ${BUILD_DOCKER_IMAGES_DIR}

docker/save: docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-init docker/save/kuma-prometheus-sd

docker/save/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar $(KUMA_CP_DOCKER_IMAGE)

docker/save/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar $(KUMA_DP_DOCKER_IMAGE)

docker/save/kumactl: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar $(KUMACTL_DOCKER_IMAGE)

docker/save/kuma-init: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar $(KUMA_INIT_DOCKER_IMAGE)

docker/save/kuma-prometheus-sd: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE)

docker/load: docker/load/kuma-cp docker/load/kuma-dp docker/load/kumactl docker/load/kuma-init docker/load/kuma-prometheus-sd

docker/load/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar

docker/load/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar

docker/load/kumactl: ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar

docker/load/kuma-init: ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar

docker/load/kuma-prometheus-sd: ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar

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

images/push: image/kuma-cp/push image/kuma-dp/push image/kumactl/push

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

run/kuma-dp: ## Dev: Run `kuma-dp` locally
	KUMA_DATAPLANE_MESH=$(EXAMPLE_DATAPLANE_MESH) \
	KUMA_DATAPLANE_NAME=$(EXAMPLE_DATAPLANE_NAME) \
	KUMA_DATAPLANE_ADMIN_PORT=$(ENVOY_ADMIN_PORT) \
	$(GO_RUN) ./app/kuma-dp/main.go run --log-level=debug

include Makefile.dev.mk
include Makefile.e2e.mk
