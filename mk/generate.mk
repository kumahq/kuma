ENVOY_IMPORTS := ./pkg/xds/envoy/imports.go
RESOURCE_GEN := $(KUMA_DIR)/build/tools-${GOOS}-${GOARCH}/resource-gen
POLICY_GEN := $(KUMA_DIR)/build/tools-${GOOS}-${GOARCH}/policy-gen/generator

PROTO_DIRS ?= ./pkg/config ./api ./pkg/plugins ./test/server/grpc/api
GO_MODULE ?= github.com/kumahq/kuma

HELM_VALUES_FILE ?= "deployments/charts/kuma/values.yaml"
HELM_CRD_DIR ?= "deployments/charts/kuma/crds/"
HELM_VALUES_FILE_POLICY_PATH ?= ".plugins.policies"

GENERATE_OAS_PREREQUISITES ?=
EXTRA_GENERATE_DEPS_TARGETS ?= generate/envoy-imports

.PHONY: clean/generated
clean/generated: clean/protos clean/builtin-crds clean/legacy-resources clean/resources clean/policies clean/tools

.PHONY: generate/protos
generate/protos:
	find $(PROTO_DIRS) -name '*.proto' -exec $(PROTOC_GO) {} \;

.PHONY: clean/tools
clean/tools:
	rm -rf $(KUMA_DIR)/build/tools-*

.PHONY: clean/proto
clean/protos: ## Dev: Remove auto-generated Protobuf files
	find $(PROTO_DIRS) -name '*.pb.go' -delete
	find $(PROTO_DIRS) -name '*.pb.validate.go' -delete

.PHONY: generate
generate: generate/protos generate/resources $(if $(findstring ./api,$(PROTO_DIRS)),resources/type generate/builtin-crds) generate/policies generate/oas $(EXTRA_GENERATE_DEPS_TARGETS) ## Dev: Run all code generation

$(POLICY_GEN):
	cd $(KUMA_DIR) && go build -o ./build/tools-${GOOS}-${GOARCH}/policy-gen/generator ./tools/policy-gen/generator/main.go

$(RESOURCE_GEN):
	cd $(KUMA_DIR) && go build -o ./build/tools-${GOOS}-${GOARCH}/resource-gen ./tools/resource-gen/main.go

.PHONY: resources/type
resources/type: $(RESOURCE_GEN)
	$(RESOURCE_GEN) -package mesh -generator type > pkg/core/resources/apis/mesh/zz_generated.resources.go
	$(RESOURCE_GEN) -package system -generator type > pkg/core/resources/apis/system/zz_generated.resources.go

.PHONY: clean/legacy-resources
clean/legacy-resources:
	find pkg -name 'zz_generated.*.go' -delete

POLICIES_DIR ?= pkg/plugins/policies
RESOURCES_DIR ?= pkg/core/resources/apis
COMMON_DIR := api/common

policies = $(foreach dir,$(shell find $(POLICIES_DIR) -maxdepth 1 -mindepth 1 -type d | grep -v -e '/core$$' | grep -v -e '/system$$' | grep -v -e '/mesh$$' | sort),$(notdir $(dir)))
kuma_policies = $(foreach dir,$(shell find $(KUMA_DIR)/pkg/plugins/policies -maxdepth 1 -mindepth 1 -type d | grep -v -e core | sort),$(notdir $(dir)))


.PHONY: clean/resources
clean/resources: POLICIES_DIR=$(RESOURCES_DIR)
clean/resources:
	POLICIES_DIR=$(RESOURCES_DIR) $(MAKE) clean/policies

generate/resources: POLICIES_DIR=$(RESOURCES_DIR)
generate/resources:
	POLICIES_DIR=$(RESOURCES_DIR) $(MAKE) $(addprefix generate/policy/,$(policies))
	POLICIES_DIR=$(RESOURCES_DIR) $(MAKE) generate/policy-import
	POLICIES_DIR=$(RESOURCES_DIR) HELM_VALUES_FILE_POLICY_PATH=".plugins.resources" $(MAKE) generate/policy-helm
	POLICIES_DIR=$(RESOURCES_DIR) $(MAKE) generate/policy-config

generate/policies: generate/deep-copy/common $(addprefix generate/policy/,$(policies)) generate/policy-import generate/policy-config generate/policy-defaults generate/policy-helm ## Generate all policies written as plugins

.PHONY: clean/policies
clean/policies: $(addprefix clean/policy/,$(policies))

# deletes all files in policy directory except *.proto, validator.go and schema.yaml
clean/policy/%:
	$(shell find $(POLICIES_DIR)/$* \( -name '*.pb.go' -o -name '*.yaml' -o -name 'zz_generated.*'  \) -not -path '*/testdata/*' -type f -delete)
	@rm -fr $(POLICIES_DIR)/$*/k8s

generate/deep-copy/common:
	for version in $(foreach dir,$(wildcard $(COMMON_DIR)/*),$(notdir $(dir))); do \
		$(CONTROLLER_GEN) object paths="./$(COMMON_DIR)/$$version/..."  ; \
	done

generate/policy/%: $(POLICY_GEN)
	$(POLICY_GEN) core-resource --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) k8s-resource --plugin-dir $(POLICIES_DIR)/$* --controller-gen-bin $(CONTROLLER_GEN) --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) plugin-file --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE) && \
	$(POLICY_GEN) helpers --plugin-dir $(POLICIES_DIR)/$* --gomodule $(GO_MODULE)
	$(POLICY_GEN) openapi --plugin-dir $(POLICIES_DIR)/$* --yq-bin $(YQ) --openapi-template-path=$(TOOLS_DIR)/policy-gen/templates/endpoints.yaml --jsonschema-template-path=$(TOOLS_DIR)/policy-gen/templates/schema.yaml --gomodule $(GO_MODULE)
	@echo "Policy $* successfully generated"

generate/policy-import:
	./tools/policy-gen/generate-policy-import.sh $(GO_MODULE) $(POLICIES_DIR) $(policies)

generate/policy-config:
	./tools/policy-gen/generate-policy-config.sh  $(POLICIES_DIR) $(policies)

generate/policy-defaults:
	./tools/policy-gen/generate-policy-defaults.sh $(KUMA_DIR) $(kuma_policies) $(policies)

generate/policy-helm:
	PATH=$(CI_TOOLS_BIN_DIR):$$PATH $(TOOLS_DIR)/policy-gen/generate-policy-helm.sh $(HELM_VALUES_FILE) $(HELM_CRD_DIR) $(HELM_VALUES_FILE_POLICY_PATH) $(POLICIES_DIR) $(policies)

endpoints = $(foreach dir,$(shell find api/openapi/specs -type f | sort),$(basename $(dir)))

generate/oas: $(GENERATE_OAS_PREREQUISITES)
	for endpoint in $(endpoints); do \
		DEST=$${endpoint#"api/openapi/specs"}; \
		PATH=$(CI_TOOLS_BIN_DIR):$$PATH oapi-codegen -config api/openapi/openapi.cfg.yaml -o api/openapi/types/$$(dirname $${DEST}})/zz_generated.$$(basename $${DEST}).go $${endpoint}.yaml; \
	done

.PHONY: generate/oas-for-ts
generate/oas-for-ts: generate/oas docs/generated/openapi.yaml ## Regenerate OpenAPI spec from `/api/openapi/specs` ready for typescript type generation


.PHONY: generate/builtin-crds
generate/builtin-crds: $(RESOURCE_GEN)
	$(RESOURCE_GEN) -package mesh -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.mesh.go
	$(RESOURCE_GEN) -package system -generator crd > ./pkg/plugins/resources/k8s/native/api/v1alpha1/zz_generated.system.go
	$(CONTROLLER_GEN) "crd:crdVersions=v1" paths=./pkg/plugins/resources/k8s/native/api/... output:crd:artifacts:config=$(HELM_CRD_DIR)
	$(CONTROLLER_GEN) object paths=./pkg/plugins/resources/k8s/native/api/...

.PHONY: clean/builtin-crds
clean/builtin-crds:
	rm -f ./deployments/charts/kuma/crds/*
	rm -f ./pkg/plugins/resources/k8s/native/test/config/crd/bases/*

.PHONY: generate/envoy-imports
generate/envoy-imports:
	printf 'package envoy\n\n' > ${ENVOY_IMPORTS}
	echo '// Import all Envoy packages so protobuf are registered and are ready to used in functions such as MarshalAny.' >> ${ENVOY_IMPORTS}
	echo '// This file is autogenerated. run "make generate/envoy-imports" to regenerate it after go-control-plane upgrade' >> ${ENVOY_IMPORTS}
	echo 'import (' >> ${ENVOY_IMPORTS}
	go list github.com/envoyproxy/go-control-plane/... | grep "github.com/envoyproxy/go-control-plane/envoy/" | awk '{printf "\t_ \"%s\"\n", $$1}' >> ${ENVOY_IMPORTS}
	echo ')' >> ${ENVOY_IMPORTS}
