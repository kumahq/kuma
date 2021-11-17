ENVOY_TAG ?= v1.18.4
ENVOY_COMMIT_HASH ?=

SOURCE_DIR ?= ${TMPDIR}envoy-sources
ifndef TMPDIR
	SOURCE_DIR ?= /tmp/envoy-sources
endif

BUILD_ENVOY_FROM_SOURCES ?= false
KUMA_DIR ?= .
# Target 'build/envoy' allows to put Envoy binary under the build/artifacts-$GOOS-$GOARCH directory.
# Depending on the flag BUILD_ENVOY_FROM_SOURCES this target either fetches Envoy from binary registry or
# builds from sources. It's possible to build binaries for darwin, linux and centos7 by specifying GOOS variable.
# Envoy version could be specified either by ENVOY_TAG or ENVOY_COMMIT_HASH, the latter takes precedence.
.PHONY: build/envoy
build/envoy: ## Envoy: build or fetch envoy binaries
	$(MAKE) build/artifacts-${GOOS}-${GOARCH}/envoy/envoy

build/artifacts-%-amd64/envoy/envoy:
ifeq ($(BUILD_ENVOY_FROM_SOURCES),true)
	ENVOY_TAG=${ENVOY_TAG} \
	ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH} \
	SOURCE_DIR=${SOURCE_DIR} \
	KUMA_DIR=${KUMA_DIR} \
	BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS} \
	BINARY_PATH=$@ ${KUMA_DIR}/tools/envoy/build_$*.sh
else
	ENVOY_TAG=${ENVOY_TAG} \
	ENVOY_COMMIT_HASH=${ENVOY_COMMIT_HASH} \
	BINARY_PATH=$@ \
	ENVOY_DISTRO=$* ${KUMA_DIR}/tools/envoy/fetch.sh
endif
