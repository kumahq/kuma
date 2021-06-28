UPDATE_GOLDEN_FILES ?=
GO_TEST := UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test $(GOFLAGS) $(LD_FLAGS)
GO_TEST_E2E := UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test -p 1 $(GOFLAGS) $(LD_FLAGS)
GO_TEST_OPTS ?=
PKG_LIST ?= ./...

BUILD_COVERAGE_DIR ?= $(BUILD_DIR)/coverage

COVERAGE_PROFILE := $(BUILD_COVERAGE_DIR)/coverage.out
COVERAGE_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage.html

COVERAGE_INTEGRATION_PROFILE := $(BUILD_COVERAGE_DIR)/coverage-integration.out
COVERAGE_INTEGRATION_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage-integration.html

# exports below are required for K8S unit tests
export TEST_ASSET_KUBE_APISERVER=$(KUBE_APISERVER_PATH)
export TEST_ASSET_ETCD=$(ETCD_PATH)
export TEST_ASSET_KUBECTL=$(KUBECTL_PATH)

TEST_TARGETS ?= test/api test/k8s test/kuma

.PHONY: test
test: ${COVERAGE_PROFILE} ## Dev: Run tests for all modules
	for test_target in $(TEST_TARGETS);\
	do \
		$(MAKE) $$test_target || exit $$?; \
	done
	$(MAKE) coverage

${COVERAGE_PROFILE}:
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"

.PHONY: coverage
coverage: ${COVERAGE_PROFILE}
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

.PHONY: test/kuma
test/kuma: # Dev: Run tests for the module github.com/kumahq/kuma
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" $(PKG_LIST)

.PHONY: test/api
test/api: \
	MODULE=./api \
	COVERAGE_PROFILE=$(BUILD_COVERAGE_DIR)/coverage-api.out
test/api: test/module

.PHONY: test/k8s
test/k8s: \
	MODULE=./pkg/plugins/resources/k8s/native \
	COVERAGE_PROFILE=$(BUILD_COVERAGE_DIR)/coverage-k8s.out
test/k8s: test/module

.PHONY: test/module
test/module:
	GO_TEST='${GO_TEST}' GO_TEST_OPTS='${GO_TEST_OPTS}' COVERAGE_PROFILE='${COVERAGE_PROFILE}' $(MAKE) test -C ${MODULE}

.PHONY: test/kuma-cp
test/kuma-cp: PKG_LIST=./app/kuma-cp/... ./pkg/config/app/kuma-cp/...
test/kuma-cp: test/kuma ## Dev: Run `kuma-cp` tests only

.PHONY: test/kuma-dp
test/kuma-dp: PKG_LIST=./app/kuma-dp/... ./pkg/config/app/kuma-dp/...
test/kuma-dp: test/kuma ## Dev: Run `kuma-dp` tests only

.PHONY: test/kumactl
test/kumactl: PKG_LIST=./app/kumactl/... ./pkg/config/app/kumactl/...
test/kumactl: test/kuma ## Dev: Run `kumactl` tests only

${COVERAGE_INTEGRATION_PROFILE}:
	mkdir -p "$(shell dirname "$(COVERAGE_INTEGRATION_PROFILE)")"

.PHONY: integration
integration: ${COVERAGE_INTEGRATION_PROFILE} ## Dev: Run integration tests
	tools/test/run-integration-tests.sh '$(GO_TEST) -race -covermode=atomic -tags=integration -count=1 -coverpkg=./... -coverprofile=$(COVERAGE_INTEGRATION_PROFILE) $(PKG_LIST)'
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_INTEGRATION_PROFILE)" -o "$(COVERAGE_INTEGRATION_REPORT_HTML)"

