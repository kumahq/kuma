KIND_CLUSTER_NAME ?= kuma

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KIND_KUBECONFIG_DIR := $(HOME)/.kube

# This is the name of the current config file to use.
KIND_KUBECONFIG := $(KIND_KUBECONFIG_DIR)/kind-$(KIND_CLUSTER_NAME)-config

# Ensure Kubernetes tooling only gets the config we explicitly specify.
unexport KUBECONFIG

ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-registry=$(DOCKER_REGISTRY) --dataplane-registry=$(DOCKER_REGISTRY) --dataplane-init-registry=$(DOCKER_REGISTRY)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif

CI_KUBERNETES_VERSION ?= v1.23.17@sha256:59c989ff8a517a93127d4a536e7014d28e235fb3529d9fba91b3951d461edfdb

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

.PHONY: kind/start
kind/start: ${KUBECONFIG_DIR} kind/setup-docker-credentials
	$(KIND) get clusters | grep $(KIND_CLUSTER_NAME) >/dev/null 2>&1 && echo "Kind cluster already running." && exit 0 || \
		($(KIND) create cluster \
			--name "$(KIND_CLUSTER_NAME)" \
			--config "$(KUMA_DIR)/test/kind/cluster-$(if $(IPV6),ipv6-,)$(KIND_CLUSTER_NAME).yaml" \
			--image=kindest/node:$(CI_KUBERNETES_VERSION) \
			--kubeconfig $(KIND_KUBECONFIG) \
			--quiet --wait 120s && \
		KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) scale deployment --replicas 1 coredns --namespace kube-system && \
		$(MAKE) kind/wait)
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="$(KIND_KUBECONFIG)"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

.PHONY: kind/wait
kind/wait:
	@TIMES_TRIED=0; \
	MAX_ALLOWED_TRIES=30; \
	until KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods; do \
    	echo "Waiting for the cluster to come up" && sleep 1; \
  		TIMES_TRIED=$$((TIMES_TRIED+1)); \
  		if [[ $$TIMES_TRIED -ge $$MAX_ALLOWED_TRIES ]]; then KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) get pods -n kube-system -o Name | KUBECONFIG=$(KIND_KUBECONFIG) xargs -I % $(KUBECTL) -n kube-system describe %; exit 1; fi \
    done

.PHONY: kind/stop
kind/stop: kind/cleanup-docker-credentials
	@$(KIND) delete cluster --name $(KIND_CLUSTER_NAME)
	@rm -f $(KUBECONFIG_DIR)/$(KIND_KUBECONFIG)

.PHONY: kind/stop/all
kind/stop/all:
	@$(KIND) delete clusters --all
	@rm -f $(KUBECONFIG_DIR)/kind-kuma-*

.PHONY: kind/load/images
kind/load/images:
	for image in ${KUMA_IMAGES}; do $(KIND) load docker-image $$image --name=$(KIND_CLUSTER_NAME); done

.PHONY: kind/load
kind/load: images docker/tag kind/load/images

.PHONY: kind/deploy/kuma
kind/deploy/kuma: build/kumactl kind/load
	@KUBECONFIG=$(KIND_KUBECONFIG) $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	until \
		KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) get mesh default ; \
	do echo "Waiting for default mesh to be present" && sleep 1; done

.PHONY: kind/deploy/helm
kind/deploy/helm: kind/load
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) delete namespace $(KUMA_NAMESPACE) | true
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) create namespace $(KUMA_NAMESPACE)
	KUBECONFIG=$(KIND_KUBECONFIG) helm install --namespace $(KUMA_NAMESPACE) \
                --set global.image.registry="$(DOCKER_REGISTRY)" \
                --set global.image.tag="$(BUILD_INFO_VERSION)-${GOARCH}" \
                --set cni.enabled=true \
                --set cni.chained=true \
                --set cni.netDir=/etc/cni/net.d \
                --set cni.binDir=/opt/cni/bin \
                --set cni.confName=10-kindnet.conflist \
                 --set controlPlane.envVars.KUMA_RUNTIME_KUBERNETES_INJECTOR_BUILTIN_DNS_ENABLED=true \
                kuma ./deployments/charts/kuma
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane

.PHONY: kind/deploy/kuma/global
kind/deploy/kuma/global: KUMA_MODE=global
kind/deploy/kuma/global: kind/deploy/kuma

.PHONY: kind/deploy/kuma/local
kind/deploy/kuma/local: KUMA_MODE=local
kind/deploy/kuma/local: kind/deploy/kuma

.PHONY: kind/deploy/observability
kind/deploy/observability: build/kumactl
	@KUBECONFIG=$(KIND_KUBECONFIG) ${BUILD_ARTIFACTS_DIR}/kumactl/kumactl install observability | KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n kuma-observability pods -l app=prometheus

.PHONY: kind/deploy/metrics-server
kind/deploy/metrics-server:
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.1/components.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) patch -n kube-system deployment/metrics-server \
		--patch='{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls", "--kubelet-preferred-address-types=InternalIP"]}]}}}}'
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=2m --for=condition=Available -n kube-system deployment/metrics-server
