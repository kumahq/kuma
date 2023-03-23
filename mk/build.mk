build_info_fields = \
	version=$(BUILD_INFO_VERSION) \
	gitTag=$(GIT_TAG) \
	gitCommit=$(GIT_COMMIT)) \
	buildDate=$(BUILD_DATE) \
	Envoy=$(ENVOY_VERSION)

build_info_ld_flags := $(foreach entry,$(build_info_fields), -X github.com/kumahq/kuma/pkg/version.$(entry))

LD_FLAGS := -ldflags="-s -w $(build_info_ld_flags) $(EXTRA_LD_FLAGS)"
EXTRA_GOENV ?=
GOENV=CGO_ENABLED=0 $(EXTRA_GOENV)
GOFLAGS := -trimpath $(EXTRA_GOFLAGS)

TOP := $(shell pwd)
BUILD_DIR ?= $(TOP)/build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts-${GOOS}-${GOARCH}
BUILD_KUMACTL_DIR := ${BUILD_ARTIFACTS_DIR}/kumactl
export PATH := $(BUILD_KUMACTL_DIR):$(PATH)

# An optional extension to the coredns packages
COREDNS_EXT ?=
COREDNS_VERSION = v1.10.1

# List of binaries that we have build/release build rules for.
BUILD_RELEASE_BINARIES := kuma-cp kuma-dp kumactl coredns envoy kuma-cni install-cni
# List of binaries that we have build/test build roles for.
BUILD_TEST_BINARIES := test-server

# This is a list of all architecture supported, this means we'll define targets for all these architectures
SUPPORTED_GOARCHES ?= amd64 arm64
# This is a list of all os supported, this means we'll define targets for all these OSes
SUPPORTED_GOOSES ?= linux darwin

# This is a list of all architecture enabled, this means generic targets like `make build` or `make images` will build for each of these arches
ENABLED_GOARCHES ?= $(GOARCH)

.PHONY: build
build: build/release build/test ## Dev: Build all binaries

.PHONY: build/release
build/release: $(addprefix build/,$(BUILD_RELEASE_BINARIES)) ## Dev: Build release binaries

.PHONY: build/test
build/test: $(addprefix build/,$(BUILD_TEST_BINARIES)) ## Dev: Build testing binaries

# create targets like `make build/kumactl` that will build binaries for all arches defined in `$ENABLED_GOARCHES`
define LOCAL_BUILD_TARGET
build/$(2): build/artifacts-$(GOOS)-$(1)/$(2)
endef
$(foreach goarch,$(ENABLED_GOARCHES),$(foreach target,$(BUILD_RELEASE_BINARIES) $(BUILD_TEST_BINARIES),$(eval $(call LOCAL_BUILD_TARGET,$(goarch),$(target)))))

# Build_Go_Application is a build command for the Kuma Go applications.
Build_Go_Application = GOOS=$(1) GOARCH=$(2) $$(GOENV) go build -v $$(GOFLAGS) $$(LD_FLAGS) -o $$@/$$(notdir $$@)

# create targets to build binaries for each OS/ARCH combination
define BUILD_TARGET
.PHONY: build/artifacts-$(1)-$(2)/kuma-cp
build/artifacts-$(1)-$(2)/kuma-cp:
	$(Build_Go_Application) ./app/kuma-cp

.PHONY: build/artifacts-$(1)-$(2)/kuma-dp
build/artifacts-$(1)-$(2)/kuma-dp:
	$(Build_Go_Application) ./app/kuma-dp

.PHONY: build/artifacts-$(1)-$(2)/kumactl
build/artifacts-$(1)-$(2)/kumactl: build/ebpf
	$(Build_Go_Application) ./app/kumactl

.PHONY: build/artifacts-$(1)-$(2)/kuma-cni
build/artifacts-$(1)-$(2)/kuma-cni:
	$(Build_Go_Application) -ldflags="-extldflags=-static" ./app/cni/cmd/kuma-cni

.PHONY: build/artifacts-$(1)-$(2)/install-cni
build/artifacts-$(1)-$(2)/install-cni:
	$(Build_Go_Application) -ldflags="-extldflags=-static" ./app/cni/cmd/install

.PHONY: build/artifacts-$(1)-$(2)/coredns
build/artifacts-$(1)-$(2)/coredns:
	mkdir -p $$(@) && \
	[ -f $$(@)/coredns ] || \
	curl -s --fail --location https://github.com/kumahq/coredns-builds/releases/download/$(COREDNS_VERSION)/coredns_$(COREDNS_VERSION)_$(1)_$(2)$(COREDNS_EXT).tar.gz | tar -C $$(@) -xz

.PHONY: build/artifacts-$(1)-$(2)/test-server
build/artifacts-$(1)-$(2)/test-server:
	$(Build_Go_Application) ./test/server
endef
$(foreach goos,$(SUPPORTED_GOOSES),$(foreach goarch,$(SUPPORTED_GOARCHES),$(eval $(call BUILD_TARGET,$(goos),$(goarch)))))

.PHONY: clean
clean: clean/build ## Dev: Clean

.PHONY: clean/build
clean/build: clean/ebpf ## Dev: Remove build/ dir
	rm -rf "$(BUILD_DIR)"
