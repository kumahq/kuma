.PHONY: docs
docs: ## Dev: Generate all docs
docs: docs/install/markdown docs/install/manpages docs/install/protobuf

.PHONY: docs/install/markdown
docs/install/markdown: DESTDIR ?= docs/markdown
docs/install/markdown: ## Generate CLI reference in markdown format
	@DESTDIR=$(DESTDIR) FORMAT=markdown go run ./tools/docs/generate.go

.PHONY: docs/install/manpages
docs/install/manpages: DESTDIR ?= docs/manpages
docs/install/manpages: ## Generate CLI reference in man(1) format
	@echo target $@ not implemented

.PHONY: docs/install/protobuf
docs/install/protobuf: DESTDIR ?= docs/protobuf
docs/install/protobuf: ## Generate protobuf API reference
	@echo target $@ not implemented
