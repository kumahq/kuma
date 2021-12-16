.PHONY: all test install crd/install crd/uninstall crd/upgrade deploy generate controller-gen

CRD_OPTIONS ?= "crd:crdVersions=v1"
CRD_OUTPUT := output:crd:artifacts:config=config/crd/bases

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

CONTROLLER_GEN := go run -mod=mod sigs.k8s.io/controller-tools/cmd/controller-gen
RESOURCE_GEN := go run -mod=mod ../../../../../tools/resource-gen/main.go

all:

GO_TEST ?= go test
GO_TEST_OPTS ?=
COVERAGE_PROFILE ?= cover.out

# Run tests
test: generate
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" ./api/... ./controllers/...

# Install CRDs into a cluster
install: crd/install

crd/install: generate ## Install CRDs into a cluster
	kubectl apply -f config/crd/bases

crd/uninstall: generate ## Uninstall CRDs from a cluster
	kubectl delete -f config/crd/bases

crd/upgrade: generate ## Upgrade CRDs in a cluster
	kubectl diff -f config/crd/bases || kubectl apply -f config/crd/bases

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: generate
	kubectl apply -f config/crd/bases
	kustomize build config/default | kubectl apply -f -

.PHONY: generate
generate: generate/types
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./api/v1alpha1/gateway_instance.go;./api/v1alpha1/groupversion_info.go" $(CRD_OUTPUT)
	$(CONTROLLER_GEN) object:headerFile=./hack/boilerplate.go.txt,year=$$(date +%Y) paths=./api/...

.PHONY:
generate/types:
	$(RESOURCE_GEN) -package mesh -generator crd > api/v1alpha1/zz_generated.mesh.go
	$(RESOURCE_GEN) -package system -generator crd > api/v1alpha1/zz_generated.system.go

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
