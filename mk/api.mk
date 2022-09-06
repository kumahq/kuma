
#
# Re-usable snippets
#

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

go_mapping_with_spaces := $(foreach entry,$(go_import_mapping_entries),M$(entry),)
go_mapping := $(subst $(space),$(empty),$(go_mapping_with_spaces))

PROTOC := $(PROTOC_BIN) \
	--proto_path=$(PROTOS_DEPS_PATH) \
	--proto_path=. \

PROTOC_GO := $(PROTOC) \
	--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
	--plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
	--go_opt=paths=source_relative \
	--go_out=$(go_mapping):. \
	--go-grpc_opt=paths=source_relative \
	--go-grpc_out=$(go_mapping):.


protoc/common/v1alpha1:
	cd api && $(PROTOC_GO) common/v1alpha1/*.proto

protoc/mesh:
	cd api && $(PROTOC_GO) mesh/*.proto

protoc/mesh/v1alpha1:
	cd api && $(PROTOC_GO) mesh/v1alpha1/*.proto

protoc/observability/v1:
	cd api && $(PROTOC_GO) observability/v1/*.proto

protoc/system/v1alpha1:
	cd api && $(PROTOC_GO) system/v1alpha1/*.proto
