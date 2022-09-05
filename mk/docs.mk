DESTDIR := docs/generated

.PHONY: docs
docs: helm-docs ## Dev: Generate local documentation
	rm -rf $(DESTDIR)
	@$(MAKE) docs/install/markdown DESTDIR=$(DESTDIR)/cmd
	@$(MAKE) docs/install/resources DESTDIR=$(DESTDIR)/resources

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS) -s="file" --chart-search-root=./deployments/charts

.PHONY: docs/install/markdown
docs/install/markdown: DESTDIR:=$(DESTDIR)/markdown
docs/install/markdown: ## Generate CLI reference in markdown format
	@DESTDIR=$(DESTDIR) FORMAT=markdown go run ./tools/docs/generate.go

.PHONY: docs/install/manpages
docs/install/manpages: DESTDIR:=$(DESTDIR)/manpages
docs/install/manpages: ## Generate CLI reference in man(1) format
	@DESTDIR=$(DESTDIR) FORMAT=man $(GO_RUN) ./tools/docs/generate.go

.PHONY: docs/install/resources
docs/install/resources: DESTDIR:=$(DESTDIR)/resources
docs/install/resources: ## Generate Mesh API reference
	mkdir -p $(DESTDIR) && cd api/ && $(PROTOC) \
		--kumadoc_out=../$(DESTDIR) \
		--plugin=protoc-gen-kumadoc=$(PROTOC_GEN_KUMADOC) \
		mesh/v1alpha1/*.proto
