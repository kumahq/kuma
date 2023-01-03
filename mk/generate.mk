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

POLICIES_DIR := pkg/plugins/policies

policies = $(foreach dir,$(shell find pkg/plugins/policies -maxdepth 1 -mindepth 1 -type d),$(notdir $(dir)))
generate_policy_targets = $(addprefix generate/policy/,$(policies))
cleanup_policy_targets = $(addprefix cleanup/policy/,$(policies))

generate/policies: $(cleanup_policy_targets) $(generate_policy_targets)

# deletes all files in policy directory except *.proto and validator.go
cleanup/policy/%:
<<<<<<< HEAD
	$(shell find $(POLICIES_DIR)/$* -not -name '*.proto' -not -name 'validator.go' -type f -delete)
	@rm -r $(POLICIES_DIR)/$*/k8s || true
=======
	$(shell find $(POLICIES_DIR)/$* \( -name '*.pb.go' -o -name '*.yaml' -o -name 'zz_generated.*'  \) -not -path '*/testdata/*' -type f -delete)
	@rm -fr $(POLICIES_DIR)/$*/k8s

generate/deep-copy/common:
	for version in $(foreach dir,$(wildcard $(COMMON_DIR)/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) object paths=$(COMMON_DIR)/$$version/targetref.go ; \
	done
>>>>>>> fdac0e281 (chore: remove Apache license header from generated files (#5565))

generate/policy/%: generate/schema/%
	@echo "Copy CRD to helm directory"
	$(MAKE) generate/helm/$*
	@echo "Policy $* successfully generated"
	@echo "Don't forget to update Helm chart's cp-rbac.yaml with $*"

generate/schema/%: generate/controller-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		tools/policy-gen/crd-extract-openapi.sh $* $$version ; \
	done

generate/controller-gen/%: generate/kumapolicy-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) "crd:crdVersions=v1,ignoreUnexportedFields=true" paths="./$(POLICIES_DIR)/$*/k8s/..." output:crd:artifacts:config=$(POLICIES_DIR)/$*/k8s/crd && \
		$(CONTROLLER_GEN) object paths=$(POLICIES_DIR)/$*/k8s/$$version/zz_generated.types.go ; \
<<<<<<< HEAD
=======
		$(CONTROLLER_GEN) object paths=$(POLICIES_DIR)/$*/api/$$version/$*.go ; \
>>>>>>> fdac0e281 (chore: remove Apache license header from generated files (#5565))
	done

generate/kumapolicy-gen/%: generate/dirs/%
	cd tools/policy-gen/protoc-gen-kumapolicy && go build && cd - ; \
	$(PROTOC) \
		--proto_path=./api \
		--kumapolicy_opt=endpoints-template=tools/policy-gen/templates/endpoints.yaml \
		--kumapolicy_out=$(POLICIES_DIR)/$* \
		--go_opt=paths=source_relative \
		--go_out=plugins=grpc,$(go_mapping):. \
		--plugin=protoc-gen-kumapolicy=tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy \
		$(POLICIES_DIR)/$*/api/*/*.proto ; \
	rm tools/policy-gen/protoc-gen-kumapolicy/protoc-gen-kumapolicy ; \

generate/dirs/%:
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		mkdir -p $(POLICIES_DIR)/$*/api/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/crd ; \
	done


<<<<<<< HEAD
generate/helm/%:
	tools/policy-gen/crd-helm-copy.sh $*


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
=======
.PHONY: crd/controller-gen
crd/controller-gen:
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths=$(IN_CRD) output:crd:artifacts:config=$(OUT_CRD)
	$(CONTROLLER_GEN) object paths=$(IN_CRD)
>>>>>>> fdac0e281 (chore: remove Apache license header from generated files (#5565))

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
