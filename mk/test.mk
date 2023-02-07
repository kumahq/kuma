UPDATE_GOLDEN_FILES ?=
TEST_PKG_LIST ?= ./...

REPORTS_DIR ?= ./build/reports
COVERAGE_PROFILE_FILENAME := coverage.out
COVERAGE_PROFILE := $(REPORTS_DIR)/$(COVERAGE_PROFILE_FILENAME)
COVERAGE_REPORT_HTML := $(REPORTS_DIR)/coverage.html

GINKGO_TEST_FLAGS += -p --keep-going \
	--keep-separate-reports \
	--junit-report results.xml

ifneq ($(REPORTS_DIR), "")
	GINKGO_TEST_FLAGS += --output-dir $(REPORTS_DIR)
endif

GINKGO_UNIT_TEST_FLAGS ?= \
	--skip-package ./test,./pkg/transparentproxy/istio/tools --race \
	--cover --covermode atomic --coverpkg ./... --coverprofile $(COVERAGE_PROFILE_FILENAME)

GINKGO_TEST:=$(GINKGO) $(GOFLAGS) $(LD_FLAGS) $(GINKGO_TEST_FLAGS)

.PHONY: test
test:
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test $(GOFLAGS) $(LD_FLAGS) -race $$(go list $(TEST_PKG_LIST) | grep -E -v "test/e2e" | grep -E -v "pkg/transparentproxy/istio/tools")

.PHONY: test-with-reports
test-with-reports: ${COVERAGE_PROFILE} ## Dev: Run tests for all modules
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) $(GINKGO_TEST) $(GINKGO_UNIT_TEST_FLAGS) $(TEST_PKG_LIST)
	$(MAKE) coverage

${COVERAGE_PROFILE}:
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"

.PHONY: coverage
coverage: ${COVERAGE_PROFILE}
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

.PHONY: test/kuma-cp
test/kuma-cp: TEST_PKG_LIST=./app/kuma-cp/... ./pkg/config/app/kuma-cp/...
test/kuma-cp: test ## Dev: Run `kuma-cp` tests only

.PHONY: test/kuma-dp
test/kuma-dp: TEST_PKG_LIST=./app/kuma-dp/... ./pkg/config/app/kuma-dp/...
test/kuma-dp: test ## Dev: Run `kuma-dp` tests only

.PHONY: test/kumactl
test/kumactl: TEST_PKG_LIST=./app/kumactl/... ./pkg/config/app/kumactl/...
test/kumactl: test ## Dev: Run `kumactl` tests only

.PHONY: test/cni
test/cni: TEST_PKG_LIST=./app/cni/...
test/cni: test ## Dev: Run `cni` tests only
