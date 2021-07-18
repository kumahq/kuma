CLANG_FORMAT_PATH ?= clang-format

.PHONY: fmt
fmt: fmt/go fmt/proto ## Dev: Run various format tools

.PHONY: fmt/go
fmt/go: ## Dev: Run go fmt
	go fmt $(GOFLAGS) ./...

.PHONY: fmt/proto
fmt/proto: ## Dev: Run clang-format on .proto files
	which $(CLANG_FORMAT_PATH) && find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT_PATH) -i || true

.PHONY: vet
vet: ## Dev: Run go vet
	go vet $(GOFLAGS) ./...

.PHONY: tidy
tidy:
	@TOP=$(shell pwd) && \
	for m in . ./api/ ; do \
		cd $$m ; \
		rm go.sum ; \
		go mod tidy ; \
		cd $$TOP; \
	done

.PHONY: golangci-lint
golangci-lint: ## Dev: Runs golangci-lint linter
	$(GOLANGCI_LINT_DIR)/golangci-lint run --timeout=10m -v

.PHONY: helm-lint
helm-lint:
	for c in ./deployments/charts/*; do \
  		if [ -d $$c ]; then \
			helm lint $$c; \
		fi \
	done

.PHONY: imports
imports: ## Dev: Runs goimports in order to organize imports
	goimports -w -local github.com/kumahq/kuma -d `find . -type f -name '*.go' -not -name '*.pb.go' -not -path './vendored/*'`

.PHONY: helm-docs
helm-docs: ## Dev: Runs helm-docs generator
	$(HELM_DOCS_PATH) -s="file" --chart-search-root=./deployments/charts

.PHONY: ginkgo/unfocus
ginkgo/unfocus:
	$(GOPATH_BIN_DIR)/ginkgo unfocus

.PHONY: check
check: generate fmt vet docs helm-lint golangci-lint imports tidy helm-docs ginkgo/unfocus ## Dev: Run code checks (go fmt, go vet, ...)
	$(MAKE) generate manifests -C pkg/plugins/resources/k8s/native
	git diff --quiet || test $$(git diff --name-only | grep -v -e 'go.mod$$' -e 'go.sum$$' | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty
