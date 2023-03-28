UPDATE_GOLDEN_FILES ?=
TEST_PKG_LIST ?= ./...
REPORTS_DIR ?= build/reports

GINKGO_TEST_FLAGS += -p --keep-going \
	--keep-separate-reports \
	--junit-report results.xml \
	--output-dir $(REPORTS_DIR)

GINKGO_UNIT_TEST_FLAGS ?= \
	--skip-package ./test,./pkg/transparentproxy/istio/tools --race

TEST_ENV=KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) $(GOENV)
GINKGO_TEST:=$(TEST_ENV) $(GINKGO) $(GOFLAGS) $(LD_FLAGS) $(GINKGO_TEST_FLAGS)

.PHONY: test
test: build/ebpf ## Dev: Run tests for all modules
	# -race required CGO_ENABLED=1 https://go.dev/doc/articles/race_detector and https://github.com/golang/go/issues/27089
	$(TEST_ENV) CGO_ENABLED=1 go test $(GOFLAGS) $(LD_FLAGS) -race $$(go list $(TEST_PKG_LIST) | grep -E -v "test/e2e" | grep -E -v "test/blackbox_network_tests" | grep -E -v "pkg/transparentproxy/istio/tools")

.PHONY: test-with-reports
test-with-reports: build/ebpf ## Dev: Run tests with test reports
	$(GINKGO_TEST) $(GINKGO_UNIT_TEST_FLAGS) $(TEST_PKG_LIST)

.PHONY: test-with-coverage
test-with-coverage: build/ebpf
	mkdir -p $(REPORTS_DIR)
	$(GINKGO_TEST) $(GINKGO_UNIT_TEST_FLAGS) --cover --covermode atomic --coverpkg ./... --coverprofile coverage.out $(TEST_PKG_LIST)
	GOFLAGS='${GOFLAGS}' go tool cover -html=$(REPORTS_DIR)/coverage.out -o "$(REPORTS_DIR)/coverage.html"

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
