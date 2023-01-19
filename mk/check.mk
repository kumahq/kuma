.PHONY: fmt
fmt: fmt/go fmt/proto ## Dev: Run various format tools

.PHONY: fmt/go
fmt/go: ## Dev: Run go fmt
	go fmt ./...

.PHONY: fmt/proto
fmt/proto: ## Dev: Run clang-format on .proto files
	find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT) -i

.PHONY: tidy
tidy:
	@TOP=$(shell pwd) && \
	for m in $$(find . -name go.mod) ; do \
		( cd $$(dirname $$m) && go mod tidy ) ; \
	done

.PHONY: shellcheck
shellcheck:
	find . -name "*.sh" -not -path "./.git/*" -exec $(SHELLCHECK) -P SCRIPTDIR -x {} +

.PHONY: golangci-lint
golangci-lint: ## Dev: Runs golangci-lint linter
	GOMEMLIMIT=7GiB $(GOLANGCI_LINT) run --timeout=10m -v

.PHONY: helm-lint
helm-lint:
	for c in ./deployments/charts/*; do \
  		if [ -d $$c ]; then \
			$(HELM) lint --strict $$c; \
		fi \
	done

.PHONY: ginkgo/unfocus
ginkgo/unfocus:
	@$(GINKGO) unfocus

.PHONY: ginkgo/lint
ginkgo/lint:
	go run $(TOOLS_DIR)/ci/check_test_files.go

.PHONY: format
format: fmt generate docs tidy ginkgo/unfocus ginkgo/lint

.PHONY: kube-lint
kube-lint:
	$(KUBE_LINTER) lint .

.PHONY: hadolint
hadolint:
	find ./tools/releases/dockerfiles/ -type f -iname "Dockerfile*" | grep -v dockerignore | xargs -I {} $(HADOLINT) {}

.PHONY: check
check: format helm-lint golangci-lint shellcheck kube-lint hadolint ## Dev: Run code checks (go fmt, go vet, ...)
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
