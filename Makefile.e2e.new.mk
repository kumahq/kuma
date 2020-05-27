
CLUSTERS = 1
CLUSTER_TARGETS  = $(addprefix test/kind/start/cluster/, $(CLUSTERS))

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

.PHONY: test/kind/integration
test/kind/integration: ${COVERAGE_INTEGRATION_PROFILE} images
	make $(CLUSTER_TARGETS)
	KUMACTL=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
		go test -tags=integration $(GOFLAGS) $(LD_FLAGS) -v -run 'Integration' ./... || true
	make kind/stop/all