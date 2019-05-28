.PHONY: help build run curl/listeners curl/clusters

BUILD_DIR ?= build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts

BIND_HOST ?= localhost
GRPC_PORT ?= 5678
HTTP_PORT ?= 5679
SIMPLE_DISCOVERY_REQUEST := '{"node": {"id": "example"}}'

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run:  ## Run Control Plane locally
	@go run ./app/konvoy-cp/main.go run --grpc-port=$(GRPC_PORT) --http-port=$(HTTP_PORT)

curl/listeners: ## Make Discovery request to LDS
	curl -s $(BIND_HOST):$(HTTP_PORT)/v2/discovery:listeners --data-binary $(SIMPLE_DISCOVERY_REQUEST)

curl/clusters: ## Make Discovery request to CDS
	curl -s $(BIND_HOST):$(HTTP_PORT)/v2/discovery:clusters --data-binary $(SIMPLE_DISCOVERY_REQUEST)
