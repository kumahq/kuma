CI_K3S_VERSION ?= $(K8S_MIN_VERSION)
METALLB_VERSION ?= v0.13.9
K3D_VERSION ?= $(shell $(TOP)/mk/dependencies/k3d.sh - get-version)

KUMA_MODE ?= zone
KUMA_NAMESPACE ?= kuma-system
# Comment about PORT_PREFIX generation
#
# First step: $(KIND_CLUSTER_NAME:kuma%=300%) will replace a string "kuma" from
# the $(KIND_CLUSTER_NAME) variable with the string "300" (default/initial
# prefix):
#
#  Initial value				Step#1
#  KIND_CLUSTER_NAME(kuma) ->	300
#  KIND_CLUSTER_NAME(kuma-1) ->	300-1
#  KIND_CLUSTER_NAME(kuma-2) ->	300-2
#  KIND_CLUSTER_NAME(kuma-3) ->	300-3
#  [...etc]
#
# The next step - $(patsubst 300-%,300+%-1,...) will replace string
# "300-[1,2,3...]" with string "300+[1,2,3...]-1" ("-1" is necessary to preserve
# the current overflow, so when the KIND_CLUSTER_NAME is equal "kuma", OR
# "kuma-1" when value of the port will be equal "300"):
#
#  Initial value				Step#1		Step#2
#  KIND_CLUSTER_NAME(kuma) ->	300 ->		300
#  KIND_CLUSTER_NAME(kuma-1) ->	300-1 ->	300+1-1
#  KIND_CLUSTER_NAME(kuma-2) ->	300-2 ->	300+2-1
#  KIND_CLUSTER_NAME(kuma-3) ->	300-3 ->	300+3-1
#  [...etc]
#
# The last step $$((...)) will call the shell to use the expression we generated
# earlier and calculate it's arithmetic value:
#
#  Initial value				Step#1		Step#2		Step#3	Result
#  KIND_CLUSTER_NAME(kuma) ->	300 ->		300 ->		300 ->	PORT_PREFIX(300)
#  KIND_CLUSTER_NAME(kuma-1) ->	300-1 ->	300+1-1 ->	300 ->	PORT_PREFIX(300)
#  KIND_CLUSTER_NAME(kuma-2) ->	300-2 ->	300+2-1 ->	301 ->	PORT_PREFIX(301)
#  KIND_CLUSTER_NAME(kuma-3) ->	300-3 ->	300+3-1 ->	302 ->	PORT_PREFIX(302)
#  [...etc]
PORT_PREFIX := $$(($(patsubst 300-%,300+%-1,$(KIND_CLUSTER_NAME:kuma%=300%))))

K3D_NETWORK_CNI ?= flannel
K3D_CLUSTER_CREATE_OPTS ?= -i rancher/k3s:$(CI_K3S_VERSION) \
	--k3s-arg '--disable=traefik@server:0' \
	--k3s-arg '--disable=metrics-server@server:0' \
	--k3s-arg '--kubelet-arg=image-gc-high-threshold=100@server:0' \
	--k3s-arg '--disable=servicelb@server:0' \
    --volume '$(subst @,\@,$(TOP)/$(KUMA_DIR))/test/framework/deployments:/tmp/deployments@server:0' \
	--network kind \
	--port "$(PORT_PREFIX)80-$(PORT_PREFIX)99:30080-30099@server:0" \
	--timeout 120s

ifeq ($(K3D_NETWORK_CNI),calico)
	K3D_CLUSTER_CREATE_OPTS += --volume "$(TOP)/$(KUMA_DIR)/test/k3d/calico.$(K3D_VERSION).yaml:/var/lib/rancher/k3s/server/manifests/calico.yaml" \
		--k3s-arg '--flannel-backend=none@server:*' --k3s-arg '--disable-network-policy@server:*'
endif

ifdef CI
ifeq ($(GOOS),linux)
ifneq (,$(findstring legacy,$(CIRCLE_JOB)))
	K3D_CLUSTER_CREATE_OPTS += --volume "/sys/fs/bpf:/sys/fs/bpf:shared"
endif
endif
endif

ifeq ($(GOOS),linux)
ifndef CI
	K3D_CLUSTER_CREATE_OPTS += --volume "/sys/fs/bpf:/sys/fs/bpf:shared"
endif
endif

KIND_NETWORK_OPTS =  -o com.docker.network.bridge.enable_ip_masquerade=true
ifdef IPV6
    KIND_NETWORK_OPTS += --ipv6 --subnet "fd00:fd12:3456::0/64"
endif

.PHONY: k3d/network/create
k3d/network/create:
	@touch $(BUILD_DIR)/k3d_network.lock && \
		if [ `which flock` ]; then flock -x $(BUILD_DIR)/k3d_network.lock -c 'docker network create -d=bridge $(KIND_NETWORK_OPTS) kind || true'; \
		else docker network create -d=bridge $(KIND_NETWORK_OPTS) kind || true; fi && \
		rm -f $(BUILD_DIR)/k3d_network.lock

$(TOP)/$(KUMA_DIR)/test/k3d/calico.$(K3D_VERSION).yaml:
	@mkdir -p $(TOP)/$(KUMA_DIR)/test/k3d
	curl --location --fail --silent --retry 5 \
		-o $(TOP)/$(KUMA_DIR)/test/k3d/calico.$(K3D_VERSION).yaml \
		https://k3d.io/v$(K3D_VERSION)/usage/advanced/calico.yaml

.PHONY: k3d/start
k3d/start: ${KIND_KUBECONFIG_DIR} k3d/network/create \
	$(if $(findstring calico,$(K3D_NETWORK_CNI)),$(TOP)/$(KUMA_DIR)/test/k3d/calico.$(K3D_VERSION).yaml)
	@echo "PORT_PREFIX=$(PORT_PREFIX)"
	@KUBECONFIG=$(KIND_KUBECONFIG) \
		$(K3D_BIN) cluster create "$(KIND_CLUSTER_NAME)" $(K3D_CLUSTER_CREATE_OPTS)
	$(MAKE) k3d/wait
	@echo
	@echo '>>> You need to manually run the following command in your shell: >>>'
	@echo
	@echo export KUBECONFIG="$(KIND_KUBECONFIG)"
	@echo
	@echo '<<< ------------------------------------------------------------- <<<'
	@echo
	$(MAKE) k3d/configure/ebpf
	$(MAKE) k3d/configure/metallb

.PHONY: k3d/configure/ebpf
k3d/configure/ebpf:
ifeq ($(GOOS),darwin)
	docker exec k3d-$(KIND_CLUSTER_NAME)-server-0 mount bpffs /sys/fs/bpf -t bpf && \
	docker exec k3d-$(KIND_CLUSTER_NAME)-server-0 mount --make-shared /sys/fs/bpf
endif

.PHONY: k3d/configure/metallb
k3d/configure/metallb:
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f https://raw.githubusercontent.com/metallb/metallb/$(METALLB_VERSION)/config/manifests/metallb-native.yaml
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=120s --for=condition=Ready -n metallb-system --all pods
	@# Construct a valid address space from the docker network and the template IPAddressPool
	@IFS=. read -ra NETWORK_ADDR_SPACE <<< "$$(docker network inspect kind --format '{{ (index .IPAM.Config 0).Subnet }}')"; \
		IFS=/ read -r _byte prefix <<< "$${NETWORK_ADDR_SPACE[3]}"; \
		    if [[ "$${prefix}" -gt 16 ]]; then echo "Unexpected docker network, expecting a prefix of at most 16 bits"; exit 1; fi; \
		IFS=. read -ra BASE_ADDR_SPACE <<< "$$(yq 'select(.kind == "IPAddressPool") | .spec.addresses[0]' $(KUMA_DIR)/mk/metallb-k3d-$(KIND_CLUSTER_NAME).yaml)"; \
		ADDR_SPACE="$${NETWORK_ADDR_SPACE[0]}.$${NETWORK_ADDR_SPACE[1]}.$${BASE_ADDR_SPACE[2]}.$${BASE_ADDR_SPACE[3]}" \
	      yq '(select(.kind == "IPAddressPool") | .spec.addresses[0]) = env(ADDR_SPACE)' $(KUMA_DIR)/mk/metallb-k3d-$(KIND_CLUSTER_NAME).yaml | \
		KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f -

.PHONY: k3d/wait
k3d/wait:
	@TIMES_TRIED=0; \
	MAX_ALLOWED_TRIES=30; \
	until KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods; do \
    	echo "Waiting for the cluster to come up" && sleep 1; \
  		TIMES_TRIED=$$((TIMES_TRIED+1)); \
  		if [[ $$TIMES_TRIED -ge $$MAX_ALLOWED_TRIES ]]; then KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) get pods -n kube-system -o Name | KUBECONFIG=$(KIND_KUBECONFIG) xargs -I % $(KUBECTL) -n kube-system describe %; exit 1; fi \
    done

.PHONY: k3d/stop
k3d/stop:
	@KUBECONFIG=$(KIND_KUBECONFIG) $(K3D_BIN) cluster delete "$(KIND_CLUSTER_NAME)"

.PHONY: k3d/stop/all
k3d/stop/all:
	@KUBECONFIG=$(KIND_KUBECONFIG) $(K3D_BIN) cluster delete --all

.PHONY: k3d/load/images
k3d/load/images:
	# https://github.com/k3d-io/k3d/issues/900 can cause failures that simple retry will fix
	for i in 1 2 3 4 5; do $(K3D_BIN) image import --mode=direct $(KUMA_IMAGES) --cluster=$(KIND_CLUSTER_NAME) --verbose && s=0 && break || s=$$? && echo "Image import failed. Retrying..."; done; (exit $$s)

.PHONY: k3d/load
k3d/load:
	$(MAKE) images
	$(MAKE) docker/tag
	$(MAKE) k3d/load/images

.PHONY: k3d/deploy/kuma
k3d/deploy/kuma: build/kumactl k3d/load
	@KUBECONFIG=$(KIND_KUBECONFIG) $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) | KUBECONFIG=$(KIND_KUBECONFIG)  $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane
	until \
		 KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) get mesh default ; \
	do echo "Waiting for default mesh to be present" && sleep 1; done

.PHONY: k3d/deploy/helm
k3d/deploy/helm: k3d/load
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) delete namespace $(KUMA_NAMESPACE) --wait | true
	KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) create namespace $(KUMA_NAMESPACE)
	KUBECONFIG=$(KIND_KUBECONFIG) helm upgrade --install --namespace $(KUMA_NAMESPACE) \
                --set global.image.registry="$(DOCKER_REGISTRY)" \
                --set global.image.tag="$(BUILD_INFO_VERSION)" \
                --set cni.enabled=true \
                --set cni.chained=true \
                --set cni.netDir=/var/lib/rancher/k3s/agent/etc/cni/net.d/ \
                --set cni.binDir=/bin/ \
                --set cni.confName=10-flannel.conflist \
                kuma ./deployments/charts/kuma
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Available -n $(KUMA_NAMESPACE) deployment/kuma-control-plane
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n $(KUMA_NAMESPACE) pods -l app=kuma-control-plane

.PHONY: k3d/deploy/demo
k3d/deploy/demo: build/kumactl
	@$(BUILD_ARTIFACTS_DIR)/kumactl/kumactl install demo | KUBECONFIG=$(KIND_KUBECONFIG)  $(KUBECTL) apply -f -
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait --timeout=60s --for=condition=Ready -n kuma-demo --all pods

.PHONY: k3d/restart
k3d/restart:
	$(MAKE) k3d/stop
	$(MAKE) k3d/start
	$(MAKE) k3d/deploy/kuma
	$(MAKE) k3d/deploy/demo
