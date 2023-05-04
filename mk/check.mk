.PHONY: fmt
fmt: golangci-lint-fmt fmt/proto fmt/ci ## Dev: Run various format tools

.PHONY: fmt/proto
fmt/proto: ## Dev: Run clang-format on .proto files
	find . -name '*.proto' | xargs -L 1 $(CLANG_FORMAT) -i

.PHONY: tidy
tidy:
	go mod edit -go=$(shell echo $(GO_VERSION) | grep -Eo '[0-9]\.[0-9]+')
	@TOP=$(shell pwd) && \
	for m in $$(find . -name go.mod) ; do \
		( cd $$(dirname $$m) && go mod tidy ) ; \
	done

.PHONY: shellcheck
shellcheck:
	find . -name "*.sh" -not -path "./.git/*" -exec $(SHELLCHECK) -P SCRIPTDIR -x {} +

.PHONY: golangci-lint
golangci-lint: ## Dev: Runs golangci-lint linter
	GOMEMLIMIT=7GiB $(GOENV) $(GOLANGCI_LINT) run --timeout=10m -v

.PHONY: golangci-lint-fmt
golangci-lint-fmt:
	GOMEMLIMIT=7GiB $(GOENV) $(GOLANGCI_LINT) run --timeout=10m -v \
		--disable-all \
		--enable gofumpt

.PHONY: fmt/ci
fmt/ci:
	$(CI_TOOLS_BIN_DIR)/yq -i '.parameters.go_version.default = "$(GO_VERSION)" | .parameters.first_k8s_version.default = "$(K8S_MIN_VERSION)" | .parameters.last_k8s_version.default = "$(K8S_MAX_VERSION)"' .circleci/config.yml
	find .github/workflows -name '*ml' | xargs -n 1 $(CI_TOOLS_BIN_DIR)/yq -i '(.jobs.* | select(. | has("steps")) | .steps[] | select(.uses == "actions/setup-go*") | .with.go-version) |= "$(GO_VERSION)"'

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
format: fmt format/common

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
	find ./tools/releases/dockerfiles/ -type f -iname "Dockerfile*" | grep -v dockerignore | xargs -I {} $(HADOLINT) {}

.PHONY: lint
lint: helm-lint golangci-lint shellcheck kube-lint hadolint ginkgo/lint

.PHONY: check
check: format/common lint ## Dev: Run code checks (go fmt, go vet, ...)
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
