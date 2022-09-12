K8SCLUSTERS = kuma-1 kuma-2
K8SCLUSTERS_START_TARGETS = $(addprefix test/e2e/k8s/start/cluster/, $(K8SCLUSTERS))
K8SCLUSTERS_STOP_TARGETS  = $(addprefix test/e2e/k8s/stop/cluster/, $(K8SCLUSTERS))
K8SCLUSTERS_LOAD_IMAGES_TARGETS  = $(addprefix test/e2e/k8s/load/images/, $(K8SCLUSTERS))
K8SCLUSTERS_WAIT_TARGETS  = $(addprefix test/e2e/k8s/wait/, $(K8SCLUSTERS))
# export `IPV6=true` to enable IPv6 testing

# Targets to run prior to running the tests
E2E_DEPS_TARGETS ?=
# Environment veriables the tests should run with
E2E_ENV_VARS ?=

ifdef CI
# In circleCI all this was built from previous targets let's reuse them!
E2E_DEPS_TARGETS+= docker/load
else
E2E_DEPS_TARGETS+= build/kumactl images
endif

ifndef KUMA_UNIVERSAL_IMAGE
	KUMA_UNIVERSAL_IMAGE=$(KUMA_UNIVERSAL_DOCKER_IMAGE)
endif
E2E_ENV_VARS += KUMA_UNIVERSAL_IMAGE=$(KUMA_UNIVERSAL_IMAGE)

# We don't use `go list` here because Ginkgo requires disk path names,
# not Go packages names.
TEST_NAMES = $(shell ls -1 ./test/e2e)
ALL_TESTS = $(addprefix ./test/e2e/, $(addsuffix /..., $(TEST_NAMES)))
E2E_PKG_LIST ?= $(ALL_TESTS)
KUBE_E2E_PKG_LIST ?= ./test/e2e_env/kubernetes
UNIVERSAL_E2E_PKG_LIST ?= ./test/e2e_env/universal
MULTIZONE_E2E_PKG_LIST ?= ./test/e2e_env/multizone
GINKGO_E2E_TEST_FLAGS ?=
GINKGO_E2E_LABEL_FILTERS ?=
GINKGO_TEST_E2E=$(GINKGO_TEST) -v --slow-spec-threshold 30s $(GINKGO_E2E_TEST_FLAGS) --label-filter="$(GINKGO_E2E_LABEL_FILTERS)"

define append_label_filter
$(if $(GINKGO_E2E_LABEL_FILTERS),$(GINKGO_E2E_LABEL_FILTERS) && $(1),$(1))
endef

ifdef K3D
	K8S_CLUSTER_TOOL=k3d
	ifeq ($(K3D_NETWORK_CNI),calico)
		E2E_ENV_VARS += KUMA_K8S_TYPE=k3d-calico
	else
		E2E_ENV_VARS += KUMA_K8S_TYPE=k3d
	endif
else
	K8S_CLUSTER_TOOL=kind
	GINKGO_E2E_LABEL_FILTERS := $(call append_label_filter,!kind-not-supported)
endif

ifeq ($(CI_K3S_VERSION),v1.19.16-k3s1)
GINKGO_E2E_LABEL_FILTERS := $(call append_label_filter,!legacy-k3s-not-supported)
endif


ifdef IPV6
KIND_CONFIG_IPV6=-ipv6
endif

define gen-k8sclusters
.PHONY: test/e2e/k8s/start/cluster/$1
test/e2e/k8s/start/cluster/$1:
	KIND_CONFIG=$(TOP)/test/kind/cluster$(KIND_CONFIG_IPV6)-$1.yaml \
	KIND_CLUSTER_NAME=$1 \
		$(MAKE) $(K8S_CLUSTER_TOOL)/start

.PHONY: test/e2e/k8s/load/images/$1
test/e2e/k8s/load/images/$1:
	KIND_CLUSTER_NAME=$1 $(MAKE) $(K8S_CLUSTER_TOOL)/load/images

.PHONY: test/e2e/k8s/wait/$1
test/e2e/k8s/wait/$1:
	KIND_CLUSTER_NAME=$1 $(MAKE) $(K8S_CLUSTER_TOOL)/wait

.PHONY: test/e2e/k8s/stop/cluster/$1
test/e2e/k8s/stop/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
		$(MAKE) $(K8S_CLUSTER_TOOL)/stop
endef

$(foreach cluster, $(K8SCLUSTERS), $(eval $(call gen-k8sclusters,$(cluster))))

ifdef K8SCLUSTERS
E2E_ENV_VARS += K8SCLUSTERS="$(K8SCLUSTERS)"
endif
E2E_ENV_VARS += KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl
E2E_ENV_VARS += PATH=$(CI_TOOLS_BIN_DIR):$$PATH
.PHONY: test/e2e/list
test/e2e/list:
	@echo $(ALL_TESTS)

.PHONY: test/e2e/k8s/start
test/e2e/k8s/start: $(K8SCLUSTERS_START_TARGETS)
	$(MAKE) $(K8SCLUSTERS_LOAD_IMAGES_TARGETS) # execute after start targets

.PHONY: test/e2e/k8s/stop
test/e2e/k8s/stop: $(K8SCLUSTERS_STOP_TARGETS)

# test/e2e/debug is used for quicker feedback of E2E tests (ex. debugging flaky tests)
# It runs tests with fail fast which means you don't have to wait for all tests to get information that something failed
# Clusters are deleted only if all tests passes, otherwise clusters are live and running current test deployment
# GINKGO_EDITOR_INTEGRATION is required to work with focused test. Normally they exit with non 0 code which prevents clusters to be cleaned up.
# We run ginkgo instead of "go test" to fail fast (builtin "go test" fail fast does not seem to work with individual ginkgo tests)
.PHONY: test/e2e/debug
test/e2e/debug: build/kumactl images test/e2e/k8s/start
	$(E2E_ENV_VARS) \
	GINKGO_EDITOR_INTEGRATION=true \
		$(GINKGO_TEST_E2E) --keep-going=false --fail-fast $(E2E_PKG_LIST)
	$(MAKE) test/e2e/k8s/stop

# test/e2e/debug-fast is an experimental target tested with K3D=true.
# test/e2e/debug-fast is an equivalent of test/e2e/debug, but with the goal to minimize time for test to start running.
# Run only with -j and K3D=true
.PHONY: test/e2e/debug-fast
test/e2e/debug-fast:
	$(MAKE) $(K8SCLUSTERS_START_TARGETS) & # start K8S clusters in the background since it takes the most time
	$(MAKE) images
	$(MAKE) build/kumactl
	$(MAKE) $(K8SCLUSTERS_LOAD_IMAGES_TARGETS) # K3D is able to load images before the cluster is ready. It retries if cluster is not able to handle images yet.
	$(MAKE) $(K8SCLUSTERS_WAIT_TARGETS) # there is no easy way of waiting for processes in the background so just wait for K8S clusters
	$(E2E_ENV_VARS) \
	GINKGO_EDITOR_INTEGRATION=true \
		$(GINKGO_TEST_E2E) --procs 1 --keep-going=false --fail-fast $(E2E_PKG_LIST)
	$(MAKE) test/e2e/k8s/stop

# test/e2e/debug-universal is the same target as 'test/e2e/debug' but builds only 'kuma-universal' image
# and doesn't start Kind clusters
.PHONY: test/e2e/debug-universal
test/e2e/debug-universal: build/kumactl images/test
	$(E2E_ENV_VARS) \
	GINKGO_EDITOR_INTEGRATION=true \
		$(GINKGO_TEST_E2E) --keep-going=false --fail-fast $(E2E_PKG_LIST)


.PHONY: test/e2e
test/e2e: $(E2E_DEPS_TARGETS)
	$(MAKE) test/e2e/k8s/start
	$(E2E_ENV_VARS) $(GINKGO_TEST_E2E) --procs 1 $(E2E_PKG_LIST) || (ret=$$?; $(MAKE) test/e2e/k8s/stop && exit $$ret)
	$(MAKE) test/e2e/k8s/stop

.PHONY: test/e2e-kubernetes
test/e2e-kubernetes: $(E2E_DEPS_TARGETS)
	$(MAKE) test/e2e/k8s/start/cluster/kuma-1
	$(MAKE) test/e2e/k8s/wait/kuma-1
	$(MAKE) test/e2e/k8s/load/images/kuma-1
	$(E2E_ENV_VARS) $(GINKGO_TEST_E2E) $(KUBE_E2E_PKG_LIST) || (ret=$$?; $(MAKE) test/e2e/k8s/stop/cluster/kuma-1 && exit $$ret)
	$(MAKE) test/e2e/k8s/stop/cluster/kuma-1

.PHONY: test/e2e-universal
test/e2e-universal: build/kumactl images/test k3d/network/create
	$(E2E_ENV_VARS) $(GINKGO_TEST_E2E) $(UNIVERSAL_E2E_PKG_LIST)

.PHONY: test/e2e-multizone
test/e2e-multizone: $(E2E_DEPS_TARGETS)
	$(MAKE) test/e2e/k8s/start
	$(E2E_ENV_VARS) $(GINKGO_TEST_E2E) $(MULTIZONE_E2E_PKG_LIST) || (ret=$$?; $(MAKE) test/e2e/k8s/stop/cluster/kuma-1 && exit $$ret)
	$(MAKE) test/e2e/k8s/stop
