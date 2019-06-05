.PHONY: help fmt vet build run image \
		curl/listeners curl/clusters \
		run/example/envoy config_dump/example/envoy \
        run/example/docker-compose wait/example/docker-compose curl/example/docker-compose stats/example/docker-compose \
        verify/example/docker-compose/inbound verify/example/docker-compose/outbound verify/example/docker-compose \
		build/example/minikube deploy/example/minikube wait/example/minikube curl/example/minikube stats/example/minikube \
        verify/example/minikube/inbound verify/example/minikube/outbound verify/example/minikube \
		ci/install/minikube ci/install/kubectl

BUILD_DIR ?= build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts

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

CI_MINIKUBE_VERSION ?= v1.1.0
CI_KUBECTL_VERSION ?= v1.14.0
CI_TOOLS_IMAGE ?= circleci/golang:1.12

DOCKER_COMPOSE_OPTIONS ?=

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

fmt: ## Dev: Run go fmt
	go fmt ./...

vet: ## Dev: Run go vet
	go vet ./...

check: fmt vet ## Dev: Run code checks (go fmt, go vet, ...)
	git diff --quiet # fail if Git working tree is dirty

build: ## Dev: Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run: fmt vet ## Dev: Run Control Plane locally
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

ci/install/minikube: ## CI: Install Minikube
	curl -Lo minikube https://storage.googleapis.com/minikube/releases/$(CI_MINIKUBE_VERSION)/minikube-linux-amd64 && chmod +x minikube && sudo mv minikube /usr/local/bin/

ci/install/kubectl: ## CI: Install kubectl
	curl -LO https://storage.googleapis.com/kubernetes-release/release/$(CI_KUBECTL_VERSION)/bin/linux/amd64/kubectl && chmod +x kubectl && sudo mv kubectl /usr/local/bin/
