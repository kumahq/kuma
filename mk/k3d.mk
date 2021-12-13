CI_K3D_VERSION ?= v5.2.0

KUMA_MODE ?= standalone
KUMA_NAMESPACE ?= kuma-system
PORT_PREFIX := $$(( $(subst kuma-,30,$(KIND_CLUSTER_NAME)) - 1 ))

.PHONY: k3d/network/create
k3d/network/create:
	@touch $(BUILD_DIR)/k3d_network.lock
	@flock -x $(BUILD_DIR)/k3d_network.lock -c 'docker network create -d=bridge -o com.docker.network.bridge.enable_ip_masquerade=true --ipv6 --subnet "fd00:fd12:3456::0/64" kind || true;' \
		|| docker network create -d=bridge -o com.docker.network.bridge.enable_ip_masquerade=true --ipv6 --subnet "fd00:fd12:3456::0/64" kind || true
	@rm -f $(BUILD_DIR)/k3d_network.lock

.PHONY: k3d/start
k3d/start: ${KIND_KUBECONFIG_DIR} k3d/network/create
	@KUBECONFIG=$(KIND_KUBECONFIG)  \
		 k3d cluster create "$(KIND_CLUSTER_NAME)" \
			  --k3s-arg '--no-deploy=traefik@server:0' \
			  --k3s-arg '--disable=metrics-server@server:0' \
			  --network kind \
			  --port "$(PORT_PREFIX)80-$(PORT_PREFIX)89:30080-30089@server:0" \
			  --timeout 120s && \
	until \
		 KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
	do echo "Waiting for the cluster to come up" && sleep 1; done
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="$(KIND_KUBECONFIG)"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo

.PHONY: k3d/stop
k3d/stop:
	@KUBECONFIG=$(KIND_KUBECONFIG) k3d cluster delete "$(KIND_CLUSTER_NAME)"

.PHONY: k3d/stop/all
k3d/stop/all:
	@KUBECONFIG=$(KIND_KUBECONFIG) k3d cluster delete --all

.PHONY: k3d/load/images
k3d/load/images:
	@k3d image import \
    	$(KUMA_CP_DOCKER_IMAGE) $(KUMA_DP_DOCKER_IMAGE) \
    	$(KUMA_INIT_DOCKER_IMAGE) $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) \
    	$(KUMACTL_DOCKER_IMAGE) $(KUMA_UNIVERSAL_DOCKER_IMAGE) \
    	--cluster=$(KIND_CLUSTER_NAME) --verbose

.PHONY: k3d/load
k3d/load: images k3d/load/images

.PHONY: k3d/deploy/kuma
k3d/deploy/kuma: build/kumactl k3d/load
	@KUBECONFIG=$(KIND_KUBECONFIG) $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG)  kubectl apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) kumactl install dns | kubectl apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete -n $(EXAMPLE_NAMESPACE) pod -l app=example-app
	@until \
	KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait -n kube-system --timeout=5s --for condition=Ready --all pods ; \
    do \
    	echo "Waiting for the cluster to come up" && sleep 1; \
    done

.PHONY: k3d/deploy/helm
k3d/deploy/helm: k3d/load
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl delete namespace $(KUMA_NAMESPACE) | true
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(KUMA_NAMESPACE)
	@KUBECONFIG=$(KIND_KUBECONFIG) helm install --namespace $(KUMA_NAMESPACE) \
                --set global.image.registry="$(DOCKER_REGISTRY)" \
                --set global.image.tag="$(BUILD_INFO_GIT_TAG)" \
                --set cni.enabled=true \
                --set cni.chained=true \
                --set cni.netDir=/var/lib/rancher/k3s/agent/etc/cni/net.d/ \
                --set cni.binDir=/bin/ \
                --set cni.confName=10-flannel.conflist \
                kuma ./deployments/charts/kuma
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane

.PHONY: k3d/deploy/example-app
k3d/deploy/example-app:
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl create namespace $(EXAMPLE_NAMESPACE) || true
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl annotate namespace $(EXAMPLE_NAMESPACE) kuma.io/sidecar-injection=enabled --overwrite
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl apply -n $(EXAMPLE_NAMESPACE) -f dev/examples/k8s/example-app/example-app.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) kubectl wait --timeout=60s --for=condition=Available -n $(EXAMPLE_NAMESPACE) pods -l app=example-app
