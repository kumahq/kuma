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
	find . -path "*/.git/*" -prune -o -type f -name "*.sh" -exec $(SHELLCHECK) -P SCRIPTDIR -x {} +

.PHONY: golangci-lint
golangci-lint: ## Dev: Runs golangci-lint linter
ifndef CI
	GOMEMLIMIT=7GiB $(GOENV) $(GOLANGCI_LINT) run --timeout=10m -v
else
	@echo "skipping golangci-lint as it's done as a github action"
endif

.PHONY: fmt/ci
fmt/ci:
	$(YQ) -i '.env.K8S_MIN_VERSION = "$(K8S_MIN_VERSION)" | .env.K8S_MAX_VERSION = "$(K8S_MAX_VERSION)"' .github/workflows/"$(ACTION_PREFIX)"_test.yaml
	grep -r "golangci/golangci-lint-action" .github/workflows --include \*ml | cut -d ':' -f 1 | xargs -n 1 $(YQ) -i '(.jobs.* | select(. | has("steps")) | .steps[] | select(.uses == "golangci/golangci-lint-action*") | .with.version) |= "$(GOLANGCI_LINT_VERSION)"'

.PHONY: helm-lint
helm-lint:
	find ./deployments/charts -maxdepth 1 -mindepth 1 -type d -exec $(HELM) lint --strict {} \;

.PHONY: ginkgo/unfocus
ginkgo/unfocus:
	@$(GINKGO) unfocus

.PHONY: ginkgo/lint
ginkgo/lint:
	go run $(TOOLS_DIR)/ci/check_test_files.go

.PHONY: format/common
format/common: generate docs tidy ginkgo/unfocus fmt/ci

.PHONY: format
format: fmt/proto fmt/ci format/common

.PHONY: kube-lint
kube-lint:
	@find ./deployments/charts -maxdepth 1 -mindepth 1 -type d -exec $(KUBE_LINTER) lint {} \;
	@if [ -d ./app/kumactl/cmd/install/testdata ]; then \
		find ./app/kumactl/cmd/install/testdata -maxdepth 1 -type f -name 'install-control-plane*.golden.yaml' -exec $(KUBE_LINTER) lint {} +; \
	fi
	@if [ -d ./app/kumactl/cmd/install/testdata/install-cp-helm ]; then \
		find ./app/kumactl/cmd/install/testdata/install-cp-helm -maxdepth 1 -type f -name '*.golden.yaml' -exec $(KUBE_LINTER) lint {} +; \
	fi

.PHONY: hadolint
hadolint:
	@if [ `find ./tools/releases/dockerfiles/ -type f -iname "*Dockerfile" | grep -vc dockerignore` -eq 0 ]; then \
	  echo "No Dockerfiles found, exiting with failure."; \
	  exit 1; \
	fi; \
	find ./tools/releases/dockerfiles/ -type f -iname "*Dockerfile" | grep -v dockerignore | xargs -I {} $(HADOLINT) {}

.PHONY: lint
lint: helm-lint golangci-lint shellcheck kube-lint hadolint ginkgo/lint

.PHONY: check
check: format lint ## Dev: Run code checks (go fmt, go vet, ...)
	@untracked() { git ls-files --other --directory --exclude-standard --no-empty-directory; }; \
	check-changes() { git --no-pager diff "$$@"; }; \
	if [ $$(untracked | wc -l) -gt 0 ]; then \
		FAILED=true; \
		echo "The following files are untracked:"; \
		untracked; \
	fi; \
	if [ $$(check-changes --name-only | wc -l) -gt 0 ]; then \
		FAILED=true; \
		echo "The following changes (result of code generators and code checks) have been detected:"; \
		check-changes; \
	fi; \
	if [ "$$FAILED" = true ]; then exit 1; fi

.PHONY: update-vulnerable-dependencies
update-vulnerable-dependencies:
	@$(KUMA_DIR)/tools/ci/update-vulnerable-dependencies/update-vulnerable-dependencies.sh
