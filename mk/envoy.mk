ENVOY_TAG ?= v1.18.4
ENVOY_COMMIT_HASH ?=

SOURCE_DIR ?= ${TMPDIR}envoy-sources
ifndef TMPDIR
	SOURCE_DIR ?= /tmp/envoy-sources
endif

BUILD_ENVOY_FROM_SOURCES ?= false

.PHONY: build/envoy
build/envoy:
	$(MAKE) build/artifacts-${GOOS}-${GOARCH}/envoy

build/artifacts-%-amd64/envoy:
ifeq ($(BUILD_ENVOY_FROM_SOURCES),true)
	ENVOY_TAG=${ENVOY_TAG} \
	ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH} \
	SOURCE_DIR=${SOURCE_DIR} \
	BINARY_PATH=$@ ./tools/envoy/build_$*.sh
else
	ENVOY_TAG=${ENVOY_TAG} \
	ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH} \
	BINARY_PATH=$@ \
	ENVOY_DISTRO=$* ./tools/envoy/fetch.sh
endif
