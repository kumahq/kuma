DOCS_PROTOS ?= api/mesh/v1alpha1/*.proto
DOCS_CP_CONFIG ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml

.PHONY: clean/docs
clean/docs:
	rm -rf docs/generated

.PHONY: docs
docs: docs/generated/cmd docs/generated/kuma-cp.md docs/generated/resources helm-docs ## Dev: Generate local documentation

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS) -s="file" --chart-search-root=./deployments/charts

DOCS_CMD_FORMAT ?= markdown
.PHONY: docs/generated/cmd
docs/generated/cmd:
	DESTDIR=$@ FORMAT=$(DOCS_CMD_FORMAT) go run $(TOOLS_DIR)/docs/generate.go

<<<<<<< HEAD
.PHONY: docs/install/manpages
docs/install/manpages: DESTDIR:=$(DESTDIR)/manpages
docs/install/manpages: ## Generate CLI reference in man(1) format
	@DESTDIR=$(DESTDIR) FORMAT=man $(GO_RUN) $(TOOLS_DIR)/docs/generate.go

.PHONY: docs/install/resources
docs/install/resources: DESTDIR:=$(DESTDIR)/resources
docs/install/resources: PROTOS=api/mesh/v1alpha1/*.proto
docs/install/resources: ## Generate Mesh API reference
	mkdir -p $(DESTDIR) && $(PROTOC) \
		--kumadoc_out=$(DESTDIR) \
=======
.PHONY: docs/generated/resources
docs/generated/resources: ## Generate Mesh API reference
	mkdir -p $@ && $(PROTOC) \
		--kumadoc_out=$@ \
>>>>>>> 518eb25f3 (ci(docs): simplify and add kuma-cp.md config docs (#6490))
		--plugin=protoc-gen-kumadoc=$(PROTOC_GEN_KUMADOC) \
		$(DOCS_PROTOS)

.PHONY: docs/generated/kuma-cp.md
docs/generated/kuma-cp.md: ## Generate Mesh API reference
	@mkdir -p $(@D)
	@echo "# Control-Plane configuration" > $@
	@echo "Here are all options to configure the control-plane:" >> $@
	@echo '```yaml' >> $@
	@cat $(DOCS_CP_CONFIG) >> $@
	@echo '```' >> $@
