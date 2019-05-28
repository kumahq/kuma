.PHONY: help build run

BUILD_DIR := build
BUILD_ARTIFACTS_DIR := ${BUILD_DIR}/artifacts

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build Control Plane binary
	CGO_ENABLED=0 go build -ldflags="-s -w" -v -o ${BUILD_ARTIFACTS_DIR}/konvoy-control-plane/konvoy-control-plane ./app/konvoy-cp

run:  ## Run Control Plane locally
	@go run ./app/konvoy-cp/main.go run
