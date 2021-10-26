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
DATAPLANE_API_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/envoyproxy,data-plane-api))
UDPA_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/cncf,udpa))
GOOGLEAPIS_ACTUAL_VERSION := $(shell $(call go_mod_latest_version,github.com/googleapis,googleapis))
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
	envoy/annotations/deprecation.proto=github.com/envoyproxy/go-control-plane/envoy/annotations \
	envoy/api/v2/core/address.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/backoff.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/base.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/core/socket_option.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2/core \
	envoy/api/v2/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/api/v2 \
	envoy/config/core/v3/address.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/backoff.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/base.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/http_uri.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/config/core/v3/socket_option.proto=github.com/envoyproxy/go-control-plane/envoy/config/core/v3 \
	envoy/service/discovery/v3/discovery.proto=github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3 \
	envoy/type/http_status.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/type/percent.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/type/semantic_version.proto=github.com/envoyproxy/go-control-plane/envoy/type \
	envoy/type/v3/percent.proto=github.com/envoyproxy/go-control-plane/envoy/type/v3 \
	envoy/type/v3/semantic_version.proto=github.com/envoyproxy/go-control-plane/envoy/type/v3 \
	google/protobuf/any.proto=google.golang.org/protobuf/types/known/anypb \
	google/protobuf/duration.proto=google.golang.org/protobuf/types/known/durationpb \
	google/protobuf/struct.proto=google.golang.org/protobuf/types/known/structpb \
	google/protobuf/timestamp.proto=google.golang.org/protobuf/types/known/timestamppb \
	google/protobuf/wrappers.proto=google.golang.org/protobuf/types/known/wrapperspb \
	udpa/annotations/migrate.proto=github.com/cncf/udpa/go/udpa/annotations \
	udpa/annotations/status.proto=github.com/cncf/udpa/go/udpa/annotations \
	udpa/annotations/versioning.proto=github.com/cncf/udpa/go/udpa/annotations \
	xds/core/v3/context_params.proto=github.com/cncf/udpa/go/xds/core/v3

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

generate: protoc/mesh protoc/mesh/v1alpha1 protoc/observability/v1 protoc/system/v1alpha1 ## Process .proto definitions

protoc/mesh:
	$(PROTOC_GO) mesh/*.proto

protoc/mesh/v1alpha1:
	$(PROTOC_GO) mesh/v1alpha1/*.proto

protoc/observability/v1:
	$(PROTOC_GO) observability/v1/*.proto

protoc/system/v1alpha1:
	$(PROTOC_GO) system/v1alpha1/*.proto

generate/docs:
	$(PROTOC_KUMADOC) mesh/v1alpha1/*.proto
