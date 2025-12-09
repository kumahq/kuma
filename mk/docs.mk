DOCS_PROTOS ?= api/mesh/v1alpha1/*.proto
DOCS_CP_CONFIG ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml
DOCS_EXTRA_TARGETS ?=
DOCS_OPENAPI_PREREQUISITES ?=

.PHONY: clean/docs
clean/docs:
	rm -rf docs/generated

.PHONY: docs
docs: helm-docs docs/generated/raw docs/generated/openapi.yaml $(DOCS_EXTRA_TARGETS) ## Dev: Generate local documentation

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS) -s="file" --chart-search-root=./deployments/charts

.PHONY: docs/generated/raw
docs/generated/raw: docs/generated/raw/rbac.yaml
	mkdir -p $@
	cp $(DOCS_CP_CONFIG) $@/kuma-cp.yaml
	cp $(HELM_VALUES_FILE) $@/helm-values.yaml

	mkdir -p $@/crds
	for f in $$(find deployments/charts -name '*.yaml' | grep '/crds/'); do cp $$f $@/crds/; done

	mkdir -p $@/protos
	$(PROTOC) \
		--jsonschema_out=$@/protos \
		--plugin=protoc-gen-jsonschema=$(PROTOC_GEN_JSONSCHEMA) \
		$(DOCS_PROTOS)

.PHONY: docs/generated/raw/rbac.yaml
docs/generated/raw/rbac.yaml:
	@mkdir -p docs/generated/raw
	@$(HELM) template --namespace $(PROJECT_NAME)-system $(PROJECT_NAME) deployments/charts/$(PROJECT_NAME) | \
	$(YQ) eval-all 'select((.kind == "ClusterRole" or .kind == "ClusterRoleBinding" or .kind == "Role" or .kind == "RoleBinding") and (.metadata.annotations["helm.sh/hook"] == null)) | del(.metadata.labels)' - | \
	grep -Ev '^\s*#' | \
	sed 's/[[:space:]]*#.*$$//' > $@

OAPI_TMP_DIR ?= $(BUILD_DIR)/oapitmp
API_DIRS     ?= $(TOP)/api/openapi/specs:base

# Generate a consolidated OpenAPI spec consumed by docs
# Keep prep and generation separate for clarity and easier maintenance
# Prep step normalizes input specs into a predictable temp layout
# Generation step runs the OpenAPI tool container against that layout

# Ensure the output directory for generated docs artifacts exists
docs/generated:
	mkdir -p $@

.PHONY: docs/generated/openapi.yaml
docs/generated/openapi.yaml: $(DOCS_OPENAPI_PREREQUISITES) | docs/generated docs/generated/openapi/prepare/specs
	@echo "Rewriting /specs/ paths in all YAML files..."
	@mkdir -p $(BUILD_DIR)/oapitmp-rewritten
	@cd $(OAPI_TMP_DIR) && find . -name '*.yaml' | while read f; do \
		mkdir -p $(BUILD_DIR)/oapitmp-rewritten/$$(dirname $$f); \
		sed 's|"/specs/|"$(BUILD_DIR)/oapitmp-rewritten/|g; s|'"'"'/specs/|'"'"'$(BUILD_DIR)/oapitmp-rewritten/|g; s| /specs/| $(BUILD_DIR)/oapitmp-rewritten/|g; s|: /specs/|: $(BUILD_DIR)/oapitmp-rewritten/|g' $$f > $(BUILD_DIR)/oapitmp-rewritten/$$f; \
	done
	@echo "Bundling individual OpenAPI specs..."
	@mkdir -p $(BUILD_DIR)/openapi-bundled
	@cd $(BUILD_DIR)/oapitmp-rewritten && \
		find . \( -name 'rest.yaml' -o -name 'api.yaml' -o -name 'kri.yaml' \) | while read f; do \
			mkdir -p $(BUILD_DIR)/openapi-bundled/$$(dirname $$f); \
			mise exec -- redocly bundle $$f -o $(BUILD_DIR)/openapi-bundled/$$f || echo "Skipping $$f"; \
		done
	@echo "Merging all bundled specs..."
	@mise exec -- oas-toolkit merge $$(find $(BUILD_DIR)/openapi-bundled -name '*.yaml' | sort) > $@
	@$(MAKE) --no-print-directory validate/openapi-generated-docs

# Prepare $(OAPI_TMP_DIR) with a normalized directory layout for the generator
# Layout
#   $(OAPI_TMP_DIR)/
#     base/            <- contents of $(TOP)/api/openapi/specs listed in API_DIRS
#     policies/<name>/ <- policy REST or OpenAPI fragments from $(POLICIES_DIR)
#     resources/<name>/
#     protoresources/<name>/
# Split into sub-targets to keep steps clear and maintainable

# Aggregate prep target that ensures all parts are ready
.PHONY: docs/generated/openapi/prepare/specs
docs/generated/openapi/prepare/specs: \
	docs/generated/openapi/prepare/base \
	docs/generated/openapi/prepare/policies \
	docs/generated/openapi/prepare/resources \
	docs/generated/openapi/prepare/protoresources

# Create or reset the top-level temp layout under $(OAPI_TMP_DIR)
.PHONY: docs/generated/openapi/prepare/layout
docs/generated/openapi/prepare/layout:
	@rm -rf $(OAPI_TMP_DIR)
	@mkdir -p $(OAPI_TMP_DIR)

# Create or reset a named subdirectory under $(OAPI_TMP_DIR)
.PHONY: docs/generated/openapi/prepare/layout/%
docs/generated/openapi/prepare/layout/%:
	@rm -rf $(OAPI_TMP_DIR)/$*
	@mkdir -p $(OAPI_TMP_DIR)/$*

# Copy base API specs into well-known subdirs
# The dst path is the part after the colon in API_DIRS entries
.PHONY: docs/generated/openapi/prepare/base
docs/generated/openapi/prepare/base: docs/generated/openapi/prepare/layout
	@for i in $(API_DIRS); do \
		src=$$(echo "$$i" | cut -d: -f1); dst=$$(echo "$$i" | cut -d: -f2); \
		mkdir -p "$(OAPI_TMP_DIR)/$${dst}"; \
		cp -R "$$src" "$(OAPI_TMP_DIR)/$${dst}"; \
	done

# Helper macro to collect YAML specs into $(OAPI_TMP_DIR)/<subdir>/<name>/
# $(1) shell command that prints file paths, for example a find invocation
# $(2) shell snippet that echoes the name component for $$i
# $(3) destination subdir under $(OAPI_TMP_DIR) such as policies or resources
define OAPI_COLLECT
	@for i in $$($(1)); do \
		name=$$($(2)); \
		dst="$(OAPI_TMP_DIR)/$(3)/$$name"; \
		mkdir -p "$$dst"; \
		cp "$$i" "$$dst/$$(basename "$$i")"; \
	done
endef

# Gather policy API YAMLs into policies/<policyName>/
.PHONY: docs/generated/openapi/prepare/policies
docs/generated/openapi/prepare/policies: docs/generated/openapi/prepare/layout/policies
	$(call OAPI_COLLECT,find $(POLICIES_DIR) -path '*/api/*.yaml' -not -path '*/testdata/*',basename $${i%/api/*},policies)

# Gather resource API YAMLs into resources/<resourceName>/
.PHONY: docs/generated/openapi/prepare/resources
docs/generated/openapi/prepare/resources: docs/generated/openapi/prepare/layout/resources
	$(call OAPI_COLLECT,find $(RESOURCES_DIR) -path '*/api/*.yaml' -not -path '*/testdata/*',basename $${i%/api/*},resources)

# Gather proto-backed mesh API YAMLs into protoresources/<name>/
.PHONY: docs/generated/openapi/prepare/protoresources
docs/generated/openapi/prepare/protoresources: docs/generated/openapi/prepare/layout/protoresources
	$(call OAPI_COLLECT,find $(MESH_API_DIR) -name '*.yaml',basename $${i%/*},protoresources)
	$(call OAPI_COLLECT,find $(SYSTEM_API_DIR) -name '*.yaml',basename $${i%/*},protoresources)
