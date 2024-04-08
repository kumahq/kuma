PROTOC := $(PROTOC_BIN) \
	--proto_path=$(PROTOS_DEPS_PATH) \
	--proto_path=$(KUMA_DIR) \
	--proto_path=.

PROTOC_GO := $(PROTOC) \
	--plugin=protoc-gen-go=$(PROTOC_GEN_GO) \
	--plugin=protoc-gen-go-grpc=$(PROTOC_GEN_GO_GRPC) \
	--go_opt=paths=source_relative \
	--go_out=. \
	--go-grpc_opt=paths=source_relative \
	--go-grpc_out=.
