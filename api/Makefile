.PHONY: help clean generate build test check \
		dev/tools install/protoc install/protoc-gen-go install/protoc-gen-validate \
		install/protobuf-wellknown-types install/data-plane-api \
		protoc protoc/mesh/v1alpha1 protoc/observability/v1alpha1

TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
export PATH := $(TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

PROTOC_PATH := $(TOOLS_DIR)/protoc
PROTOBUF_WKT_DIR := $(TOOLS_DIR)/protobuf.d

PROTOC_VERSION := 3.6.1
PROTOC_PGV_VERSION := v0.1.0
GOLANG_PROTOBUF_VERSION := v1.3.2
DATAPLANE_API_LATEST_VERSION := master
DATAPLANE_API_ACTUAL_VERSION := $(shell find $(GOPATH_DIR)/pkg/mod/github.com/envoyproxy -maxdepth 1 -name 'data-plane-api@*' | xargs -L 1 basename | sort -r | head -1 | awk -F@ '{print $$2}')
UDPA_LATEST_VERSION := master
UDPA_ACTUAL_VERSION := $(shell find $(GOPATH_DIR)/pkg/mod/github.com/cncf -maxdepth 1 -name 'udpa@*' | xargs -L 1 basename | sort -r | head -1 | awk -F@ '{print $$2}')
GOOGLEAPIS_LATEST_VERSION := master
GOOGLEAPIS_ACTUAL_VERSION := $(shell find $(GOPATH_DIR)/pkg/mod/github.com/googleapis -maxdepth 1 -name 'googleapis@*' | xargs -L 1 basename | sort -r | head -1 | awk -F@ '{print $$2}')

protoc_search_go_packages := \
	github.com/golang/protobuf@$(GOLANG_PROTOBUF_VERSION) \
	github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION) \
	github.com/envoyproxy/data-plane-api@$(DATAPLANE_API_ACTUAL_VERSION) \
	github.com/cncf/udpa@$(UDPA_ACTUAL_VERSION) \
	github.com/googleapis/googleapis@$(GOOGLEAPIS_ACTUAL_VERSION)

protoc_search_go_paths := $(foreach go_package,$(protoc_search_go_packages),--proto_path=$(GOPATH_DIR)/pkg/mod/$(go_package))

go_import_mapping_entries := \
	google/protobuf/any.proto=github.com/golang/protobuf/ptypes/any \
	google/protobuf/duration.proto=github.com/golang/protobuf/ptypes/duration \
	google/protobuf/struct.proto=github.com/golang/protobuf/ptypes/struct \
	google/protobuf/timestamp.proto=github.com/golang/protobuf/ptypes/timestamp \
	google/protobuf/wrappers.proto=github.com/golang/protobuf/ptypes/wrappers \
	envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2

# see https://makefiletutorial.com/
comma := ,
empty:=
space := $(empty) $(empty)

go_mapping_with_spaces := $(foreach entry,$(go_import_mapping_entries),M$(entry),)
go_mapping := $(subst $(space),$(empty),$(go_mapping_with_spaces))

PROTOC_GO := protoc \
	--proto_path=$(PROTOBUF_WKT_DIR)/include \
	--proto_path=. \
	$(protoc_search_go_paths) \
	--go_out=plugins=grpc,$(go_mapping):. \
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

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

dev/tools: install/protoc install/protobuf-wellknown-types install/data-plane-api install/protoc-gen-go install/protoc-gen-validate ## Install development tools

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

install/protobuf-wellknown-types:
	@if [ -e $(PROTOBUF_WKT_DIR) ]; then echo "Protobuf well-known types are already installed at $(PROTOBUF_WKT_DIR)" ; fi
	@if [ ! -e $(PROTOBUF_WKT_DIR) ]; then \
		echo "Installing Protobuf well-known types $(PROTOC_VERSION) ..." \
		&& set -x \
		&& curl -Lo /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip 'include/*' -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(PROTOBUF_WKT_DIR) \
		&& cp -r /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH)/include $(PROTOBUF_WKT_DIR) \
		&& rm -rf /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& rm /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& set +x \
		&& echo "Protobuf well-known types $(PROTOC_VERSION) have been installed at $(PROTOBUF_WKT_DIR)" ; fi

install/protoc-gen-go:
	go get -u github.com/golang/protobuf/protoc-gen-go@$(GOLANG_PROTOBUF_VERSION)

install/protoc-gen-validate:
	go get -u github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION)

install/data-plane-api:
	go get -u github.com/envoyproxy/data-plane-api@$(DATAPLANE_API_LATEST_VERSION)
	go get -u github.com/cncf/udpa@$(UDPA_LATEST_VERSION)
	go get -u github.com/googleapis/googleapis@$(GOOGLEAPIS_LATEST_VERSION)

clean: ## Remove generated files
	find . -name '*.pb.go' -delete
	find . -name '*.pb.validate.go' -delete

generate: clean protoc/mesh/v1alpha1 protoc/observability/v1alpha1 ## Process .proto definitions

protoc/mesh/v1alpha1:
	$(PROTOC_GO) mesh/v1alpha1/*.proto

protoc/observability/v1alpha1:
	$(PROTOC_GO) observability/v1alpha1/*.proto

build: ## Build generated files
	go build ./...

test: ## Run tests
	go test ./...

check: generate build test ## Verify that auto-generated code is up-to-date
	git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty
