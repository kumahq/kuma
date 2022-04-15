ENVOY_IMPORTS := ./pkg/xds/envoy/imports.go
PROTO_DIRS := ./pkg/config ./api

CONTROLLER_GEN := go run -mod=mod sigs.k8s.io/controller-tools/cmd/controller-gen@master

.PHONY: clean/proto
clean/proto: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIRS) -name '*.pb.go' -delete
	find $(PROTO_DIRS) -name '*.pb.validate.go' -delete

.PHONY: generate
generate:  ## Dev: Run code generators
generate: clean/proto generate/api protoc/pkg/config/app/kumactl/v1alpha1 protoc/pkg/test/apis/sample/v1alpha1 protoc/plugins resources/type generate/kubernetes

.PHONY: resources/type
resources/type:
	$(GO_RUN) ./tools/resource-gen/main.go -package mesh -generator type > pkg/core/resources/apis/mesh/zz_generated.resources.go
	$(GO_RUN) ./tools/resource-gen/main.go -package system -generator type > pkg/core/resources/apis/system/zz_generated.resources.go

.PHONY: protoc/pkg/config/app/kumactl/v1alpha1
protoc/pkg/config/app/kumactl/v1alpha1:
	$(PROTOC_GO) pkg/config/app/kumactl/v1alpha1/*.proto

.PHONY: protoc/pkg/test/apis/sample/v1alpha1
protoc/pkg/test/apis/sample/v1alpha1:
	$(PROTOC_GO) pkg/test/apis/sample/v1alpha1/*.proto

.PHONY: protoc/plugins
protoc/plugins:
	$(PROTOC_GO) --proto_path=./api pkg/plugins/ca/provided/config/*.proto
	$(PROTOC_GO) --proto_path=./api pkg/plugins/ca/builtin/config/*.proto

generate/dirs/%:
	pushd pkg/plugins/policies/$* && \
	mkdir -p api/v1alpha1 && \
	mkdir -p k8s/v1alpha1 && \
	mkdir -p k8s/crd && \
	popd

generate/kumapolicy-gen/%: generate/dirs/%
	pushd tools/policy-gen/protoc-gen-kumapolicy && go build && popd
	$(PROTOC) \
		--proto_path=./api \
		--kumapolicy_opt=endpoints-template=tools/policy-gen/templates/endpoints.yaml \
		--kumapolicy_out=pkg/plugins/policies/$* \
		--go_opt=paths=source_relative \
		--go_out=plugins=grpc,$(go_mapping):. \
		--plugin=protoc-gen-kumapolicy=tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy \
		pkg/plugins/policies/$*/api/v1alpha1/*.proto
	@rm tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy

generate/controller-gen/%: generate/kumapolicy-gen/%
	$(CONTROLLER_GEN) "crd:crdVersions=v1,ignoreUnexportedFields=true" paths="./pkg/plugins/policies/$*/k8s/..." output:crd:artifacts:config=pkg/plugins/policies/$*/k8s/crd
	$(CONTROLLER_GEN) object paths=pkg/plugins/policies/$*/k8s/v1alpha1/zz_generated.types.go

generate/schema/%: generate/controller-gen/%
	tools/policy-gen/crd-extract-openapi.sh $*

generate/policy/%: generate/schema/%
	@echo "Policy $* successfully generated"

generate/helm/%:
	tools/policy-gen/crd-helm-copy.sh $* && \


KUMA_GUI_GIT_URL=https://github.com/kumahq/kuma-gui.git
KUMA_GUI_VERSION=master
KUMA_GUI_FOLDER=app/kuma-ui/pkg/resources/data
KUMA_GUI_WORK_FOLDER=app/kuma-ui/data/work

.PHONY: upgrade/gui
upgrade/gui:
	rm -rf $(KUMA_GUI_WORK_FOLDER)
	git clone --depth 1 -b $(KUMA_GUI_VERSION) $(KUMA_GUI_GIT_URL) $(KUMA_GUI_WORK_FOLDER)
	cd $(KUMA_GUI_WORK_FOLDER) && yarn install && yarn build
	rm -rf $(KUMA_GUI_FOLDER) && mv $(KUMA_GUI_WORK_FOLDER)/dist/ $(KUMA_GUI_FOLDER)
	rm -rf $(KUMA_GUI_WORK_FOLDER)

.PHONY: generate/envoy-imports
generate/envoy-imports:
	echo 'package envoy\n' > ${ENVOY_IMPORTS}
	echo '// Import all Envoy packages so protobuf are registered and are ready to used in functions such as MarshalAny.' >> ${ENVOY_IMPORTS}
	echo '// This file is autogenerated. run "make generate/envoy-imports" to regenerate it after go-control-plane upgrade' >> ${ENVOY_IMPORTS}
	echo 'import (' >> ${ENVOY_IMPORTS}
	go list github.com/envoyproxy/go-control-plane/... | grep "github.com/envoyproxy/go-control-plane/envoy/" | awk '{printf "\t_ \"%s\"\n", $$1}' >> ${ENVOY_IMPORTS}
	echo ')' >> ${ENVOY_IMPORTS}

.PHONY: generate/kubernetes
generate/kubernetes:
	$(MAKE) -C pkg/plugins/resources/k8s/native generate

.PHONY: generate/api
generate/api: protoc/mesh protoc/mesh/v1alpha1 protoc/observability/v1 protoc/system/v1alpha1 ## Process Kuma API .proto definitions
