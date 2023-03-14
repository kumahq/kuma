CLANG_FORMAT_PATH ?= clang-format

.PHONY: fmt
fmt: fmt/go fmt/proto ## Dev: Run various format tools

.PHONY: fmt/go
fmt/go: ## Dev: Run go fmt
	go fmt $(GOFLAGS) ./...

.PHONY: fmt/proto
fmt/proto: ## Dev: Run clang-format on .proto files
	which $(CLANG_FORMAT_PATH) && find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT_PATH) -i || true

.PHONY: tidy
tidy:
	@TOP=$(shell pwd) && \
	for m in $$(find . -name go.mod) ; do \
		( cd $$(dirname $$m) && go mod tidy ) ; \
	done

.PHONY: shellcheck
shellcheck:
	find . -name "*.sh" -not -path "./.git/*" -exec shellcheck -P SCRIPTDIR -x {} +

.PHONY: golangci-lint
golangci-lint: ## Dev: Runs golangci-lint linter
	$(GOLANGCI_LINT_DIR)/golangci-lint run --timeout=10m -v

.PHONY: helm-lint
helm-lint:
	for c in ./deployments/charts/*; do \
  		if [ -d $$c ]; then \
			helm lint --strict $$c; \
		fi \
	done

.PHONY: ginkgo/unfocus
ginkgo/unfocus:
	$(GOPATH_BIN_DIR)/ginkgo unfocus

.PHONY: format
format: fmt generate docs tidy ginkgo/unfocus

.PHONY: kube-lint
kube-lint:
	$(GOPATH_BIN_DIR)/kube-linter lint .

.PHONY: check
check: format helm-lint golangci-lint shellcheck kube-lint ## Dev: Run code checks (go fmt, go vet, ...)
	# fail if Git working tree is dirty or there are untracked files
	git diff --quiet || \
	git ls-files --other --directory --exclude-standard --no-empty-directory | wc -l | read UNTRACKED_FILES; if [ "$$UNTRACKED_FILES" != "0" ]; then false; fi || \
	test $$(git diff --name-only | wc -l) -eq 0 || \
	( \
		echo "The following changes (result of code generators and code checks) have been detected:" && \
		git --no-pager diff && \
		echo "The following files are untracked:" && \
		git ls-files --other --directory --exclude-standard --no-empty-directory && \
		false \
	)
