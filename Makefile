.PHONY: clean \
	build/binary upload/binary \
	run/tests \
	collect/coverage upload/coverage

# Build arguments
TARGET_OS ?= debian

# CI coordinates
COMMIT ?= 00000000000000000000
S3_CI_BUCKET ?= konvoy-ci
S3_KONVOY_BINARIES ?= $(COMMIT)/binaries
S3_KONVOY_COVERAGE ?= $(COMMIT)/coverage

# Build tools
DO_CI ?= ci/do_ci.sh

# Conditional variables
ifeq ($(TARGET_OS),debian)
  export IMAGE_NAME=envoyproxy/envoy-build-ubuntu
endif
ifeq ($(TARGET_OS),centos)
  export IMAGE_NAME=envoyproxy/envoy-build-centos
  # TODO(yskopets): `envoy-build-centos` images were not published until March 2019,
  #                 that is why it's not possible to use the same tag as for `envoy-build-ubuntu` yet  
  export IMAGE_ID=latest
endif
ifeq ($(TARGET_OS),macos)
  DO_CI=ci/do_mac_ci.sh
endif

# Bazel files
BAZEL_STRIPPED_BINARY ?= $(PWD)/build_release_stripped/konvoy
BAZEL_COVERAGE_DIR ?= $(PWD)/generated/coverage

# Temporary files
BUILD_DIR ?= $(PWD)/_build
STAGING_DIR ?= $(BUILD_DIR)/staging
OUTPUT_DIR ?= $(BUILD_DIR)/output
BINARIES_DIR ?= $(OUTPUT_DIR)/binaries
COVERAGE_DIR ?= $(OUTPUT_DIR)/coverage
COVERAGE_TAR ?= $(COVERAGE_DIR)/coverage.tar.gz

bazel/build:
	$(DO_CI) build

bazel/test:
	$(DO_CI) test

bazel/coverage:
	$(DO_CI) coverage

build/binary: bazel/build
	@echo "Creating STAGING_DIR at '$(STAGING_DIR)' ..."
	mkdir -p $(STAGING_DIR)/bin
	cp $(BAZEL_STRIPPED_BINARY) $(STAGING_DIR)/bin/konvoy
	@echo "Creating BINARIES_DIR at '$(BINARIES_DIR)' ..."
	mkdir -p $(BINARIES_DIR)
	@echo "Creating .tar.gz in '$(BINARIES_DIR)' ..."
	tar -cvzf $(BINARIES_DIR)/konvoy-$(TARGET_OS).tar.gz -C $(STAGING_DIR) .

require/binary:
	[ $(shell find $(BINARIES_DIR) -maxdepth 1 -type f | wc -l) -eq 1 ]

upload/binary: require/binary
	aws s3 sync $(BINARIES_DIR) s3://$(S3_CI_BUCKET)/$(S3_KONVOY_BINARIES)/

run/tests: bazel/test

collect/coverage: bazel/coverage

require/coverage:
	[ -d $(BAZEL_COVERAGE_DIR) ]

archive/coverage: require/coverage
	@echo "Creating COVERAGE_DIR at '$(COVERAGE_DIR)' ..."
	mkdir -p $(COVERAGE_DIR)
	tar -czf $(COVERAGE_TAR) -C $(BAZEL_COVERAGE_DIR) .	

upload/coverage: archive/coverage
	aws s3 cp $(COVERAGE_TAR) s3://$(S3_CI_BUCKET)/$(S3_KONVOY_COVERAGE)/
