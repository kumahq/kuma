# Auto-include dependencies (guarded, safe to include multiple times)
include $(dir $(lastword $(MAKEFILE_LIST)))common.mk
include $(dir $(lastword $(MAKEFILE_LIST)))k8s.mk

ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-registry=$(DOCKER_REGISTRY) --dataplane-registry=$(DOCKER_REGISTRY) --dataplane-init-registry=$(DOCKER_REGISTRY)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif

# renovate[docker]: depName=kindest/node
CI_KUBERNETES_VERSION ?= v1.35.1@sha256:05d7bcdefbda08b4e038f644c4df690cdac3fba8b06f8289f30e10026720a1ab

KUMA_MODE ?= zone
KUMA_NAMESPACE ?= kuma-system

DOCKERHUB_PULL_CREDENTIAL ?=
.PHONY: kind/setup-docker-credentials
kind/setup-docker-credentials:
	@mkdir -p /tmp/.kuma-dev ; \
	echo '{"auths":{}}' > /tmp/.kuma-dev/kind-config.json ; \
	if [[ "$(DOCKERHUB_PULL_CREDENTIAL)" != "" ]]; then \
		echo "{\"auths\":{\"https://index.docker.io/v1/\":{\"auth\":\"$$(echo -n "$(DOCKERHUB_PULL_CREDENTIAL)" | base64)\"}}}" > /tmp/.kuma-dev/kind-config.json ; \
	fi

.PHONY: kind/cleanup-docker-credentials
kind/cleanup-docker-credentials:
	@rm -f /tmp/.kuma-dev/kind-config.json

.PHONY: kind/cluster/start
kind/cluster/start: $(KUBECONFIG_DIR) kind/setup-docker-credentials
	$(KIND) get clusters | grep -x "$(CLUSTER_NAME)" >/dev/null 2>&1 && echo "Kind cluster already running." || \
		($(KIND) create cluster \
			--name "$(CLUSTER_NAME)" \
			--config "$(KUMA_DIR)/test/kind/cluster-$(if $(IPV6),ipv6-,)$(CLUSTER_NAME).yaml" \
			--image=kindest/node:$(CI_KUBERNETES_VERSION) \
			--kubeconfig $(KIND_CLUSTER_KUBECONFIG) \
			--quiet --wait 120s && \
		KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) scale deployment --replicas 1 coredns --namespace kube-system && \
		$(MAKE) kind/cluster/wait)
	@$(MAKE) --no-print-directory kind/docker/network/connect
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="$(KIND_CLUSTER_KUBECONFIG)"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

.PHONY: kind/docker/network/connect
kind/docker/network/connect: k8s/docker/network/create
	$(Q)for node in $$($(KIND) get nodes --name $(CLUSTER_NAME)); do \
	  docker network connect $(DOCKER_NETWORK) $$node 2>/dev/null || true; \
	done

.PHONY: kind/cluster/wait
kind/cluster/wait:
	@TIMES_TRIED=0; \
	MAX_ALLOWED_TRIES=30; \
	until KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods; do \
		echo "Waiting for the cluster to come up" && sleep 1; \
		TIMES_TRIED=$$((TIMES_TRIED+1)); \
		if [[ $$TIMES_TRIED -ge $$MAX_ALLOWED_TRIES ]]; then \
			KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) get pods -n kube-system -o name | while read pod; do \
				echo "=== Describe $$pod ==="; \
				KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) -n kube-system describe $$pod; \
				echo "\n=== Logs for $$pod ==="; \
				KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) -n kube-system logs $$pod || true; \
			done; \
			exit 1; \
		fi; \
	done

.PHONY: kind/cluster/stop
kind/cluster/stop: kind/cleanup-docker-credentials
	@$(KIND) delete cluster --kubeconfig $(KIND_CLUSTER_KUBECONFIG) --name $(CLUSTER_NAME)
	@rm -f $(KIND_CLUSTER_KUBECONFIG)

.PHONY: kind/clusters/stop
kind/clusters/stop:
	@$(KIND) delete clusters --all
	@rm -f $(KUBECONFIG_DIR)/kind-kuma*.yaml

.PHONY: kind/cluster/load/images
kind/cluster/load/images:
	for image in ${KUMA_IMAGES}; do $(KIND) load docker-image $$image --name=$(CLUSTER_NAME); done

.PHONY: kind/cluster/load
kind/cluster/load: images docker/tag kind/cluster/load/images

.PHONY: kind/cluster/deploy/kuma
kind/cluster/deploy/kuma: build/kumactl kind/cluster/load
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	until \
		KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) get mesh default ; \
	do echo "Waiting for default mesh to be present" && sleep 1; done

.PHONY: kind/cluster/deploy/helm
kind/cluster/deploy/helm: kind/cluster/load
	KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) delete namespace $(KUMA_NAMESPACE) | true
	KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) create namespace $(KUMA_NAMESPACE)
	KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) helm install --namespace $(KUMA_NAMESPACE) \
                --set global.image.registry="$(DOCKER_REGISTRY)" \
                --set global.image.tag="$(BUILD_INFO_VERSION)-${GOARCH}" \
                --set cni.enabled=true \
                --set cni.chained=true \
                --set cni.netDir=/etc/cni/net.d \
                --set cni.binDir=/opt/cni/bin \
                --set cni.confName=10-kindnet.conflist \
                 --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED=true \
                kuma ./deployments/charts/kuma
	KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane

.PHONY: kind/cluster/deploy/kuma/global
kind/cluster/deploy/kuma/global: KUMA_MODE=global
kind/cluster/deploy/kuma/global: kind/cluster/deploy/kuma

.PHONY: kind/cluster/deploy/kuma/local
kind/cluster/deploy/kuma/local: KUMA_MODE=local
kind/cluster/deploy/kuma/local: kind/cluster/deploy/kuma

.PHONY: kind/cluster/deploy/observability
kind/cluster/deploy/observability: build/kumactl
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) ${BUILD_ARTIFACTS_DIR}/kumactl/kumactl install observability | KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n kuma-observability pods -l app=prometheus

.PHONY: kind/cluster/deploy/metrics-server
kind/cluster/deploy/metrics-server:
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.1/components.yaml
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) patch -n kube-system deployment/metrics-server \
		--patch='{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls", "--kubelet-preferred-address-types=InternalIP"]}]}}}}'
	@KUBECONFIG=$(KIND_CLUSTER_KUBECONFIG) $(KUBECTL) wait --timeout=2m --for=condition=Available -n kube-system deployment/metrics-server

LEGACY_TARGET_ALIASES := \
	kind/start=kind/cluster/start \
	kind/wait=kind/cluster/wait \
	kind/stop=kind/cluster/stop \
	kind/stop/all=kind/clusters/stop \
	kind/load/images=kind/cluster/load/images \
	kind/load=kind/cluster/load \
	kind/deploy/kuma=kind/cluster/deploy/kuma \
	kind/deploy/helm=kind/cluster/deploy/helm \
	kind/deploy/kuma/global=kind/cluster/deploy/kuma/global \
	kind/deploy/kuma/local=kind/cluster/deploy/kuma/local \
	kind/deploy/observability=kind/cluster/deploy/observability \
	kind/deploy/metrics-server=kind/cluster/deploy/metrics-server

define _LEGACY_ALIAS
.PHONY: $(1)
$(1):
	@echo "[DEPRECATED] '$(1)' has been renamed to '$(2)'" >&2
	@$$(MAKE) --no-print-directory $(2)
endef

$(foreach m,$(LEGACY_TARGET_ALIASES),\
  $(eval $(call _LEGACY_ALIAS,$(word 1,$(subst =, ,$(m))),$(word 2,$(subst =, ,$(m))))))
