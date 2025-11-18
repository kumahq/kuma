ENVOY_IMPORTS := ./pkg/xds/envoy/imports.go
RESOURCE_GEN := ./build/tools-${GOOS}-${GOARCH}/resource-gen
OAPI_GEN := ./build/tools-${GOOS}-${GOARCH}/oapi-gen
POLICY_GEN := $(KUMA_DIR)/build/tools-${GOOS}-${GOARCH}/policy-gen/generator

PROTO_DIRS ?= ./pkg/config ./api ./pkg/plugins ./test/server/grpc/api
GO_MODULE ?= github.com/kumahq/kuma/v2

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
generate: generate/protos generate/resources $(if $(findstring ./api,$(PROTO_DIRS)),resources/type generate/builtin-crds) generate/policies api-lint generate/oas $(EXTRA_GENERATE_DEPS_TARGETS) ## Dev: Run all code generation

$(POLICY_GEN): $(wildcard $(KUMA_DIR)/tools/policy-gen/**/*)
	cd $(KUMA_DIR) && $(GO) build -o ./build/tools-${GOOS}-${GOARCH}/policy-gen/generator ./tools/policy-gen/generator/main.go

$(RESOURCE_GEN): $(wildcard $(KUMA_DIR)/tools/resource-gen/**/*)  $(wildcard $(KUMA_DIR)/tools/policy-gen/**/*)
	$(GO) build -o ./build/tools-${GOOS}-${GOARCH}/resource-gen ./tools/resource-gen/main.go

$(OAPI_GEN): $(wildcard $(KUMA_DIR)/tools/openapi/**/*) $(wildcard $(KUMA_DIR)/tools/resource-gen/**/*)  $(wildcard $(KUMA_DIR)/tools/policy-gen/**/*)
	$(GO) build -o ./build/tools-${GOOS}-${GOARCH}/oapi-gen ./tools/openapi/generator/main.go

.PHONY: resources/type
resources/type: $(RESOURCE_GEN)
	$(RESOURCE_GEN) -package mesh -generator type > pkg/core/resources/apis/mesh/zz_generated.resources.go
	$(RESOURCE_GEN) -package system -generator type > pkg/core/resources/apis/system/zz_generated.resources.go

.PHONY: clean/legacy-resources
clean/legacy-resources:
	find pkg -name 'zz_generated.*.go' -delete

POLICIES_DIR ?= pkg/plugins/policies
RESOURCES_DIR ?= pkg/core/resources/apis
MESH_API_DIR ?= api/mesh/v1alpha1
SYSTEM_API_DIR ?= api/system/v1alpha1
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

# deletes all files in policy directory except *.proto and validator.go
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
	$(POLICY_GEN) openapi --plugin-dir $(POLICIES_DIR)/$* --yq-bin $(YQ) --openapi-template-path=$(TOOLS_DIR)/openapi/templates/endpoints.yaml --jsonschema-template-path=$(TOOLS_DIR)/openapi/templates/schema.yaml --gomodule $(GO_MODULE)
	@echo "Policy $* successfully generated"

generate/policy-import:
	./tools/policy-gen/generate-policy-import.sh $(GO_MODULE) $(POLICIES_DIR) $(policies)

generate/policy-config:
	./tools/policy-gen/generate-policy-config.sh  $(POLICIES_DIR) $(policies)

generate/policy-defaults:
	./tools/policy-gen/generate-policy-defaults.sh $(KUMA_DIR) $(kuma_policies) $(policies)

generate/policy-helm:
	PATH=$(CI_TOOLS_BIN_DIR):$$PATH $(TOOLS_DIR)/policy-gen/generate-policy-helm.sh $(HELM_VALUES_FILE) $(HELM_CRD_DIR) $(HELM_VALUES_FILE_POLICY_PATH) $(POLICIES_DIR) $(policies)

# Discover OpenAPI specification files
# - Searches api/openapi/specs/ and immediate subdirectories for *.yaml files
# - Excludes kri/ subdirectory (handled separately by oapi-gen tool)
# - Allows additional exclusions via OAS_SPECS_EXTRA_FILTER (for downstream projects)
# - Sorts results for deterministic ordering
OAS_SPECS_EXTRA_FILTER ?=
OAS_SPECS := $(sort $(filter-out api/openapi/specs/kri/% $(OAS_SPECS_EXTRA_FILTER), \
	$(wildcard api/openapi/specs/*.yaml) \
	$(wildcard api/openapi/specs/*/*.yaml)))

# Function: Map OpenAPI spec path to generated Go types file path
# Transforms: api/openapi/specs/common/error.yaml
#         →  api/openapi/types/common/zz_generated.error.go
#
# Transformation steps:
#   1. $(dir $(1))                    → api/openapi/specs/common/
#   2. patsubst specs/% → %           → api/openapi/common/
#   3. basename $(notdir $(1))        → error (removes path and extension)
#   4. Construct final path with zz_generated. prefix
define OAS_OUT
api/openapi/types/$(patsubst api/openapi/specs/%,%,$(dir $(1)))zz_generated.$(basename $(notdir $(1))).go
endef

# Compute all generated Go type files corresponding to discovered specs
# This list becomes a prerequisite for generate/oas, ensuring all files are built
OAS_TYPES := $(foreach s,$(OAS_SPECS),$(call OAS_OUT,$(s)))

# Pattern rule: Create output directories on demand
# This is used as an order-only prerequisite (see OAS_RULE below)
api/openapi/types%/: ; @mkdir -p $@

# Template: Define a rule pattern for generating Go types from one OpenAPI spec
#
# Parameters:
#   $(1) - The OpenAPI spec file path (e.g., api/openapi/specs/common/error.yaml)
#
# Generated rule structure:
#   <output-file>: <spec-file> <config-file> | <output-dir>
#       @$(OAPI_CODEGEN) -config <config> -o <output> <spec>
#
# Make syntax details:
#   $$(...)           - Double-$ escapes variable expansion until rule instantiation
#   $$(@D)            - Automatic variable: directory part of target (output file)
#   $$@               - Automatic variable: full target name (output file)
#   $$<               - Automatic variable: first prerequisite (spec file)
#   | $$(@D)          - Order-only prerequisite: ensure directory exists, but don't
#                       rebuild if directory timestamp changes (avoids spurious rebuilds)
define OAS_RULE
$(call OAS_OUT,$(1)): $(1) api/openapi/openapi.cfg.yaml | $$(@D)
	@$$(OAPI_CODEGEN) -config api/openapi/openapi.cfg.yaml -o $$@ $$<
endef

# Instantiate one concrete rule per OpenAPI spec
# How it works:
#   1. foreach iterates over each spec in OAS_SPECS
#   2. OAS_RULE template is expanded with the spec path as $(1)
#   3. eval processes the expanded text as Make syntax, creating actual rules
#
# Example: if OAS_SPECS contains "api/openapi/specs/common/resource.yaml", this creates:
#   api/openapi/types/common/zz_generated.resource.go: api/openapi/specs/common/resource.yaml api/openapi/openapi.cfg.yaml | api/openapi/types/common/
#       @$(OAPI_CODEGEN) -config api/openapi/openapi.cfg.yaml -o api/openapi/types/common/zz_generated.resource.go api/openapi/specs/common/resource.yaml
$(foreach s,$(OAS_SPECS),$(eval $(call OAS_RULE,$(s))))

.PHONY: generate/oas
generate/oas: $(GENERATE_OAS_PREREQUISITES) $(RESOURCE_GEN) $(OAPI_GEN) $(OAS_TYPES)
	@$(RESOURCE_GEN) -package mesh   -generator openapi -readDir $(KUMA_DIR) -writeDir .
	@$(RESOURCE_GEN) -package system -generator openapi -readDir $(KUMA_DIR) -writeDir .
	@$(OAPI_GEN) kri

.PHONY: validate/openapi-generated-docs
validate/openapi-generated-docs:
	@schema=docs/generated/openapi.yaml; \
	if [ ! -f $$schema ]; then \
		echo "Error: $$schema not found. Run 'make docs/generated/openapi.yaml' first."; \
		exit 1; \
	fi; \
	mkdir -p $(BUILD_DIR); \
	tmp_file=$$(mktemp $(BUILD_DIR)/openapi-validate.XXXXXX.go); \
	if ! $(OAPI_CODEGEN) -config api/openapi/openapi.cfg.yaml -o $$tmp_file $$schema; then \
		rm -f $$tmp_file; \
		exit 1; \
	fi; \
	rm -f $$tmp_file

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
	$(GO) list github.com/envoyproxy/go-control-plane/... | grep "github.com/envoyproxy/go-control-plane/envoy/" | awk '{printf "\t_ \"%s\"\n", $$1}' >> ${ENVOY_IMPORTS}
	echo ')' >> ${ENVOY_IMPORTS}

.PHONY: api-lint/policies
api-lint/policies:
	$(GO) run $(TOOLS_DIR)/ci/api-linter/main.go $$(find ./$(POLICIES_DIR)/*/api/v1alpha1 -type d -maxdepth 0 | sed 's|^|$(GO_MODULE)/|')

.PHONY: api-lint/resources
api-lint/resources:
	$(GO) run $(TOOLS_DIR)/ci/api-linter/main.go $$(find ./$(RESOURCES_DIR)/*/api/v1alpha1 -type d -maxdepth 0 | sed 's|^|$(GO_MODULE)/|')

.PHONY: api-lint
api-lint: api-lint/policies api-lint/resources
