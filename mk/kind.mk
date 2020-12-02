EXAMPLE_NAMESPACE ?= kuma-example
KIND_KUBECONFIG_DIR ?= $(HOME)/.kube
KIND_KUBECONFIG ?= $(KIND_KUBECONFIG_DIR)/kind-kuma-config
KIND_CLUSTER_NAME ?= kuma

ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES := --control-plane-registry=$(DOCKER_REGISTRY) --dataplane-registry=$(DOCKER_REGISTRY) --dataplane-init-registry=$(DOCKER_REGISTRY)
else
	KUMACTL_INSTALL_CONTROL_PLANE_IMAGES :=
endif
ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	KUMACTL_INSTALL_METRICS_IMAGES := --kuma-prometheus-sd-image=$(KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME)
else
	KUMACTL_INSTALL_METRICS_IMAGES :=
endif

define KIND_EXAMPLE_DATAPLANE_MESH
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_MESH)
endef
define KIND_EXAMPLE_DATAPLANE_NAME
$(shell KUBECONFIG=$(KIND_KUBECONFIG) kubectl -n $(EXAMPLE_NAMESPACE) exec $$(kubectl -n $(EXAMPLE_NAMESPACE) get pods -l app=example-app -o=jsonpath='{.items[0].metadata.name}') -c kuma-sidecar printenv KUMA_DATAPLANE_NAME)
endef

CI_KIND_VERSION ?= v0.9.0
CI_KUBERNETES_VERSION ?= v1.18.8@sha256:f4bcc97a0ad6e7abaf3f643d890add7efe6ee4ab90baeb374b4f41a4c95567eb

KIND_PATH := $(CI_TOOLS_DIR)/kind

KUMA_MODE ?= standalone
KUMA_NAMESPACE ?= kuma-system

.PHONY: ${KIND_KUBECONFIG_DIR}
${KIND_KUBECONFIG_DIR}:
	@mkdir -p ${KIND_KUBECONFIG_DIR}

.PHONY: kind/start
kind/start: ${KIND_KUBECONFIG_DIR}
	@kind get clusters | grep $(KIND_CLUSTER_NAME) >/dev/null 2>&1 && echo "Kind cluster already running." && exit 0 || \
		(kind create cluster \
			--name "$(KIND_CLUSTER_NAME)" \
			--image=kindest/node:$(CI_KUBERNETES_VERSION) \
			--kubeconfig $(KIND_KUBECONFIG) \
			--quiet --wait 120s && \
		KUBECONFIG=$(KIND_KUBECONFIG) kubectl scale deployment --replicas 1 coredns --namespace kube-system && \
		until \
			KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
		do echo "Waiting for the cluster to come up" && sleep 1; done )
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="${KIND_KUBECONFIG}"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

.PHONY: kind/stop
kind/stop:
	@kind delete cluster --name $(KIND_CLUSTER_NAME)

.PHONY: kind/stop/all
kind/stop/all:
	@kind delete clusters --all

.PHONY: kind/load/control-plane
kind/load/control-plane:
	@kind load docker-image $(KUMA_CP_DOCKER_IMAGE) --name=$(KIND_CLUSTER_NAME)

.PHONY: kind/load/kuma-dp
kind/load/kuma-dp:
	@kind load docker-image $(KUMA_DP_DOCKER_IMAGE) --name=$(KIND_CLUSTER_NAME)

.PHONY: kind/load/kuma-init
kind/load/kuma-init:
	@kind load docker-image $(KUMA_INIT_DOCKER_IMAGE) --name=$(KIND_CLUSTER_NAME)

.PHONY: kind/load/kuma-prometheus-sd
kind/load/kuma-prometheus-sd:
	@kind load docker-image $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) --name=$(KIND_CLUSTER_NAME)

.PHONY: kind/load/images
kind/load/images: kind/load/control-plane kind/load/kuma-dp kind/load/kuma-init kind/load/kuma-prometheus-sd

.PHONY: kind/load
kind/load: image/kuma-cp image/kuma-dp image/kuma-init image/kuma-prometheus-sd kind/load/images

.PHONY: kind/deploy/kuma
kind/deploy/kuma: build/kumactl kind/load
	@${BUILD_ARTIFACTS_DIR}/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG)  kubectl apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete -n $(EXAMPLE_NAMESPACE) pod -l app=example-app
	@until \
    	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
    do \
    	echo "Waiting for the cluster to come up" && sleep 1; \
    done

.PHONY: kind/deploy/helm
kind/deploy/helm: kind/load
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete namespace $(KUMA_NAMESPACE) | true
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(KUMA_NAMESPACE)
	KUBECONFIG=$(KIND_KUBECONFIG) helm install --namespace $(KUMA_NAMESPACE) --set global.image.registry="$(DOCKER_REGISTRY)",global.image.tag="$(BUILD_INFO_GIT_TAG)",cni.enabled=true kuma ./deployments/charts/kuma
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane

.PHONY: kind/deploy/kuma/global
kind/deploy/kuma/global: KUMA_MODE=global
kind/deploy/kuma/global: kind/deploy/kuma

.PHONY: kind/deploy/kuma/local
kind/deploy/kuma/local: KUMA_MODE=local
kind/deploy/kuma/local: kind/deploy/kuma

.PHONY: kind/deploy/metrics
kind/deploy/metrics: build/kumactl
	@${BUILD_ARTIFACTS_DIR}/kumactl/kumactl install metrics $(KUMACTL_INSTALL_METRICS_IMAGES) | kubectl apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n kuma-metrics pods -l app=prometheus

.PHONY: kind/deploy/metrics-server
kind/deploy/metrics-server:
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/download/v0.4.1/components.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl patch -n kube-system deployment/metrics-server \
		--patch='{"spec":{"template":{"spec":{"containers":[{"name":"metrics-server","args":["--cert-dir=/tmp", "--secure-port=4443", "--kubelet-insecure-tls", "--kubelet-preferred-address-types=InternalIP"]}]}}}}'
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n kube-system deployment/metrics-server

.PHONY: kind/deploy/example-app
kind/deploy/example-app:
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(EXAMPLE_NAMESPACE) || true
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl annotate namespace $(EXAMPLE_NAMESPACE) kuma.io/sidecar-injection=enabled --overwrite
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/example-app/example-app.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=120s --for=condition=Available -n $(EXAMPLE_NAMESPACE) deployment/example-app
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(EXAMPLE_NAMESPACE) pods -l app=example-app

.PHONY: run/k8s
run/k8s: fmt vet ## Dev: Run Control Plane locally in Kubernetes mode
	@KUBECONFIG=$(KIND_KUBECONFIG) $(MAKE) crd/upgrade -C pkg/plugins/resources/k8s/native
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
