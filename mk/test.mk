UPDATE_GOLDEN_FILES ?=
TEST_PKG_LIST ?= ./...
REPORTS_DIR ?= build/reports

GINKGO_UNIT_TEST_FLAGS ?= \
	--skip-package ./test --race

# -race requires CGO_ENABLED=1 https://go.dev/doc/articles/race_detector and https://github.com/golang/go/issues/27089
UNIT_TEST_ENV=$(GOENV) CGO_ENABLED=1 KUBEBUILDER_ASSETS=$(KUBEBUILDER_ASSETS) TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) $(if $(CI),TESTCONTAINERS_RYUK_DISABLED=true,GINKGO_EDITOR_INTEGRATION=true)
GINKGO_TEST:=$(GINKGO) $(GOFLAGS) $(LD_FLAGS) -p --keep-going --keep-separate-reports --junit-report results.xml --output-dir $(REPORTS_DIR)

.PHONY: test
test: build/ebpf | $(REPORTS_DIR) ## Dev: Run tests for all modules. to include reports set `make TEST_REPORTS=1` and `make TEST_REPORTS=coverage` to include coverage. To run only some tests by set `TEST_PKG_LIST=./pkg/...` for example
ifdef TEST_REPORTS
	$(UNIT_TEST_ENV) $(GINKGO_TEST) $(GINKGO_UNIT_TEST_FLAGS) $(if $(findstring coverage,$(TEST_REPORTS)),--cover --covermode atomic --coverpkg ./... --coverprofile coverage.out) $(TEST_PKG_LIST)
	$(if $(findstring coverage,$(TEST_REPORTS)),GOFLAGS='${GOFLAGS}' go tool cover -html=$(REPORTS_DIR)/coverage.out -o "$(REPORTS_DIR)/coverage.html")
endif
ifndef TEST_REPORTS
	$(UNIT_TEST_ENV) go test $(GOFLAGS) $(LD_FLAGS) -race $$(go list $(TEST_PKG_LIST) | grep -E -v "test/e2e" | grep -E -v "test/blackbox_network_tests")
endif

$(REPORTS_DIR):
	mkdir -p $(REPORTS_DIR)

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
