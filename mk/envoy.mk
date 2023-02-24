BUILD_ENVOY_FROM_SOURCES ?= false
ENVOY_TAG ?= v$(ENVOY_VERSION)
ENVOY_ARTIFACT_EXT ?=

ifeq ($(GOOS),linux)
	ENVOY_DISTRO ?= alpine
endif
ENVOY_DISTRO ?= $(GOOS)

ifeq ($(ENVOY_DISTRO),centos)
	BUILD_ENVOY_SCRIPT = $(KUMA_DIR)/tools/envoy/build_centos.sh
endif
BUILD_ENVOY_SCRIPT ?= $(KUMA_DIR)/tools/envoy/build_$(GOOS).sh

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

.PHONY: build/artifacts-linux-arm64/envoy/envoy
build/artifacts-linux-arm64/envoy/envoy:
	GOOS=linux GOARCH=arm64 $(MAKE) build/envoy

build/artifacts-${GOOS}-${GOARCH}/envoy/envoy-${ENVOY_VERSION}-${ENVOY_DISTRO}:
ifeq ($(BUILD_ENVOY_FROM_SOURCES),true)
	ENVOY_TAG=$(ENVOY_TAG) \
	SOURCE_DIR=${SOURCE_DIR} \
	KUMA_DIR=${KUMA_DIR} \
	BAZEL_BUILD_EXTRA_OPTIONS=${BAZEL_BUILD_EXTRA_OPTIONS} \
	BINARY_PATH=$@ $(BUILD_ENVOY_SCRIPT)
else
	ENVOY_TAG=$(ENVOY_TAG) \
	ENVOY_DISTRO=${ENVOY_DISTRO} \
	ENVOY_ARTIFACT_EXT=${ENVOY_ARTIFACT_EXT} \
	BINARY_PATH=$@ ${KUMA_DIR}/tools/envoy/fetch.sh
endif

build/envoy/$(GOOS)-$(GOARCH)/%/envoy:
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	ENVOY_TARGET=$* \
	ENVOY_TAG=$(ENVOY_TAG) \
	BINARY_PATH=$@ ${KUMA_DIR}/tools/envoy/fetch.sh

.PHONY: clean/envoy
clean/envoy:
	rm -rf ${SOURCE_DIR}
	rm -rf build/artifacts-${GOOS}-${GOARCH}/envoy/
