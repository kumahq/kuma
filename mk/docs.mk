.PHONY: docs
docs: DESTDIR ?= docs/generated
docs: ## Dev: Generate local documentation
	rm -rf $(DESTDIR)
	@$(MAKE) docs/install/markdown DESTDIR=$(DESTDIR)/cmd
	@$(MAKE) docs/install/resources DESTDIR=$(DESTDIR)/resources

.PHONY: docs/install/markdown
docs/install/markdown: DESTDIR ?= docs/markdown
docs/install/markdown: ## Generate CLI reference in markdown format
	@DESTDIR=$(DESTDIR) FORMAT=markdown go run ./tools/docs/generate.go

.PHONY: docs/install/manpages
docs/install/manpages: DESTDIR ?= docs/manpages
docs/install/manpages: ## Generate CLI reference in man(1) format
	@DESTDIR=$(DESTDIR) FORMAT=man $(GO_RUN) ./tools/docs/generate.go

.PHONY: docs/install/resources
docs/install/resources: DESTDIR ?= docs/resources
docs/install/resources: ## Generate Mesh API reference
	mkdir -p $(DESTDIR) && cd api/ && $(PROTOC) \
		--kumadoc_out=../$(DESTDIR) \
		--plugin=protoc-gen-kumadoc=$(GOPATH_BIN_DIR)/protoc-gen-kumadoc \
		mesh/v1alpha1/*.proto
