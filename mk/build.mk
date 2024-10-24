define LD_FLAGS
-ldflags="-s -w \
-X github.com/kumahq/kuma/pkg/version.version=$(BUILD_INFO_VERSION) \
-X github.com/kumahq/kuma/pkg/version.gitTag=$(GIT_TAG) \
-X github.com/kumahq/kuma/pkg/version.gitCommit=$(GIT_COMMIT) \
-X github.com/kumahq/kuma/pkg/version.buildDate=$(BUILD_DATE) \
-X github.com/kumahq/kuma/pkg/version.Envoy=$(if $(ENVOY_VERSION_$(1)_$(2)),$(ENVOY_VERSION_$(1)_$(2)),$(ENVOY_VERSION)) \
$(EXTRA_LD_FLAGS)"
endef

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
COREDNS_VERSION = v1.11.3

# List of binaries that we have build/release build rules for.
BUILD_RELEASE_BINARIES := kuma-cp kuma-dp kumactl coredns kuma-cni install-cni envoy
# List of binaries that we have build/test build roles for.
BUILD_TEST_BINARIES += test-server

# This is a list of all architecture supported, this means we'll define targets for all these architectures
SUPPORTED_GOARCHES ?= amd64 arm64
# This is a list of all os supported, this means we'll define targets for all these OSes
SUPPORTED_GOOSES ?= linux darwin

# This is a list of all architecture enabled, this means generic targets like `make build` or `make images` will build for each of these arches
ENABLED_GOARCHES ?= $(GOARCH)
# This is a list of all osses enabled, this means generic targets like `make build/distributions` will build for each of these arches
ENABLED_GOOSES ?= $(GOOS)

ifeq ($(FULL_MATRIX), true)
ENABLED_GOARCHES = $(SUPPORTED_GOARCHES)
ENABLED_GOOSES = $(SUPPORTED_GOOSES)
endif
# We can remove some specific combination that may be invalid with this
IGNORED_ARCH_OS ?=
ENABLED_ARCH_OS = $(filter-out $(IGNORED_ARCH_OS), $(foreach os,$(ENABLED_GOOSES),$(foreach arch,$(ENABLED_GOARCHES),$(os)-$(arch))))

.PHONY: build/info
build/info: ## Dev: Show build info
	@echo version=$(BUILD_INFO_VERSION)
	@echo gitTag=$(GIT_TAG)
	@echo gitCommit=$(GIT_COMMIT)
	@echo buildDate=$(BUILD_DATE)
	@echo Envoy=$(ENVOY_VERSION_$(GOOS)_$(GOARCH))
	@echo tools-dir: $(CI_TOOLS_DIR)
	@echo arch: supported=$(SUPPORTED_GOARCHES), enabled=$(ENABLED_GOARCHES)
	@echo os: supported=$(SUPPORTED_GOOSES), enabled=$(ENABLED_GOOSES)
	@echo arch-os ignored=$(IGNORED_ARCH_OS), enabled=$(ENABLED_ARCH_OS)
	$(EXTRA_BUILD_INFO)

.PHONY: build/info/short
build/info/short:
	@echo enabled arch-os:$(ENABLED_ARCH_OS)
	$(EXTRA_BUILD_INFO)

.PHONY: build/info/version
build/info/version:
	@echo $(BUILD_INFO_VERSION)

.PHONY: build
build: build/release build/test ## Dev: Build all binaries

.PHONY: build/release
build/release: $(addprefix build/,$(BUILD_RELEASE_BINARIES)) ## Dev: Build release binaries

.PHONY: build/test
build/test: $(addprefix build/,$(BUILD_TEST_BINARIES)) ## Dev: Build testing binaries

# create targets like `make build/kumactl` that will build binaries for all arches defined in `$ENABLED_GOARCHES` and `$ENABLED_GOOSES`
# $(1) - GOOS to build for
define LOCAL_BUILD_TARGET
build/$(1): $$(patsubst %,build/artifacts-%/$(1),$$(ENABLED_ARCH_OS))
endef
$(foreach target,$(BUILD_RELEASE_BINARIES) $(BUILD_TEST_BINARIES),$(eval $(call LOCAL_BUILD_TARGET,$(target))))

# Build_Go_Application is a build command for the Kuma Go applications.
Build_Go_Application = GOOS=$(1) GOARCH=$(2) $$(GOENV) go build -v $$(GOFLAGS) $(call LD_FLAGS,$(1),$(2)) -o $$@/$$(notdir $$@)

# create targets to build binaries for each OS/ARCH combination
# $(1) - GOOS to build for
# $(2) - GOARCH to build for
define BUILD_TARGET
ENVOY_VERSION_$(1)_$(2)=$(if $(ENVOY_VERSION_$(1)_$(2)),$(ENVOY_VERSION_$(1)_$(2)),$(ENVOY_VERSION))
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

.PHONY: build/artifacts-$(1)-$(2)/envoy
build/artifacts-$(1)-$(2)/envoy:
	mkdir -p $$(@) && \
	[ -f $$(@)/envoy ] || \
	curl -s --fail --location https://github.com/kumahq/envoy-builds/releases/download/v$$(ENVOY_VERSION_$(1)_$(2))/envoy-$(1)-$(2)-v$$(ENVOY_VERSION_$(1)_$(2))$(ENVOY_EXT_$(1)_$(2)).tar.gz | tar -C $$(@) -xz

.PHONY: build/artifacts-$(1)-$(2)/test-server
build/artifacts-$(1)-$(2)/test-server:
	$(Build_Go_Application) ./test/server

endef
$(foreach goos,$(SUPPORTED_GOOSES),$(foreach goarch,$(SUPPORTED_GOARCHES),$(eval $(call BUILD_TARGET,$(goos),$(goarch)))))

.PHONY: clean/build
clean/build: clean/ebpf ## Dev: Remove build/ dir
	rm -rf "$(BUILD_DIR)"
