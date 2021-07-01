.PHONY: help clean generate build test check \
		dev/tools install/protoc install/protoc-gen-go install/protoc-gen-validate \
		install/protobuf-wellknown-types install/data-plane-api \
		protoc protoc/mesh/v1alpha1 protoc/observability/v1alpha1 protoc/system/v1alpha1

CI_TOOLS_DIR ?= $(HOME)/bin
GOPATH_DIR := $(shell go env GOPATH | awk -F: '{print $$1}')
GOPATH_BIN_DIR := $(GOPATH_DIR)/bin
export PATH := $(CI_TOOLS_DIR):$(GOPATH_BIN_DIR):$(PATH)

PROTOC_PATH := $(CI_TOOLS_DIR)/protoc
PROTOBUF_WKT_DIR := $(CI_TOOLS_DIR)/protobuf.d

#
# Re-usable snippets
#

define go_mod_latest_version
	find $(GOPATH_DIR)/pkg/mod/$(1) -maxdepth 1 -name '$(2)@*' | xargs -L 1 basename | sort -r | head -1 | awk -F@ '{print $$2}'
endef

PROTOC_VERSION := 3.14.0
PROTOC_PGV_VERSION := v0.4.1
GOLANG_PROTOBUF_VERSION := v1.5.2
DATAPLANE_API_LATEST_VERSION := master
DATAPLANE_API_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/envoyproxy,data-plane-api))
UDPA_LATEST_VERSION := master
UDPA_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/cncf,udpa))
GOOGLEAPIS_LATEST_VERSION := master
GOOGLEAPIS_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/googleapis,googleapis))
KUMADOC_VERSION := v0.1.7
KUMADOC_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/kumahq,protoc-gen-kumadoc))

protoc_search_go_packages := \
	github.com/golang/protobuf@$(GOLANG_PROTOBUF_VERSION) \
	github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION) \
	github.com/envoyproxy/data-plane-api@$(DATAPLANE_API_ACTUAL_VERSION) \
	github.com/cncf/udpa@$(UDPA_ACTUAL_VERSION) \
	github.com/googleapis/googleapis@$(GOOGLEAPIS_ACTUAL_VERSION) \
	github.com/kumahq/protoc-gen-kumadoc@$(KUMADOC_ACTUAL_VERSION)/proto

protoc_search_go_paths := $(foreach go_package,$(protoc_search_go_packages),--proto_path=$(GOPATH_DIR)/pkg/mod/$(go_package))

go_import_mapping_entries := \
	google/protobuf/any.proto=google.golang.org/protobuf/types/known/anypb \
	google/protobuf/duration.proto=google.golang.org/protobuf/types/known/durationpb \
	google/protobuf/struct.proto=google.golang.org/protobuf/types/known/structpb \
	google/protobuf/timestamp.proto=google.golang.org/protobuf/types/known/timestamppb \
	google/protobuf/wrappers.proto=google.golang.org/protobuf/types/known/wrapperspb \
	envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2 \
	envoy/service/discovery/v3/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3 \
	udpa/annotations/migrate.proto=github.com/cncf/udpa/go/udpa/annotations \
	udpa/annotations/status.proto=github.com/cncf/udpa/go/udpa/annotations \
	udpa/annotations/versioning.proto=github.com/cncf/udpa/go/udpa/annotations \
	envoy/api/v2/core/socket_option.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/address.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/backoff.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/base.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/type/percent.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/type/semantic_version.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/type/http_status.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/config/core/v3/socket_option.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/address.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/backoff.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/type/v3/percent.proto=github.com/envoyproxy/go-control-plane/envoy/type/v3 \
	envoy/type/v3/semantic_version.proto=github.com/envoyproxy/go-control-plane/envoy/type/v3 \
	envoy/config/core/v3/base.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3


# see https://makefiletutorial.com/
comma := ,
empty:=
space := $(empty) $(empty)

DOCS_OUTPUT_PATH ?= .

additional_params ?=

ifdef PLUGIN
	additional_params += --plugin=protoc-gen-kumadoc=$(PLUGIN)
endif

go_mapping_with_spaces := $(foreach entry,$(go_import_mapping_entries),M$(entry),)
go_mapping := $(subst $(space),$(empty),$(go_mapping_with_spaces))

PROTOC := protoc \
	--proto_path=$(PROTOBUF_WKT_DIR)/include \
	--proto_path=. \
	$(protoc_search_go_paths)

PROTOC_GO := $(PROTOC) \
	--go_opt=paths=source_relative \
	--go_out=plugins=grpc,$(go_mapping):.

PROTOC_KUMADOC := $(PROTOC) \
	--kumadoc_out=$(DOCS_OUTPUT_PATH) \
	$(additional_params)

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

dev/tools: install/protoc install/protobuf-wellknown-types install/data-plane-api install/protoc-gen-go install/protoc-gen-validate install/protoc-gen-kumadoc ## Install development tools

install/protoc:
	@if [ -e $(PROTOC_PATH) ]; then echo "Protoc $$( $(PROTOC_PATH) --version ) is already installed at $(PROTOC_PATH)" ; fi
	@if [ ! -e $(PROTOC_PATH) ]; then \
		echo "Installing Protoc $(PROTOC_VERSION) ..." \
		&& set -x \
		&& curl -Lo /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip https://github.com/protocolbuffers/protobuf/releases/download/v$(PROTOC_VERSION)/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip \
		&& unzip /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH).zip bin/protoc -d /tmp/protoc-$(PROTOC_VERSION)-$(PROTOC_OS)-$(PROTOC_ARCH) \
		&& mkdir -p $(CI_TOOLS_DIR) \
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
	go install github.com/golang/protobuf/protoc-gen-go@$(GOLANG_PROTOBUF_VERSION)

install/protoc-gen-kumadoc:
	go install github.com/kumahq/protoc-gen-kumadoc@$(KUMADOC_VERSION)

install/protoc-gen-validate:
	go install github.com/envoyproxy/protoc-gen-validate@$(PROTOC_PGV_VERSION)

install/data-plane-api:
	go get github.com/envoyproxy/data-plane-api@$(DATAPLANE_API_LATEST_VERSION)
	go get github.com/cncf/udpa@$(UDPA_LATEST_VERSION)
	go get github.com/googleapis/googleapis@$(GOOGLEAPIS_LATEST_VERSION)

clean: ## Remove generated files
	find . -name '*.pb.go' -delete
	find . -name '*.pb.validate.go' -delete

generate: clean protoc/mesh protoc/mesh/v1alpha1 protoc/observability/v1alpha1 protoc/observability/v1 protoc/system/v1alpha1 ## Process .proto definitions

protoc/mesh:
	$(PROTOC_GO) mesh/*.proto

protoc/mesh/v1alpha1:
	$(PROTOC_GO) mesh/v1alpha1/*.proto

protoc/observability/v1alpha1:
	$(PROTOC_GO) observability/v1alpha1/*.proto

protoc/observability/v1:
	$(PROTOC_GO) observability/v1/*.proto

protoc/system/v1alpha1:
	$(PROTOC_GO) system/v1alpha1/*.proto

build: ## Build generated files
	go build ./...

GO_TEST ?= go test
GO_TEST_OPTS ?=

test: ## Run tests
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" ./...

check: generate build test ## Verify that auto-generated code is up-to-date
	git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty

generate/docs:
	$(PROTOC_KUMADOC) mesh/v1alpha1/*.proto
