DOCS_PROTOS ?= api/mesh/v1alpha1/*.proto
DOCS_CP_CONFIG ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml

.PHONY: clean/docs
clean/docs:
	rm -rf docs/generated

.PHONY: docs
docs: docs/generated/cmd docs/generated/kuma-cp.md docs/generated/resources helm-docs docs/generated/raw ## Dev: Generate local documentation

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS_PATH) -s="file" --chart-search-root=./deployments/charts

DOCS_CMD_FORMAT ?= markdown
.PHONY: docs/generated/cmd
docs/generated/cmd:
	DESTDIR=$@ FORMAT=$(DOCS_CMD_FORMAT) go run $(TOOLS_DIR)/docs/generate.go

.PHONY: docs/generated/resources
docs/generated/resources: ## Generate Mesh API reference
	mkdir -p $(@) && cd api/ && $(PROTOC) \
		--kumadoc_out=../$(@) \
		--plugin=protoc-gen-kumadoc=$(GOPATH_BIN_DIR)/protoc-gen-kumadoc \
		mesh/v1alpha1/*.proto

.PHONY: docs/generated/kuma-cp.md
docs/generated/kuma-cp.md: ## Generate Mesh API reference
	@mkdir -p $(@D)
	@echo "# Control-Plane configuration" > $@
	@echo "Here are all options to configure the control-plane:" >> $@
	@echo '```yaml' >> $@
	@cat $(DOCS_CP_CONFIG) >> $@
	@echo '```' >> $@

<<<<<<< HEAD
.PHONY: docs/output
docs/output: clean/docs/output | $(DOCS_OUTPUT_DIR)
	cp $(DOCS_CP_CONFIG) $(DOCS_OUTPUT_DIR)/kuma-cp.yaml

.PHONY: clean/docs/output
clean/docs/output:
	rm -rf $(DOCS_OUTPUT_DIR)

$(DOCS_OUTPUT_DIR):
	mkdir -p $@

=======
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
>>>>>>> fa652ae34 (ci(docs): persist generated docs in docs/generated/raw (#7170))
