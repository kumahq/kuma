K8SCLUSTERS = kuma-1 kuma-2
K8SCLUSTERS_START_TARGETS = $(addprefix test/e2e/kind/start/cluster/, $(K8SCLUSTERS))
K8SCLUSTERS_STOP_TARGETS  = $(addprefix test/e2e/kind/stop/cluster/, $(K8SCLUSTERS))
API_VERSION ?= v3

HELM_CHART_PATH ?=
KUMA_GLOBAL_IMAGE_TAG ?=
KUMA_GLOBAL_IMAGE_REGISTRY ?=
KUMA_CP_IMAGE_REPOSITORY ?=
KUMA_DP_IMAGE_REPOSITORY ?=
KUMA_DP_INIT_IMAGE_REPOSITORY ?=
KUMA_USE_LOAD_BALANCER ?=
KUMA_IN_EKS ?=
KUMA_UNIVERSAL_IMAGE ?= $(KUMA_UNIVERSAL_DOCKER_IMAGE)

TEST_NAMES = $(shell ls -1 ./test/e2e)
ALL_TESTS = $(addprefix ./test/e2e/, $(addsuffix /..., $(TEST_NAMES)))
E2E_PKG_LIST ?= $(ALL_TESTS)

define gen-k8sclusters
.PHONY: test/e2e/kind/start/cluster/$1
test/e2e/kind/start/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		$(MAKE) kind/start
	KIND_CLUSTER_NAME=$1 \
		$(MAKE) kind/load/images
	@kind load docker-image $(KUMA_UNIVERSAL_DOCKER_IMAGE) --name=$1

.PHONY: test/e2e/kind/stop/cluster/$1
test/e2e/kind/stop/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		$(MAKE) kind/stop

.PHONE: kind/load/images/$1
kind/load/images/$1:
	KIND_CLUSTER_NAME=$1 $(MAKE) kind/load/images
endef

$(foreach cluster, $(K8SCLUSTERS), $(eval $(call gen-k8sclusters,$(cluster))))

.PHHONY: test/e2e/list
test/e2e/list:
	@echo $(ALL_TESTS)

.PHONY: test/e2e/kind/start
test/e2e/kind/start: $(K8SCLUSTERS_START_TARGETS)

.PHONY: test/e2e/kind/stop
test/e2e/kind/stop: $(K8SCLUSTERS_STOP_TARGETS)

.PHONY: test/e2e/test
test/e2e/test:
	for t in $(E2E_PKG_LIST); do \
		K8SCLUSTERS="$(K8SCLUSTERS)" \
		KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
		API_VERSION="$(API_VERSION)" \
		HELM_CHART_PATH="$(HELM_CHART_PATH)" \
		KUMA_GLOBAL_IMAGE_TAG="$(KUMA_GLOBAL_IMAGE_TAG)" \
		KUMA_GLOBAL_IMAGE_REGISTRY="$(KUMA_GLOBAL_IMAGE_REGISTRY)" \
		KUMA_CP_IMAGE_REPOSITORY="$(KUMA_CP_IMAGE_REPOSITORY)" \
		KUMA_DP_IMAGE_REPOSITORY="$(KUMA_DP_IMAGE_REPOSITORY)" \
		KUMA_DP_INIT_IMAGE_REPOSITORY="$(KUMA_DP_INIT_IMAGE_REPOSITORY)" \
		KUMA_USE_LOAD_BALANCER="$(KUMA_USE_LOAD_BALANCER)" \
		KUMA_IN_EKS="$(KUMA_IN_EKS)" \
		KUMA_UNIVERSAL_IMAGE="$(KUMA_UNIVERSAL_IMAGE)" \
			$(GO_TEST) -v -timeout=45m $$t; \
	done

# test/e2e/debug is used for quicker feedback of E2E tests (ex. debugging flaky tests)
# It runs tests with fail fast which means you don't have to wait for all tests to get information that something failed
# Clusters are deleted only if all tests passes, otherwise clusters are live and running current test deployment
# GINKGO_EDITOR_INTEGRATION is required to work with focused test. Normally they exit with non 0 code which prevents clusters to be cleaned up.
# We run ginkgo instead of "go test" to fail fast (builtin "go test" fail fast does not seem to work with individual ginkgo tests)
.PHONY: test/e2e/debug
test/e2e/debug: build/kumactl images docker/build/kuma-universal test/e2e/kind/start
	K8SCLUSTERS="$(K8SCLUSTERS)" \
	KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
	API_VERSION="$(API_VERSION)" \
	GINKGO_EDITOR_INTEGRATION=true \
		ginkgo --failFast $(GOFLAGS) $(LD_FLAGS) $(E2E_PKG_LIST)
	$(MAKE) test/e2e/kind/stop

.PHONY: test/e2e
test/e2e: build/kumactl images docker/build/kuma-universal test/e2e/kind/start
	$(MAKE) test/e2e/test || \
	(ret=$$?; \
	$(MAKE) test/e2e/kind/stop && \
	exit $$ret)
	$(MAKE) test/e2e/kind/stop
