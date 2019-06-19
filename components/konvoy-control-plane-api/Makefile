.PHONY: help clean generate build test check \
		dev/tools install/protoc install/protoc-gen-gogofast \
		protoc protoc/discovery/v1alpha1 protoc/mesh/v1alpha1

TOOLS_DIR ?= $(HOME)/bin
export PATH := $(TOOLS_DIR):$(shell go env GOPATH)/bin:$(PATH)

PROTOC_PATH := $(TOOLS_DIR)/protoc

PROTOC_VERSION := 3.6.1
GOGO_PROTOBUF_VERSION := v1.2.1

PROTOC_GO := protoc \
	--proto_path=. \
	--gogofast_out=plugins=grpc:.

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

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev/tools: install/protoc install/protoc-gen-gogofast ## Install development tools

install/protoc:
	@if [ -e $(PROTOC_PATH) ]; then echo "Protoc $$( $(PROTOC_PATH) --version ) is already installed at $(PROTOC_PATH)" ; fi
	@if [ ! -e $(PROTOC_PATH) ]; then \
		echo "Installing Protoc $(PROTOC_VERSION) ..." \
		&& set -x \
		&& curl -Lo /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip bin/protoc -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(TOOLS_DIR) \
		&& cp /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH)/bin/protoc $(PROTOC_PATH) \
		&& rm -rf /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& rm /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& set +x \
		&& echo "Protoc $(PROTOC_VERSION) has been installed at $(PROTOC_PATH)" ; fi

install/protoc-gen-gogofast:
	go get -u github.com/gogo/protobuf/protoc-gen-gogofast@$(GOGO_PROTOBUF_VERSION)

clean: ## Remove generated files
	find . -name '*.pb.go' -delete

generate: clean protoc/discovery/v1alpha1 protoc/mesh/v1alpha1 ## Process .proto definitions

protoc/discovery/v1alpha1:
	$(PROTOC_GO) discovery/v1alpha1/*.proto

protoc/mesh/v1alpha1:
	$(PROTOC_GO) mesh/v1alpha1/*.proto

build: ## Build generated files
	go build ./...

test: ## Run tests
	go test ./...

check: generate build test ## Verify that auto-generated code is up-to-date
	git diff --quiet || ( echo "The following changes (result of code generators) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty
