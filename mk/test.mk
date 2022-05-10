UPDATE_GOLDEN_FILES ?=
GO_TEST := TMPDIR=/tmp UPDATE_GOLDEN_FILES=$(UPDATE_GOLDEN_FILES) go test $(GOFLAGS) $(LD_FLAGS)
GO_TEST_OPTS ?=
PKG_LIST ?= ./...
VENDORED_PKGS = pkg/transparentproxy/istio/tools
E2E_PKGS = test/e2e

BUILD_COVERAGE_DIR ?= $(BUILD_DIR)/coverage

COVERAGE_PROFILE := $(BUILD_COVERAGE_DIR)/coverage.out
COVERAGE_REPORT_HTML := $(BUILD_COVERAGE_DIR)/coverage.html

# This environment variable sets where the kubebuilder envtest framework looks
# for etcd and other tools that is consumes. The `dev/install/kubebuilder` make
# target guaranteed to link these tools into $CI_TOOLS_DIR.
export KUBEBUILDER_ASSETS=$(CI_TOOLS_DIR)

.PHONY: test
test: ${COVERAGE_PROFILE} ## Dev: Run tests for all modules
	$(GO_TEST) $(GO_TEST_OPTS) -race -covermode=atomic -coverpkg=./... -coverprofile="$(COVERAGE_PROFILE)" $$(go list $(PKG_LIST) | grep -E -v "$(E2E_PKGS)" | grep -E -v "$(VENDORED_PKGS)")
	$(MAKE) coverage

${COVERAGE_PROFILE}:
	mkdir -p "$(shell dirname "$(COVERAGE_PROFILE)")"

.PHONY: coverage
coverage: ${COVERAGE_PROFILE}
	GOFLAGS='${GOFLAGS}' go tool cover -html="$(COVERAGE_PROFILE)" -o "$(COVERAGE_REPORT_HTML)"

.PHONY: test/kuma-cp
test/kuma-cp: PKG_LIST=./app/kuma-cp/... ./pkg/config/app/kuma-cp/...
test/kuma-cp: test ## Dev: Run `kuma-cp` tests only

.PHONY: test/kuma-dp
test/kuma-dp: PKG_LIST=./app/kuma-dp/... ./pkg/config/app/kuma-dp/...
test/kuma-dp: test ## Dev: Run `kuma-dp` tests only

.PHONY: test/kumactl
test/kumactl: PKG_LIST=./app/kumactl/... ./pkg/config/app/kumactl/...
test/kumactl: test ## Dev: Run `kumactl` tests only
