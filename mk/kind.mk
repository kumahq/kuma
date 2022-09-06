EXAMPLE_NAMESPACE ?= kuma-example
KIND_CLUSTER_NAME ?= kuma

# The e2e tests depend on Kind kubeconfigs being in this directory,
# so this is location should not be changed by developers.
KIND_KUBECONFIG_DIR := $(HOME)/.kube

# This is the name of the current config file to use.
KIND_KUBECONFIG := $(KIND_KUBECONFIG_DIR)/kind-$(KIND_CLUSTER_NAME)-config

# Ensure Kubernetes tooling only gets the config we explicity specify.
unexport KUBECONFIG

METRICS_SERVER_VERSION := 0.4.1

ifdef IPV6
KIND_CONFIG ?= $(TOP)/test/kind/cluster-ipv6.yaml
else
KIND_CONFIG ?= $(TOP)/test/kind/cluster.yaml
endif

ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-registry=$(DOCKER_REGISTRY) --dataplane-registry=$(DOCKER_REGISTRY) --dataplane-init-registry=$(DOCKER_REGISTRY)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif

define KIND_EXAMPLE_DATAPLANE_MESH
$(shell KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) -n $(EXAMPLE_NAMESPACE) exec $$($(KUBECTL) -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_MESH)
endef
define KIND_EXAMPLE_DATAPLANE_NAME
$(shell KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) -n $(EXAMPLE_NAMESPACE) exec $$($(KUBECTL) -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_NAME)
endef

CI_KUBERNETES_VERSION ?= v1.20.7@sha256:cbeaf907fc78ac97ce7b625e4bf0de16e3ea725daf6b04f930bd14c67c671ff9

KUMA_MODE ?= standalone
KUMA_NAMESPACE ?= kuma-system

.PHONY: kind/start
kind/start: ${KUBECONFIG_DIR}
	@$(KIND) get clusters | grep $(KIND_CLUSTER_NAME) >/dev/null 2>&1 && echo "Kind cluster already running." && exit 0 || \
		($(KIND) create cluster \
			--name "$(KIND_CLUSTER_NAME)" \
			--config "$(KIND_CONFIG)" \
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
	until \
		KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
	do echo "Waiting for the cluster to come up" && sleep 1; done

.PHONY: kind/stop
kind/stop:
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
kind/load: images kind/load/images

.PHONY: kind/deploy/kuma
kind/deploy/kuma: build/kumactl kind/load
	@KUBECONFIG=$(KIND_KUBECONFIG) $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) delete -n $(EXAMPLE_NAMESPACE) pod -l app=example-app
	@until \
      KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
    do \
      echo "Waiting for the cluster to come up" && sleep 1; \
    done

.PHONY: kind/deploy/helm
kind/deploy/helm: kind/load
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) delete namespace $(KUMA_NAMESPACE) | true
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) create namespace $(KUMA_NAMESPACE)
	KUBECONFIG=$(KIND_KUBECONFIG) helm install --namespace $(KUMA_NAMESPACE) \
                --set global.image.registry="$(DOCKER_REGISTRY)" \
                --set global.image.tag="$(BUILD_INFO_GIT_TAG)" \
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
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v$(METRICS_SERVER_VERSION)/components.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) patch -n kube-system deployment/metrics-server \
		--patch='{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls", "--kubelet-preferred-address-types=InternalIP"]}]}}}}'
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=2m --for=condition=Available -n kube-system deployment/metrics-server

.PHONY: kind/deploy/example-app
kind/deploy/example-app:
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/meshes/no-passthrough.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/external-services/httpbin.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/external-services/mockbin.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) create namespace $(EXAMPLE_NAMESPACE) || true
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) annotate namespace $(EXAMPLE_NAMESPACE) kuma.io/sidecar-injection=enabled --overwrite
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/example-app/example-app.yaml

.PHONY: run/k8s
run/k8s: generate/builtin-crds ## Dev: Run Control Plane locally in Kubernetes mode
	$(KUBECTL) diff -f pkg/plugins/resources/k8s/native/config/crd/bases || $(KUBECTL) apply -f pkg/plugins/resources/k8s/native/config/crd/bases
	KUBECONFIG=$(KIND_KUBECONFIG) \
	KUMA_ENVIRONMENT=kubernetes \
	KUMA_STORE_TYPE=kubernetes \
	KUMA_SDS_SERVER_TLS_CERT_FILE=app/kuma-cp/cmd/testdata/tls.crt \
	KUMA_SDS_SERVER_TLS_KEY_FILE=app/kuma-cp/cmd/testdata/tls.key \
	KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_PORT=$(CP_K8S_ADMISSION_PORT) \
	KUMA_RUNTIME_KUBERNETES_ADMISSION_SERVER_CERT_DIR=app/kuma-cp/cmd/testdata \
	$(GO_RUN) ./app/kuma-cp/main.go run --log-level=debug

run/example/envoy/k8s: EXAMPLE_DATAPLANE_MESH=$(KIND_EXAMPLE_DATAPLANE_MESH)
run/example/envoy/k8s: EXAMPLE_DATAPLANE_NAME=$(KIND_EXAMPLE_DATAPLANE_NAME)
run/example/envoy/k8s: run/example/envoy
