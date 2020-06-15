
K8SCLUSTERS = kuma-1 kuma-2
K8SCLUSTERS_START_TARGETS = $(addprefix test/e2e/kind/start/cluster/, $(K8SCLUSTERS))
K8SCLUSTERS_STOP_TARGETS  = $(addprefix test/e2e/kind/stop/cluster/, $(K8SCLUSTERS))

define gen-k8sclusters
.PHONY: test/e2e/kind/start/cluster/$1
test/e2e/kind/start/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		make kind/start
	KIND_CLUSTER_NAME=$1 \
		make kind/load/images

.PHONY: test/e2e/kind/stop/cluster/$1
test/e2e/kind/stop/cluster/$1:
	KIND_CLUSTER_NAME=$1 \
	KIND_KUBECONFIG=$(KIND_KUBECONFIG_DIR)/kind-$1-config \
		make kind/stop

.PHONE: kind/load/images/$1
kind/load/images/$1:
	KIND_CLUSTER_NAME=$1 make kind/load/images
endef

$(foreach cluster, $(K8SCLUSTERS), $(eval $(call gen-k8sclusters,$(cluster))))

.PHONY: test/e2e/kind/start
test/e2e/kind/start: $(K8SCLUSTERS_START_TARGETS)

.PHONY: test/e2e/kind/stop
test/e2e/kind/stop: $(K8SCLUSTERS_STOP_TARGETS)

.PHONY: test/e2e/test
test/e2e/test:
	K8SCLUSTERS="$(K8SCLUSTERS)" \
	KUMACTLBIN=${BUILD_ARTIFACTS_DIR}/kumactl/kumactl \
		$(GO_TEST) -v -timeout=30m ./test/e2e/...

.PHONY: test/e2e
test/e2e: build/kumactl images test/e2e/kind/start
	make test/e2e/test || \
	(ret=$$?; \
	make test/e2e/kind/stop && \
	exit $$ret)
	make test/e2e/kind/stop
