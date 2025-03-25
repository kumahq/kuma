# this is a separate file to make sure it's imported *after* other makefiles
# since it relies on variables defined there we also keep it separate from something
# like e2e.new.mk to avoid confusion

.PHONY: test/skipped/report/%
test/skipped/report/%:
	@echo "Running Ginkgo in dry-run mode to generate '$(TEMP_DIR)/$*.json' report... this may take a few minutes"
	@$(ADDITIONAL_ENV) $(GINKGO) $(GOFLAGS) $(call LD_FLAGS,$(GOOS),$(GOARCH)) \
		--output-dir $(TEMP_DIR) \
		--json-report "$*.json" \
		--dry-run \
		--keep-going \
		--succinct \
		$(DIRECTORIES)

test/skipped/report/tproxy: DIRECTORIES := "./test/transparentproxy/..."
test/skipped/report/unit: DIRECTORIES := $(shell find . -type f -name "*suite_test.go*" ! -path "*gatewayapi*" ! -path "*e2e*" ! -path "*transparentproxy*" -exec dirname "{}" \; | uniq)
test/skipped/report/e2e: DIRECTORIES := $(E2E_PKG_LIST) $(MULTIZONE_E2E_PKG_LIST) $(UNIVERSAL_E2E_PKG_LIST) $(MULTIZONE_E2E_PKG_LIST)

.PHONY: test/skipped/report
test/skipped/report: test/skipped/report/unit test/skipped/report/e2e
	@$(MAKE) build/kumactl GOOS=linux
	@$(MAKE) test/skipped/report/tproxy \
		TEMP_DIR=$(TEMP_DIR) \
		ADDITIONAL_ENV="KUMACTL_LINUX_BIN=$(KUMACTL_LINUX_BIN)"

.PHONY: test/skipped/list/%
test/skipped/list/%:
	@go run $(KUMA_DIR)/tools/ci/list-disabled-tests/main.go --input-file "$(TEMP_DIR)/$*.json" $(ADDITIONAL_FLAGS)

.PHONY: test/skipped
test/skipped: TEMP_DIR := $(shell mktemp --directory)
test/skipped: ADDITIONAL_FLAGS += $(if $(NO_COLOR), --no-color,)
test/skipped: test/skipped/report test/skipped/list/unit test/skipped/list/e2e test/skipped/list/tproxy
