UPDATE_GOLDEN_FILES ?=
GO_TEST := TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test $(GOFLAGS) $(LD_FLAGS)
GO_TEST_E2E := TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test -p 1 $(GOFLAGS) $(LD_FLAGS)
GO_TEST_OPTS ?=
PKG_LIST ?= ./...

BUILD_COVERAGE_DIR ?= $(BUILD_DIR)/coverage

COVERAGE_PROFILE := $(BUILD_COVERAGE_DIR)/coverage.out
COVERAGE_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage.html

# exports below are required for K8S unit tests
export TEST_ASSET_KUBE_APISERVER=$(KUBE_APISERVER_PATH)
export TEST_ASSET_ETCD=$(ETCD_PATH)
export TEST_ASSET_KUBECTL=$(KUBECTL_PATH)

.PHONY: test
test: ${COVERAGE_PROFILE} ## Dev: Run tests for all modules
	$(GO_TEST) $(GO_TEST_OPTS) --tags=gateway -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" $(PKG_LIST)
	$(MAKE) coverage

${COVERAGE_PROFILE}:
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"

.PHONY: coverage
coverage: ${COVERAGE_PROFILE}
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

.PHONY: test/kuma-cp
test/kuma-cp: PKG_LIST=./app/kuma-cp/... ./pkg/config/app/kuma-cp/...
test/kuma-cp: test/kuma ## Dev: Run `kuma-cp` tests only

.PHONY: test/kuma-dp
test/kuma-dp: PKG_LIST=./app/kuma-dp/... ./pkg/config/app/kuma-dp/...
test/kuma-dp: test/kuma ## Dev: Run `kuma-dp` tests only

.PHONY: test/kumactl
test/kumactl: PKG_LIST=./app/kumactl/... ./pkg/config/app/kumactl/...
test/kumactl: test/kuma ## Dev: Run `kumactl` tests only

.PHONY: test/release
test/release: # Dev: Run release tests
	$(GO_TEST) $(GO_TEST_OPTS) -tags=release ./test/release/...
<<<<<<< HEAD

.PHONY: integration
integration: ${COVERAGE_INTEGRATION_PROFILE} ## Dev: Run integration tests
	tools/test/run-integration-tests.sh '$(GO_TEST) -tags=integration,gateway -race -covermode=atomic -count=1 -coverpkg=./... -coverprofile=$(COVERAGE_INTEGRATION_PROFILE) $(PKG_LIST)'
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_INTEGRATION_PROFILE)" -o "$(COVERAGE_INTEGRATION_REPORT_HTML)"

=======
>>>>>>> f22bff87 (chore(ci) use testcontainers for postgres tests (#3065))
