.PHONY: help build run curl/listeners curl/clusters run/example/envoy config_dump/example/envoy

BUILD_DIR ?= build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts

CP_BIND_HOST ?= localhost
CP_GRPC_PORT ?= 5678
CP_HTTP_PORT ?= 5679
SIMPLE_DISCOVERY_REQUEST := '{"node": {"id": "example"}}'

ENVOY_BINARY ?= envoy
EXAMPLE_ENVOY_CONFIG ?= $(PWD)/examples/envoy/konvoy-control-plane.yaml
ENVOY_ADMIN_PORT ?= 9901

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run:  ## Run Control Plane locally
	@go run ./app/konvoy-cp/main.go run --grpc-port=$(CP_GRPC_PORT) --http-port=$(CP_HTTP_PORT)

curl/listeners: ## Make Discovery request to LDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:listeners --data-binary $(SIMPLE_DISCOVERY_REQUEST)

curl/clusters: ## Make Discovery request to CDS
	curl -s $(CP_BIND_HOST):$(CP_HTTP_PORT)/v2/discovery:clusters --data-binary $(SIMPLE_DISCOVERY_REQUEST)

run/example/envoy: ## Run Envoy configured against local Control Plane
	$(ENVOY_BINARY) -c $(EXAMPLE_ENVOY_CONFIG)

config_dump/example/envoy: ## Dump effective configuration of example Envoy
	curl -s localhost:$(ENVOY_ADMIN_PORT)/config_dump
