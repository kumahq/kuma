K8SCLUSTERS = kuma-1 kuma-2
K8SCLUSTERS_START_TARGETS = $(addprefix test/e2e/kind/start/cluster/, $(K8SCLUSTERS))
K8SCLUSTERS_STOP_TARGETS  = $(addprefix test/e2e/kind/stop/cluster/, $(K8SCLUSTERS))
API_VERSION ?= v2

KUMA_UNIVERSAL_DOCKER_IMAGE ?= kuma-universal
KUMA_UNIVERSAL_DOCKERFILE ?= test/dockerfiles/Dockerfile.universal

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

.PHONY: docker/build/universal
docker/build/universal: build/artifacts-linux-amd64/kuma-cp/kuma-cp build/artifacts-linux-amd64/kuma-dp/kuma-dp build/artifacts-linux-amd64/kumactl/kumactl
	DOCKER_BUILDKIT=1 \
	docker build -t $(KUMA_UNIVERSAL_DOCKER_IMAGE) -f $(KUMA_UNIVERSAL_DOCKERFILE) .

.PHONY: test/e2e/kind/start
test/e2e/kind/start: $(K8SCLUSTERS_START_TARGETS)

.PHONY: test/e2e/kind/stop
test/e2e/kind/stop: $(K8SCLUSTERS_STOP_TARGETS)

.PHONY: test/e2e/test
test/e2e/test: PKG_LIST=./test/e2e/...
test/e2e/test:
	K8SCLUSTERS="$(K8SCLUSTERS)" \
	KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
	API_VERSION="$(API_VERSION)" \
		$(GO_TEST) -v -timeout=45m $(PKG_LIST)

# test/e2e/debug is used for quicker feedback of E2E tests (ex. debugging flaky tests)
# It runs tests with fail fast which means you don't have to wait for all tests to get information that something failed
# Clusters are deleted only if all tests passes, otherwise clusters are live and running current test deployment
# GINKGO_EDITOR_INTEGRATION is required to work with focused test. Normally they exit with non 0 code which prevents clusters to be cleaned up.
# We run ginkgo instead of "go test" to fail fast (builtin "go test" fail fast does not seem to work with individual ginkgo tests)
.PHONY: test/e2e/debug
test/e2e/debug: PKG_LIST=./test/e2e/...
test/e2e/debug: build/kumactl images docker/build/universal test/e2e/kind/start
	K8SCLUSTERS="$(K8SCLUSTERS)" \
	KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
	API_VERSION="$(API_VERSION)" \
	GINKGO_EDITOR_INTEGRATION=true \
		ginkgo --failFast $(GOFLAGS) $(LD_FLAGS) $(PKG_LIST)
	$(MAKE) test/e2e/kind/stop

.PHONY: test/e2e
test/e2e: build/kumactl images docker/build/universal test/e2e/kind/start
	$(MAKE) test/e2e/test || \
	(ret=$$?; \
	$(MAKE) test/e2e/kind/stop && \
	exit $$ret)
	$(MAKE) test/e2e/kind/stop
