ENVOY_IMPORTS := ./pkg/xds/envoy/imports.go
PROTO_DIRS ?= ./pkg/config ./api
CONTROLLER_GEN := go run -mod=mod sigs.k8s.io/controller-tools/cmd/controller-gen
RESOURCE_GEN := go run -mod=mod $(TOOLS_DIR)/resource-gen/main.go
POLICY_GEN := go run -mod=mod $(TOOLS_DIR)/policy-gen/generator/main.go

GENERATE_TARGET ?= build/ebpf clean/proto generate/api protoc/pkg/config/app/kumactl/v1alpha1 protoc/plugins resources/type generate/policies
GO_MODULE ?= github.com/kumahq/kuma

HELM_VALUES_FILE ?= "deployments/charts/kuma/values.yaml"
HELM_CRD_DIR ?= "deployments/charts/kuma/crds/"
HELM_VALUES_FILE_POLICY_PATH ?= ".plugins.policies"

.PHONY: clean/proto
clean/proto: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIRS) -name '*.pb.go' -delete
	find $(PROTO_DIRS) -name '*.pb.validate.go' -delete

.PHONY: generate
generate:  ## Dev: Run code generators
generate: $(GENERATE_TARGET)

.PHONY: resources/type
resources/type:
	go run $(TOOLS_DIR)/resource-gen/main.go -package mesh -generator type > pkg/core/resources/apis/mesh/zz_generated.resources.go
	go run $(TOOLS_DIR)/resource-gen/main.go -package system -generator type > pkg/core/resources/apis/system/zz_generated.resources.go

.PHONY: protoc/pkg/config/app/kumactl/v1alpha1
protoc/pkg/config/app/kumactl/v1alpha1:
	$(PROTOC_GO) pkg/config/app/kumactl/v1alpha1/*.proto

.PHONY: protoc/plugins
protoc/plugins:
	$(PROTOC_GO) pkg/plugins/ca/provided/config/*.proto
	$(PROTOC_GO) pkg/plugins/ca/builtin/config/*.proto

POLICIES_DIR := pkg/plugins/policies
COMMON_DIR := api/common

policies = $(foreach dir,$(shell find pkg/plugins/policies -maxdepth 1 -mindepth 1 -type d | grep -v -e core -e matchers -e xds -e validation -e common -e jsonpatch | sort),$(notdir $(dir)))
generate_policy_targets = $(addprefix generate/policy/,$(policies))
cleanup_policy_targets = $(addprefix cleanup/policy/,$(policies))

GENERATE_POLICIES_TARGET ?= cleanup/crds cleanup/policies generate/deep-copy/common $(generate_policy_targets) generate/policy-import generate/policy-helm generate/builtin-crds generate/fix-embed
generate/policies: $(GENERATE_POLICIES_TARGET) ## Generate all policies written as plugins

cleanup/crds:
	rm -f ./deployments/charts/kuma/crds/*
	rm -f ./pkg/plugins/resources/k8s/native/test/config/crd/bases/*

cleanup/policies: $(cleanup_policy_targets)

# deletes all files in policy directory except *.proto, validator.go and schema.yaml
cleanup/policy/%:
	$(shell find $(POLICIES_DIR)/$* \( -name '*.pb.go' -o -name '*.yaml' -o -name 'zz_generated.*'  \) -not -path '*/testdata/*' -type f -delete)
	@rm -fr $(POLICIES_DIR)/$*/k8s

generate/deep-copy/common:
	for version in $(foreach dir,$(wildcard $(COMMON_DIR)/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) object paths="./$(COMMON_DIR)/$$version/..."  ; \
	done

generate/policy/%: generate/schema/%
	@echo "Policy $* successfully generated"

generate/schema/%: generate/controller-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		PATH=$(CI_TOOLS_BIN_DIR):$$PATH $(TOOLS_DIR)/policy-gen/crd-extract-openapi.sh $* $$version $(TOOLS_DIR) ; \
	done

generate/policy-import:
	$(TOOLS_DIR)/policy-gen/generate-policy-import.sh $(GO_MODULE) $(policies)

generate/policy-helm:
	PATH=$(CI_TOOLS_BIN_DIR):$$PATH $(TOOLS_DIR)/policy-gen/generate-policy-helm.sh $(HELM_VALUES_FILE) $(HELM_CRD_DIR) $(HELM_VALUES_FILE_POLICY_PATH) $(policies)

generate/controller-gen/%: generate/kumapolicy-gen/%
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) "crd:crdVersions=v1,ignoreUnexportedFields=true" paths="./$(POLICIES_DIR)/$*/k8s/..." output:crd:artifacts:config=$(POLICIES_DIR)/$*/k8s/crd && \
		$(CONTROLLER_GEN) object paths=$(POLICIES_DIR)/$*/k8s/$$version/zz_generated.types.go && \
		$(CONTROLLER_GEN) object paths=$(POLICIES_DIR)/$*/api/$$version/$*.go; \
	done

generate/kumapolicy-gen/%: generate/dirs/%
	$(POLICY_GEN) core-resource --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) k8s-resource --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) openapi --plugin-dir $(POLICIES_DIR)/$* --openapi-template-path=$(TOOLS_DIR)/policy-gen/templates/endpoints.yaml --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) plugin-file --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) helpers --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE)

generate/dirs/%:
	for version in $(foreach dir,$(wildcard $(POLICIES_DIR)/$*/api/*),$(notdir $(dir))); do \
		mkdir -p $(POLICIES_DIR)/$*/api/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/$$version ; \
		mkdir -p $(POLICIES_DIR)/$*/k8s/crd ; \
	done

.PHONY: generate/builtin-crds
generate/builtin-crds:
	$(RESOURCE_GEN) -package mesh -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.mesh.go
	$(RESOURCE_GEN) -package system -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.system.go
	$(MAKE) OUT_CRD=./deployments/charts/kuma/crds IN_CRD=./pkg/plugins/resources/k8s/native/api/... crd/controller-gen

.PHONY: crd/controller-gen
crd/controller-gen:
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths=$(IN_CRD) output:crd:artifacts:config=$(OUT_CRD)
	$(CONTROLLER_GEN) object paths=$(IN_CRD)

.PHONY: generate/envoy-imports
generate/envoy-imports:
	echo 'package envoy\n' > ${ENVOY_IMPORTS}
	echo '// Import all Envoy packages so protobuf are registered and are ready to used in functions such as MarshalAny.' >> ${ENVOY_IMPORTS}
	echo '// This file is autogenerated. run "make generate/envoy-imports" to regenerate it after go-control-plane upgrade' >> ${ENVOY_IMPORTS}
	echo 'import (' >> ${ENVOY_IMPORTS}
	go list github.com/envoyproxy/go-control-plane/... | grep "github.com/envoyproxy/go-control-plane/envoy/" | awk '{printf "\t_ \"%s\"\n", $$1}' >> ${ENVOY_IMPORTS}
	echo ')' >> ${ENVOY_IMPORTS}

.PHONY: generate/api
generate/api: protoc/mesh protoc/mesh/v1alpha1 protoc/observability/v1 protoc/system/v1alpha1 ## Process Kuma API .proto definitions

generate/test-server:
	$(PROTOC_GO) test/server/grpc/api/*.proto

# this is a fix for controller-gen complaining that there is no schema.yaml file
# controller-gen it imports a policy package and then the //go:embed match is triggered and causes an error
# the workaround works by having _DELETE_GO_EMBED_WORKAROUND_go:embed that is not evaluated and removing the
# _DELETE_GO_EMBED_WORKAROUND_ part after generation
.PHONY: generate/fix-embed
generate/fix-embed:
	find $(POLICIES_DIR) -name zz_generated.resource.go -type f -exec $(SHELL) -c "sed -i.bak 's/ _DELETE_GO_EMBED_WORKAROUND_//g' {} && rm {}.bak" \;
