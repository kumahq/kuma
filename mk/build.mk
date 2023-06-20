build_info := $(shell $(TOOLS_DIR)/releases/version.sh)
BUILD_INFO_VERSION ?= $(word 1, $(build_info))

build_info_fields := \
	version=$(BUILD_INFO_VERSION) \
	gitTag=$(word 2, $(build_info)) \
	gitCommit=$(word 3, $(build_info)) \
	buildDate=$(word 4, $(build_info)) \
	Envoy=$(word 5, $(build_info))
build_info_ld_flags := $(foreach entry,$(build_info_fields), -X github.com/kumahq/kuma/pkg/version.$(entry))

LD_FLAGS := -ldflags="-s -w $(build_info_ld_flags) $(EXTRA_LD_FLAGS)"
GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
CGO_ENABLED := 0
GOFLAGS :=

ENVOY_DISTRO ?= $(GOOS)

TOP := $(shell pwd)
BUILD_DIR ?= $(TOP)/build
BUILD_ARTIFACTS_DIR ?= $(BUILD_DIR)/artifacts-${GOOS}-${GOARCH}
BUILD_KUMACTL_DIR := ${BUILD_ARTIFACTS_DIR}/kumactl
export PATH := $(BUILD_KUMACTL_DIR):$(PATH)

GO_BUILD := GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -v $(GOFLAGS) $(LD_FLAGS)
GO_BUILD_COREDNS := GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=${CGO_ENABLED} go build -v

COREDNS_GIT_REPOSITORY ?= https://github.com/coredns/coredns.git
COREDNS_VERSION ?= v1.10.1
COREDNS_TMP_DIRECTORY ?= $(BUILD_DIR)/coredns
COREDNS_PLUGIN_CFG_PATH ?= $(TOP)/tools/builds/coredns/templates/plugin.cfg

# List of binaries that we have release build rules for.
BUILD_RELEASE_BINARIES := kuma-cp kuma-dp kumactl kuma-prometheus-sd coredns envoy kuma-cni install-cni

# List of binaries that we have test build roles for.
BUILD_TEST_BINARIES := test-server

# Build_Go_Application is a build command for the Kuma Go applications.
Build_Go_Application = $(GO_BUILD) -o $(BUILD_ARTIFACTS_DIR)/$(notdir $@)/$(notdir $@)

.PHONY: build
build: build/release build/test

.PHONY: build/release
build/release: $(patsubst %,build/%,$(BUILD_RELEASE_BINARIES)) ## Dev: Build all binaries

.PHONY: build/test
build/test: $(patsubst %,build/%,$(BUILD_TEST_BINARIES)) ## Dev: Build testing binaries

.PHONY: build/linux-amd64
build/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: build/linux-arm64
build/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build

.PHONY: build/release/linux-amd64
build/release/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/release

.PHONY: build/release/linux-arm64
build/release/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/release

.PHONY: build/test/linux-amd64
build/test/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/test

.PHONY: build/test/linux-arm64
build/test/linux-arm64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/test

.PHONY: build/kuma-cp
build/kuma-cp: ## Dev: Build `Control Plane` binary
	$(Build_Go_Application) ./app/$(notdir $@)

.PHONY: build/kuma-dp
build/kuma-dp: ## Dev: Build `kuma-dp` binary
	$(Build_Go_Application) ./app/$(notdir $@)

.PHONY: build/kumactl
build/kumactl: ## Dev: Build `kumactl` binary
	$(Build_Go_Application) ./app/$(notdir $@)

.PHONY: build/kuma-cni
build/kuma-cni: ## Dev: Build `kuma-cni` binary
	$(Build_Go_Application) -ldflags="-extldflags=-static" ./app/cni/cmd/kuma-cni

.PHONY: build/install-cni
build/install-cni: ## Dev: Build `install-cni` binary
	$(Build_Go_Application) -ldflags="-extldflags=-static" ./app/cni/cmd/install

.PHONY: build/coredns
build/coredns:
ifeq (,$(wildcard $(BUILD_ARTIFACTS_DIR)/coredns/coredns))
	rm -rf "$(COREDNS_TMP_DIRECTORY)"
	git clone --branch $(COREDNS_VERSION) --depth 1 $(COREDNS_GIT_REPOSITORY) $(COREDNS_TMP_DIRECTORY)
	cp $(COREDNS_PLUGIN_CFG_PATH) $(COREDNS_TMP_DIRECTORY)
	cd $(COREDNS_TMP_DIRECTORY) && \
		GOOS= GOARCH= go generate coredns.go && \
		go get github.com/coredns/alternate && \
		$(GO_BUILD_COREDNS) -ldflags="-s -w -X github.com/coredns/coredns/coremain.GitCommit=$(shell git describe --dirty --always)" -o $(BUILD_ARTIFACTS_DIR)/coredns/coredns
	rm -rf "$(COREDNS_TMP_DIRECTORY)"
else
	echo "CoreDNS is already built. If you want to rebuild it, remove the binary: rm $(BUILD_ARTIFACTS_DIR)/coredns/coredns"
endif

.PHONY: build/envoy
build/envoy:
	mkdir -p $(BUILD_ARTIFACTS_DIR)/envoy && \
	if [ ! -f $(BUILD_ARTIFACTS_DIR)/envoy/envoy-$(ENVOY_DISTRO) ]; then \
		curl -s --fail --location https://github.com/kumahq/envoy-builds/releases/download/v$(ENVOY_VERSION)/envoy-$(ENVOY_DISTRO)-$(GOARCH)-v$(ENVOY_VERSION)$(ENVOY_EXT_$(GOOS)_$(GOARCH)).tar.gz | tar -C $(BUILD_ARTIFACTS_DIR)/envoy -xz; \
		mv $(BUILD_ARTIFACTS_DIR)/envoy/envoy $(BUILD_ARTIFACTS_DIR)/envoy/envoy-$(ENVOY_DISTRO); \
	fi

.PHONY: build/kuma-prometheus-sd
build/kuma-prometheus-sd: ## Dev: Build `kuma-prometheus-sd` binary
	$(Build_Go_Application) ./app/$(notdir $@)

.PHONY: build/test-server
build/test-server: ## Dev: Build `test-server` binary
	$(Build_Go_Application) ./test/server

.PHONY: build/kuma-cni/linux-amd64
build/kuma-cni/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-cni

.PHONY: build/kuma-cni/linux-arm64
build/kuma-cni/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/kuma-cni

.PHONY: build/install-cni/linux-amd64
build/install-cni/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/install-cni

.PHONY: build/install-cni/linux-arm64
build/install-cni/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/install-cni

.PHONY: build/kuma-cp/linux-amd64
build/kuma-cp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-cp

.PHONY: build/kuma-cp/linux-arm64
build/kuma-cp/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/kuma-cp

.PHONY: build/kuma-dp/linux-amd64
build/kuma-dp/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-dp

.PHONY: build/kuma-dp/linux-arm64
build/kuma-dp/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/kuma-dp

.PHONY: build/kumactl/linux-amd64
build/kumactl/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kumactl

.PHONY: build/kumactl/linux-arm64
build/kumactl/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/kumactl

.PHONY: build/kuma-prometheus-sd/linux-amd64
build/kuma-prometheus-sd/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/kuma-prometheus-sd

.PHONY: build/kuma-prometheus-sd/linux-arm64
build/kuma-prometheus-sd/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/kuma-prometheus-sd

.PHONY: build/coredns/linux-amd64
build/coredns/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/coredns

.PHONY: build/coredns/linux-arm64
build/coredns/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/coredns

.PHONY: build/test-server/linux-amd64
build/test-server/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/test-server

.PHONY: build/test-server/linux-arm64
build/test-server/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/test-server

.PHONY: build/envoy/linux-amd64
build/envoy/linux-amd64:
	GOOS=linux GOARCH=amd64 $(MAKE) build/envoy

.PHONY: build/envoy/linux-arm64
build/envoy/linux-arm64:
	GOOS=linux GOARCH=arm64 $(MAKE) build/envoy

.PHONY: clean
clean: clean/build ## Dev: Clean

.PHONY: clean/build
clean/build: ## Dev: Remove build/ dir
	rm -rf "$(BUILD_DIR)"
