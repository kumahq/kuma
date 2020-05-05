.PHONY: all test manager run install crd/install crd/uninstall crd/upgrade deploy manifests fmt vet generate \
		docker-build docker-push controller-gen

# Image URL to use all building/pushing image targets
IMG ?= kuma/kuma-controller-manager:latest

# TODO(yskopets): CRD generation is broken since introduction of .proto definitions
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
#CRD_OPTIONS ?= "crd:trivialVersions=true"
#CRD_OUTPUT := output:crd:artifacts:config=config/crd/bases

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
export PATH := $(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

KUBE_APISERVER_PATH := $(CI_TOOLS_DIR)/kube-apiserver
ETCD_PATH := $(CI_TOOLS_DIR)/etcd
KUBECTL_PATH := $(CI_TOOLS_DIR)/kubectl

export TEST_ASSET_KUBE_APISERVER=$(KUBE_APISERVER_PATH)
export TEST_ASSET_ETCD=$(ETCD_PATH)
export TEST_ASSET_KUBECTL=$(KUBECTL_PATH)

all: manager

GO_TEST ?= go test
GO_TEST_OPTS ?=
COVERAGE_PROFILE ?= cover.out

# Run tests
test: generate fmt vet manifests
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" ./api/... ./controllers/...

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet
	go run ./main.go

# Install CRDs into a cluster
install: manifests
	kubectl apply -f config/crd/bases

crd/install: ## Install CRDs into a cluster
	kubectl apply -f config/crd/bases

crd/uninstall: manifests ## Uninstall CRDs from a cluster
	kubectl delete -f config/crd/bases

crd/upgrade: manifests ## Upgrade CRDs in a cluster
	kubectl diff -f config/crd/bases || kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./api/...;./controllers/..." $(CRD_OUTPUT)

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt paths=./api/...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}
	@echo "updating kustomize image patch file for manager resource"
	sed -i'' -e 's@image: .*@image: '"${IMG}"'@' ./config/default/manager_image_patch.yaml

# Push the docker image
docker-push:
	docker push ${IMG}

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.3.0
CONTROLLER_GEN=$(GOPATH_BIN_DIR)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
