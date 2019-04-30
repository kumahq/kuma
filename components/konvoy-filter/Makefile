.PHONY: help clean \
    bazel/build bazel/test bazel/coverage \
	build/binary build/binary.tar.gz require/binary.tar.gz upload/binary.tar.gz \
	run/tests collect/coverage require/coverage archive/coverage upload/coverage \
	dev/shell

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z./_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# Build arguments
TARGET_OS ?=

# CI coordinates
COMMIT ?= 0000000000000000000000000000000000000000
S3_CI_BUCKET ?= konvoy-ci
S3_KONVOY_BINARIES ?= $(COMMIT)/binaries
S3_KONVOY_COVERAGE ?= $(COMMIT)/coverage

# Build tools
DO_CI ?= ci/do_ci.sh

# Conditional variables
ifeq ($(TARGET_OS),)
	OS_NAME := $(shell uname -s)
	ifeq ($(OS_NAME),Darwin)
		TARGET_OS=macos
	else ifeq ($(OS_NAME),Linux)
		TARGET_OS=debian
	else
$(error Build in $(OS_NAME) is not supported yet)
	endif
endif
ifeq ($(TARGET_OS),debian)
	export IMAGE_NAME=envoyproxy/envoy-build-ubuntu
else ifeq ($(TARGET_OS),centos)
	export IMAGE_NAME=envoyproxy/envoy-build-centos
	# TODO(yskopets): `envoy-build-centos` images were not published until March 2019,
	#                 that is why it's not possible to use the same tag as for `envoy-build-ubuntu` yet
	export IMAGE_ID=latest
else ifeq ($(TARGET_OS),macos)
	DO_CI=ci/do_mac_ci.sh
else
$(error TARGET_OS must be one of [debian,centos,macos] but got "$(TARGET_OS)" instead)
endif

# Bazel output files
BAZEL_KONVOY_BINARY := $(PWD)/_build/build_release/konvoy
BAZEL_KONVOY_STRIPPED_BINARY := $(PWD)/_build/build_release_stripped/konvoy
BAZEL_COVERAGE_DIR := $(PWD)/generated/coverage

# Temporary files
BUILD_DIR ?= /tmp/build/konvoy/components/konvoy-filter
OUTPUT_DIR := $(PWD)/_build
STAGING_DIR ?= $(OUTPUT_DIR)/staging
ARTIFACTS_DIR ?= $(OUTPUT_DIR)/artifacts
BINARIES_DIR ?= $(ARTIFACTS_DIR)/binaries
COVERAGE_DIR ?= $(ARTIFACTS_DIR)/coverage
COVERAGE_TAR ?= $(COVERAGE_DIR)/coverage.tar.gz

# Commands
RUN_DO_CI = BUILD_DIR=$(BUILD_DIR) $(DO_CI)

clean: ## Remove output files
	rm -rf $(OUTPUT_DIR)

bazel/build:
	$(RUN_DO_CI) build

bazel/test:
	$(RUN_DO_CI) test

bazel/coverage:
	$(RUN_DO_CI) coverage

build/binary: bazel/build ## Build Konvoy binary
	@echo "Find Konvoy binaries at:"
	@echo "  $(BAZEL_KONVOY_BINARY)"
	@echo "  $(BAZEL_KONVOY_STRIPPED_BINARY)"

build/binary.tar.gz: bazel/build ## Build Konvoy binary (.tar.gz)
	@echo "Creating STAGING_DIR at '$(STAGING_DIR)' ..."
	mkdir -p $(STAGING_DIR)/bin
	cp -f $(BAZEL_KONVOY_STRIPPED_BINARY) $(STAGING_DIR)/bin/konvoy
	@echo "Creating BINARIES_DIR at '$(BINARIES_DIR)' ..."
	mkdir -p $(BINARIES_DIR)
	@echo "Creating .tar.gz in '$(BINARIES_DIR)' ..."
	tar -cvzf $(BINARIES_DIR)/konvoy-$(TARGET_OS).tar.gz -C $(STAGING_DIR) .

require/binary.tar.gz:
	[ $(shell find $(BINARIES_DIR) -maxdepth 1 -type f | wc -l) -eq 1 ]

upload/binary.tar.gz: require/binary.tar.gz ## Upload Konvoy binary (.tar.gz) into S3 bucket
	aws s3 sync $(BINARIES_DIR) s3://$(S3_CI_BUCKET)/$(S3_KONVOY_BINARIES)/

run/tests: bazel/test ## Run unit and integration tests

collect/coverage: bazel/coverage ## Collect test coverage

require/coverage:
	[ -d $(BAZEL_COVERAGE_DIR) ]

archive/coverage: require/coverage
	@echo "Creating COVERAGE_DIR at '$(COVERAGE_DIR)' ..."
	mkdir -p $(COVERAGE_DIR)
	tar -czf $(COVERAGE_TAR) -C $(BAZEL_COVERAGE_DIR) .	

upload/coverage: archive/coverage ## Upload test coverage report into S3 bucket
	aws s3 cp $(COVERAGE_TAR) s3://$(S3_CI_BUCKET)/$(S3_KONVOY_COVERAGE)/

run/demo: build/binary ## Run Konvoy with demo configuration
	$(OUTPUT_DIR)/build_release/konvoy -c configs/konvoy.yaml

dev/shell: ## Launch a shell inside `envoy-build` Docker image
	ci/run_envoy_docker.sh shell
