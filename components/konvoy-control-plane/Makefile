.PHONY: help clean \
		dev/tools dev/install/ginkgo dev/install/kubebuilder dev/install/kustomize dev/install/kubectl dev/install/kind dev/install/minikube \
		start/deps \
		fmt fmt/go fmt/proto vet check test build run image \
		curl/listeners curl/clusters \
		run/example/envoy config_dump/example/envoy \
        run/example/docker-compose wait/example/docker-compose curl/example/docker-compose stats/example/docker-compose \
        verify/example/docker-compose/inbound verify/example/docker-compose/outbound verify/example/docker-compose \
		build/example/minikube deploy/example/minikube wait/example/minikube curl/example/minikube stats/example/minikube \
        verify/example/minikube/inbound verify/example/minikube/outbound verify/example/minikube \
		print/kubebuilder/test_assets

PKG_LIST := ./... ./pkg/plugins/resources/k8s/native/...

BUILD_DIR ?= build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts

BUILD_COVERAGE_DIR ?= $(BUILD_DIR)/coverage

COVERAGE_PROFILE := $(BUILD_COVERAGE_DIR)/coverage.out
COVERAGE_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage.html

CP_BIND_HOST ?= localhost
CP_GRPC_PORT ?= 5678
CP_HTTP_PORT ?= 5679

LOCAL_IP ?= $(shell ifconfig en0 | grep 'inet ' | awk '{print $$2}')

ENVOY_BINARY ?= envoy
EXAMPLE_ENVOY_CONFIG ?= $(PWD)/examples/local/konvoy-control-plane.yaml
EXAMPLE_ENVOY_IP ?= $(LOCAL_IP)
EXAMPLE_ENVOY_PORT ?= 8080
ENVOY_ADMIN_PORT ?= 9901

SIMPLE_DISCOVERY_REQUEST ?= '{"node": {"id": "example", "metadata": {"IPS": "$(EXAMPLE_ENVOY_IP)", "PORTS": "$(EXAMPLE_ENVOY_PORT)"}}}'

CP_DOCKER_IMAGE ?= konvoy/konvoy-control-plane:latest

CI_KUBEBUILDER_VERSION ?= 2.0.0-alpha.4
CI_KIND_VERSION ?= v0.3.0
CI_MINIKUBE_VERSION ?= v1.1.0
CI_KUBECTL_VERSION ?= v1.14.0
CI_TOOLS_IMAGE ?= circleci/golang:1.12

CI_TOOLS_DIR ?= $(HOME)/bin
PATH := $(CI_TOOLS_DIR):$(PATH)

KUBEBUILDER_DIR := $(CI_TOOLS_DIR)/kubebuilder.d
KUBEBUILDER_PATH := $(CI_TOOLS_DIR)/kubebuilder
KUSTOMIZE_PATH := $(CI_TOOLS_DIR)/kustomize
KIND_PATH := $(CI_TOOLS_DIR)/kind
MINIKUBE_PATH := $(CI_TOOLS_DIR)/minikube
KUBECTL_PATH := $(CI_TOOLS_DIR)/kubectl
KUBE_APISERVER_PATH := $(CI_TOOLS_DIR)/kube-apiserver
ETCD_PATH := $(CI_TOOLS_DIR)/etcd

# tools we expect to be pre-installed
CLANG_FORMAT_PATH ?= clang-format

export TEST_ASSET_KUBE_APISERVER=$(KUBE_APISERVER_PATH)
export TEST_ASSET_ETCD=$(ETCD_PATH)
export TEST_ASSET_KUBECTL=$(KUBECTL_PATH)

DOCKER_COMPOSE_OPTIONS ?=

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev/tools: dev/install/ginkgo dev/install/kubebuilder dev/install/kustomize dev/install/kubectl dev/install/kind dev/install/minikube ## Bootstrap: Install all development tools

dev/install/ginkgo: ## Bootstrap: Install Ginkgo (BDD testing framework)
	# see https://github.com/onsi/ginkgo#set-me-up
	echo "Installing Ginkgo ..."
	go get -u github.com/onsi/ginkgo/ginkgo  # installs the ginkgo CLI
	echo "Ginkgo has been installed at $(shell go env GOPATH)/bin/ginkgo"
	echo "Installing Gomega ..."
	go get -u github.com/onsi/gomega/...     # fetches the matcher library
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

start/deps: start/kind ## Bootstrap: Start dependencies (etcd + kube-apiserver)

start/kind:
	kind create cluster --name konvoy 2>/dev/null || true
	@echo
	@echo
	@echo "!!! You need to manually run the following command in your shell: !!!"
	@echo
	@echo
	@echo export KUBECONFIG="$$(kind get kubeconfig-path --name=konvoy)"

clean: ## Dev: Clean
	rm -rf "$(BUILD_DIR)"

fmt: fmt/go fmt/proto ## Dev: Run various format tools

fmt/go: ## Dev: Run go fmt
	go fmt ./...
	@# apparently, it's not possible to simply use `go fmt ./pkg/plugins/resources/k8s/native/...`
	make fmt -C pkg/plugins/resources/k8s/native

fmt/proto: ## Dev: Run clang-format on .proto files
	find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT_PATH) -i

vet: ## Dev: Run go vet
	go vet ./...
	@# for consistency with `fmt`
	make vet -C pkg/plugins/resources/k8s/native

check: fmt vet ## Dev: Run code checks (go fmt, go vet, ...)
	make generate manifests -C pkg/plugins/resources/k8s/native
	git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

test: ## Dev: Run tests
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"
	go test -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" $(PKG_LIST)
	go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

build: ## Dev: Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run: fmt vet ## Dev: Run Control Plane locally
	make crd/upgrade -C pkg/plugins/resources/k8s/native
	go run ./app/konvoy-cp/main.go run --grpc-port=$(CP_GRPC_PORT) --http-port=$(CP_HTTP_PORT)

curl/listeners: ## Dev: Make Discovery request to LDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:listeners --data-binary $(SIMPLE_DISCOVERY_REQUEST)

curl/clusters: ## Dev: Make Discovery request to CDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:clusters --data-binary $(SIMPLE_DISCOVERY_REQUEST)

run/example/envoy: ## Dev: Run Envoy configured against local Control Plane
	$(ENVOY_BINARY) -c $(EXAMPLE_ENVOY_CONFIG) --config-yaml "node: {metadata: {'IPS': '$(EXAMPLE_ENVOY_IP)', 'PORTS': '$(EXAMPLE_ENVOY_PORT)'}}"

config_dump/example/envoy: ## Dev: Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

image: ## Dev: Build Control Plane Docker image
	docker build -t $(CP_DOCKER_IMAGE) .

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

build/example/minikube: ## Minikube: build demo setup
	eval $(minikube docker-env)
	make image

deploy/example/minikube: ## Minikube: deploy demo setup
	make crd/install -C pkg/plugins/resources/k8s/native
	kubectl kustomize pkg/plugins/resources/k8s/native/config/rbac | kubectl apply -f -
	kubectl apply -f examples/minikube/konvoy-demo.yaml

wait/example/minikube: ## Minikube: Wait for demo setup to get ready
	kubectl -n konvoy-demo run wait --rm -ti --restart=Never --image=$(CI_TOOLS_IMAGE) -- dockerize -wait http://demo-app:8000/request -timeout 1m

curl/example/minikube: ## Minikube: Make sample requests to demo setup
	kubectl -n konvoy-demo run curl --rm -ti --restart=Never --image=$(CI_TOOLS_IMAGE) -- sh -c 'set -e ; for i in `seq 1 10`; do test $$(curl -s http://demo-app:8000/request | jq -r .url) = "http://mockbin.org/request" && echo "request #$$i successful" ; sleep 1 ; done'

stats/example/minikube: ## Minikube: Observe Envoy metrics from demo setup
	kubectl -n konvoy-demo exec $$(kubectl -n konvoy-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c envoy-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep upstream_rq_total

verify/example/minikube/inbound:
	@echo "Checking number of Inbound requests via Envoy ..."
	test $$( kubectl -n konvoy-demo exec $$(kubectl -n konvoy-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c envoy-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="localhost_8000"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 10
	@echo "Check passed!"

verify/example/minikube/outbound:
	@echo "Checking number of Outbound requests via Envoy ..."
	test $$( kubectl -n konvoy-demo exec $$(kubectl -n konvoy-demo get pods -l app=demo-app -o=jsonpath='{.items[0].metadata.name}') -c envoy-sidecar -- wget -qO- http://localhost:9901/stats/prometheus | grep 'envoy_cluster_upstream_rq_total{envoy_cluster_name="pass_through"}' | awk '{print $$2}' | tr -d [:space:] ) -ge 1
	@echo "Check passed!"

verify/example/minikube: verify/example/minikube/inbound verify/example/minikube/outbound ## Minikube: Verify Envoy stats (after sample requests)
