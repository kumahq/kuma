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
	$(GOENV) $(GOLANGCI_LINT) run \
		--disable-all \
<<<<<<< HEAD
		--enable bodyclose,contextcheck,errcheck,gci,gocritic,gofmt,gomodguard,govet,importas,ineffassign,misspell,typecheck,unconvert,unparam,whitespace \
		--timeout=10m -v
	$(GOENV) $(GOLANGCI_LINT) run \
		--disable-all \
		--enable gosimple,staticcheck,unused \
		--timeout=10m -v
=======
		--enable gofumpt

.PHONY: fmt/ci
fmt/ci:
	$(CI_TOOLS_BIN_DIR)/yq -i '.parameters.go_version.default = "$(GO_VERSION)" | .parameters.first_k8s_version.default = "$(K8S_MIN_VERSION)" | .parameters.last_k8s_version.default = "$(K8S_MAX_VERSION)"' .circleci/config.yml
	find .github/workflows -name '*ml' | xargs -n 1 $(CI_TOOLS_BIN_DIR)/yq -i '(.jobs.* | select(. | has("steps")) | .steps[] | select(.uses == "golangci/golangci-lint-action*") | .with.version) |= "$(GOLANGCI_LINT_VERSION)"'
>>>>>>> b9f215e7c (build(mk): use go.mod as source of truth for go version in makefiles (#7843))

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

.PHONY: format
format: fmt generate docs tidy ginkgo/unfocus

.PHONY: kube-lint
kube-lint:
	$(KUBE_LINTER) lint .

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

.PHONY: update-vulnerable-dependencies
update-vulnerable-dependencies:
	@$(KUMA_DIR)/tools/ci/update-vulnerable-dependencies.sh
