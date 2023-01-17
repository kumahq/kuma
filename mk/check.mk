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
	# Starting with golangci-lint v1.47.1, the CI job runs OOM if all of these
	# linters are used together. The first set is the largest that ran without
	# OOM.
	GOMEMLIMIT=7GiB $(GOLANGCI_LINT) run \
		--disable-all \
		--enable bodyclose,contextcheck,errcheck,gci,gocritic,gofmt,gomodguard,govet,importas,ineffassign,misspell,typecheck,unconvert,unparam,whitespace \
		--timeout=10m -v
	GOMEMLIMIT=7GiB $(GOLANGCI_LINT) run \
		--disable-all \
		--enable gosimple,nonamedreturns,staticcheck,unused \
		--timeout=10m -v

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
	git diff --quiet || test $$(git diff --name-only | wc -l) -eq 0 || ( echo "The following changes (result of code generators and code checks) have been detected:" && git --no-pager diff && false ) # fail if Git working tree is dirty
