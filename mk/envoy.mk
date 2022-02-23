KUMA_DIR ?= .

BUILD_ENVOY_FROM_SOURCES ?= false

ENVOY_TAG ?= v1.21.1 # commit hash or git tag
# Remember to update pkg/version/compatibility.go
ENVOY_VERSION = $(shell ${KUMA_DIR}/tools/envoy/version.sh ${ENVOY_TAG})

ifeq ($(GOOS),linux)
	ENVOY_DISTRO ?= alpine
endif
ENVOY_DISTRO ?= ${GOOS}

SOURCE_DIR ?= ${TMPDIR}envoy-sources
ifndef TMPDIR
	SOURCE_DIR ?= /tmp/envoy-sources
endif

# Target 'build/envoy' allows to put Envoy binary under the build/artifacts-$GOOS-$GOARCH/envoy directory.
# Depending on the flag BUILD_ENVOY_FROM_SOURCES this target either fetches Envoy from binary registry or
# builds from sources. It's possible to build binaries for darwin, linux and centos by specifying GOOS
# and ENVOY_DISTRO variables. Envoy version could be specified by ENVOY_TAG that accepts git tag or commit
# hash values.
.PHONY: build/envoy
build/envoy:
	GOOS=${GOOS} \
	GOARCH=${GOARCH} \
	ENVOY_DISTRO=${ENVOY_DISTRO} \
	ENVOY_VERSION=${ENVOY_VERSION} \
	$(MAKE) build/artifacts-${GOOS}-${GOARCH}/envoy/envoy-${ENVOY_VERSION}-${ENVOY_DISTRO}

.PHONY: build/artifacts-linux-amd64/envoy/envoy
build/artifacts-linux-amd64/envoy/envoy:
	GOOS=linux GOARCH=amd64 $(MAKE) build/envoy

build/artifacts-${GOOS}-${GOARCH}/envoy/envoy-${ENVOY_VERSION}-${ENVOY_DISTRO}:
ifeq ($(BUILD_ENVOY_FROM_SOURCES),true)
	ENVOY_TAG=${ENVOY_TAG} \
	SOURCE_DIR=${SOURCE_DIR} \
	KUMA_DIR=${KUMA_DIR} \
	BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS} \
	BINARY_PATH=$@ ${KUMA_DIR}/tools/envoy/build_${ENVOY_DISTRO}.sh
else
	ENVOY_VERSION=${ENVOY_VERSION} \
	ENVOY_DISTRO=${ENVOY_DISTRO} \
	BINARY_PATH=$@ ${KUMA_DIR}/tools/envoy/fetch.sh
endif

.PHONY: clean/envoy
clean/envoy:
	rm -rf ${SOURCE_DIR}
	rm -rf build/artifacts-${GOOS}-${GOARCH}/envoy/
