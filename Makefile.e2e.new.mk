
CLUSTERS = kuma-1
CLUSTERS_START_TARGETS = $(addprefix test/e2e/kind/start/cluster/, $(CLUSTERS))
CLUSTERS_STOP_TARGETS  = $(addprefix test/e2e/kind/stop/cluster/, $(CLUSTERS))

define gen-clusters
.PHONY: test/e2e/kind/start/cluster/$1
test/e2e/kind/start/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		make kind/start
	KIND_CLUSTER_NAME=$1 \
		make kind/load

.PHONY: test/e2e/kind/stop/cluster/$1
test/e2e/kind/stop/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		make kind/stop
endef

$(foreach cluster, $(CLUSTERS), $(eval $(call gen-clusters,$(cluster))))

.PHONY: test/e2e/kind/start
test/e2e/kind/start: $(CLUSTERS_START_TARGETS)

.PHONY: test/e2e/kind/stop
test/e2e/kind/stop: $(CLUSTERS_STOP_TARGETS)

.PHONY: test/e2e/test
test/e2e/test:
	KUMACTL=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
		$(GO_TEST) -v -timeout=30m ./test/e2e/...

.PHONY: test/e2e
test/e2e: vet build/kumactl test/e2e/kind/start
	make test/e2e/test || \
	(ret=$$?; \
	make test/e2e/kind/stop && \
	exit $$ret)
	make test/e2e/kind/stop
