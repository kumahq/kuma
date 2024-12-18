DOCS_PROTOS ?= api/mesh/v1alpha1/*.proto
DOCS_CP_CONFIG ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml
DOCS_EXTRA_TARGETS ?=

.PHONY: clean/docs
clean/docs:
	rm -rf docs/generated

.PHONY: docs
docs: helm-docs docs/generated/raw docs/generated/openapi.yaml $(DOCS_EXTRA_TARGETS) ## Dev: Generate local documentation

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS) -s="file" --chart-search-root=./deployments/charts

.PHONY: docs/generated/raw
docs/generated/raw:
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

OAPI_TMP_DIR ?= $(BUILD_DIR)/oapitmp
API_DIRS="$(TOP)/api/openapi/specs:base"

.PHONY: docs/generated/openapi.yaml
docs/generated/openapi.yaml:
	rm -rf $(OAPI_TMP_DIR)
	mkdir -p $(dir $@)
	mkdir -p $(OAPI_TMP_DIR)/policies
	mkdir -p $(OAPI_TMP_DIR)/resources
	mkdir -p $(OAPI_TMP_DIR)/protoresources
	for i in $(API_DIRS); do mkdir -p $(OAPI_TMP_DIR)/$$(echo $${i} | cut -d: -f2); cp -R $$(echo $${i} | cut -d: -f1) $(OAPI_TMP_DIR)/$$(echo $${i} | cut -d: -f2); done
	for i in $$( find $(POLICIES_DIR) -name '*.yaml' | grep '/api/' | grep -v '/testdata/'); do DIR=$(OAPI_TMP_DIR)/policies/$$(echo $${i} | awk -F/ '{print $$(NF-3)}'); mkdir -p $${DIR}; cp $${i} $${DIR}/$$(echo $${i} | awk -F/ '{print $$(NF)}'); done
	for i in $$( find $(RESOURCES_DIR) -name '*.yaml' | grep '/api/' | grep -v '/testdata/'); do DIR=$(OAPI_TMP_DIR)/resources/$$(echo $${i} | awk -F/ '{print $$(NF-3)}'); mkdir -p $${DIR}; cp $${i} $${DIR}/$$(echo $${i} | awk -F/ '{print $$(NF)}'); done
	for i in $$( find $(MESH_API_DIR) -name '*.yaml'); do DIR=$(OAPI_TMP_DIR)/protoresources/$$(echo $${i} | awk -F/ '{print $$(NF-1)}'); mkdir -p $${DIR}; cp $${i} $${DIR}/$$(echo $${i} | awk -F/ '{print $$(NF)}'); done

ifdef BASE_API
	docker run --rm -v $$PWD/$(dir $(BASE_API)):/base -v $(OAPI_TMP_DIR):/specs ghcr.io/kumahq/openapi-tool:v0.12.0 generate /base/$(notdir $(BASE_API)) '/specs/**/*.yaml'  '!/specs/kuma/**' > $@
else
	docker run --rm -v $(OAPI_TMP_DIR):/specs ghcr.io/kumahq/openapi-tool:v0.12.0 generate '/specs/**/*.yaml' > $@
endif
