.PHONY: help build run curl/listeners curl/clusters run/example/envoy config_dump/example/envoy image \
        run/example/docker-compose curl/example/docker-compose stats/example/docker-compose

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

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run:  ## Run Control Plane locally
	go run ./app/konvoy-cp/main.go run --grpc-port=$(CP_GRPC_PORT) --http-port=$(CP_HTTP_PORT)

curl/listeners: ## Make Discovery request to LDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:listeners --data-binary $(SIMPLE_DISCOVERY_REQUEST)

curl/clusters: ## Make Discovery request to CDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:clusters --data-binary $(SIMPLE_DISCOVERY_REQUEST)

run/example/envoy: ## Run Envoy configured against local Control Plane
	$(ENVOY_BINARY) -c $(EXAMPLE_ENVOY_CONFIG) --config-yaml "node: {metadata: {'IPS': '$(EXAMPLE_ENVOY_IP)', 'PORTS': '$(EXAMPLE_ENVOY_PORT)'}}"

config_dump/example/envoy: ## Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump

image: ## Build Control Plane Docker image
	docker build -t $(CP_DOCKER_IMAGE) .

run/example/docker-compose: ## Run example setup inside Docker Compose
	cd examples/docker-compose && docker-compose pull && docker-compose up

curl/example/docker-compose: ## Make sample requests into Docker Compose setup
	docker run --network docker-compose_envoymesh --rm -ti tutum/curl sh -c 'while true ; do curl http://app:8080 && sleep 1 ; done'

stats/example/docker-compose: ## Observe Envoy metrics from Docker Compose setup
	cd examples/docker-compose && watch 'docker-compose exec envoy curl -s localhost:9901/stats | grep upstream_rq_total'
