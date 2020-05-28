
CLUSTERS = 1
CLUSTERS_START_TARGETS  = $(addprefix test/kind/start/cluster/, $(CLUSTERS))
CLUSTERS_STOP_TARGETS  = $(addprefix test/kind/stop/cluster/, $(CLUSTERS))

define gen-clusters
.PHONY: test/kind/start/cluster/$1
test/kind/start/cluster/$1:
	KIND_CLUSTER_NAME=kuma-$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-kuma-$1-config \
		make kind/start
	KIND_CLUSTER_NAME=kuma-$1 \
		make kind/load

.PHONY: test/kind/stop/cluster/$1
test/kind/stop/cluster/$1:
	KIND_CLUSTER_NAME=kuma-$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-kuma-$1-config \
		make kind/stop
endef

$(foreach cluster, $(CLUSTERS), $(eval $(call gen-clusters,$(cluster))))

.PHONY: test/integration/kind/start
test/integration/kind/start: $(CLUSTERS_START_TARGETS)

.PHONY: test/integration/kind/stop
test/integration/kind/stop: $(CLUSTERS_STOP_TARGETS)

.PHONY: test/integration/test
test/integration/test:
	KUMACTL=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
		$(GO_TEST) -tags=k8s,integration -v -run "Integration" -timeout=10m ./...

.PHONY: test/integration
test/integration: vet ${COVERAGE_INTEGRATION_PROFILE} build/kumactl test/integration/kind/start
	make test/integration/test || true
	make test/integration/kind/stop
