CI_K3S_VERSION ?= $(K8S_MIN_VERSION)
METALLB_VERSION ?= v0.13.9
K3D_VERSION ?= $(shell $(TOP)/$(KUMA_DIR)/mk/dependencies/k3d.sh - get-version)

PROJECT_NAME ?= kuma
KUMA_MODE ?= zone
KUMA_NAMESPACE ?= kuma-system
KUMA_SETTINGS_PREFIX ?=
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
K3D_REGISTRY_FILE ?=
K3D_CLUSTER_CREATE_OPTS ?= -i rancher/k3s:$(CI_K3S_VERSION) \
	--k3s-arg '--disable=traefik@server:0' \
	--k3s-arg '--disable=metrics-server@server:0' \
	--k3s-arg '--kubelet-arg=image-gc-high-threshold=100@server:0' \
	--k3s-arg '--disable=servicelb@server:0' \
    --volume '$(subst @,\@,$(TOP)/$(KUMA_DIR))/test/framework/deployments:/tmp/deployments@server:0' \
	--network kind \
	--port "$(PORT_PREFIX)80-$(PORT_PREFIX)99:30080-30099@server:0" \
	--registry-config "/tmp/.kuma-dev/k3d-registry.yaml" \
	--timeout 120s

ifeq ($(K3D_NETWORK_CNI),calico)
	K3D_CLUSTER_CREATE_OPTS += --k3s-arg '--flannel-backend=none@server:*' --k3s-arg '--disable-network-policy@server:*'
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

K3D_HELM_DEPLOY_OPTS = \
	--set $(KUMA_SETTINGS_PREFIX)global.image.registry="$(DOCKER_REGISTRY)" \
	--set $(KUMA_SETTINGS_PREFIX)controlPlane.image.tag="$(BUILD_INFO_VERSION)" \
	--set $(KUMA_SETTINGS_PREFIX)cni.image.tag="$(BUILD_INFO_VERSION)" \
	--set $(KUMA_SETTINGS_PREFIX)dataPlane.image.tag="$(BUILD_INFO_VERSION)" \
	--set $(KUMA_SETTINGS_PREFIX)dataPlane.initImage.tag="$(BUILD_INFO_VERSION)" \
	--set $(KUMA_SETTINGS_PREFIX)kumactl.image.tag="$(BUILD_INFO_VERSION)"

ifndef K3D_HELM_DEPLOY_NO_CNI
	K3D_HELM_DEPLOY_OPTS += \
		--set $(KUMA_SETTINGS_PREFIX)cni.enabled=true \
		--set $(KUMA_SETTINGS_PREFIX)cni.chained=true \
		--set $(KUMA_SETTINGS_PREFIX)cni.netDir=/var/lib/rancher/k3s/agent/etc/cni/net.d/ \
		--set $(KUMA_SETTINGS_PREFIX)cni.binDir=/bin/ \
		--set $(KUMA_SETTINGS_PREFIX)cni.confName=10-flannel.conflist
endif

ifdef K3D_HELM_DEPLOY_ADDITIONAL_OPTS
	K3D_HELM_DEPLOY_OPTS += $(K3D_HELM_DEPLOY_ADDITIONAL_OPTS)
endif

ifdef K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS
K3D_HELM_DEPLOY_OPTS += $(strip \
  $(foreach setting,$(strip $(K3D_HELM_DEPLOY_ADDITIONAL_SETTINGS)),\
    --set $(KUMA_SETTINGS_PREFIX)$(setting) \
  ) \
)
endif

define maybe_with_flock
  if which flock >/dev/null 2>&1; then \
    SHELL=bash flock -x $(BUILD_DIR)/k3d_network.lock -c $(1); \
  else \
    bash -c $(1); \
  fi
endef

.PHONY: k3d/network/create
k3d/network/create:
	@touch $(BUILD_DIR)/k3d_network.lock && \
		$(call maybe_with_flock,"if ! docker network inspect kind >/dev/null 2>&1; then \
		  docker network create -d=bridge $(KIND_NETWORK_OPTS) kind >/dev/null; \
		fi; \
		echo 'Using docker network: '; \
		docker network inspect kind --format='{{ .Id }}'")
	@rm -f $(BUILD_DIR)/k3d_network.lock

DOCKERHUB_PULL_CREDENTIAL ?=
.PHONY: k3d/setup-docker-credentials
k3d/setup-docker-credentials:
	@mkdir -p /tmp/.kuma-dev ; \
	echo '{"configs": {}}' > /tmp/.kuma-dev/k3d-registry.yaml ; \
	if [[ "$(DOCKERHUB_PULL_CREDENTIAL)" != "" ]]; then \
  		DOCKER_USER=$$(echo "$(DOCKERHUB_PULL_CREDENTIAL)" | cut -d ':' -f 1); \
  		DOCKER_PWD=$$(echo "$(DOCKERHUB_PULL_CREDENTIAL)" | cut -d ':' -f 2); \
  		echo "{\"configs\": {\"registry-1.docker.io\": {\"auth\": {\"username\": \"$${DOCKER_USER}\",\"password\":\"$${DOCKER_PWD}\"}}}}" > /tmp/.kuma-dev/k3d-registry.yaml ; \
  	fi

.PHONY: k3d/cleanup-docker-credentials
k3d/cleanup-docker-credentials:
	@rm -f /tmp/.kuma-dev/k3d-registry.yaml

.PHONY: k3d/start
k3d/start: ${KIND_KUBECONFIG_DIR} k3d/network/create k3d/setup-docker-credentials
	@echo "PORT_PREFIX=$(PORT_PREFIX)"
	@KUBECONFIG=$(KIND_KUBECONFIG) \
		$(K3D_BIN) cluster create "$(KIND_CLUSTER_NAME)" $(K3D_CLUSTER_CREATE_OPTS)
	$(MAKE) k3d/configure/calico
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

.PHONY: k3d/configure/calico
k3d/configure/calico:
ifeq ($(K3D_NETWORK_CNI),calico)
    # https://docs.tigera.io/calico/latest/getting-started/kubernetes/k3s/quickstart
	@KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f https://raw.githubusercontent.com/projectcalico/calico/v3.29.2/manifests/calico.yaml
endif

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
	@# Make sure we only take an IPv4 network
	@IFS=. read -ra NETWORK_ADDR_SPACE <<< "$$(docker network inspect kind --format json | jq -r '.[0].IPAM.Config[].Subnet | select(. | contains(":") | not)')"; \
		IFS=/ read -r _byte prefix <<< "$${NETWORK_ADDR_SPACE[3]}"; \
		    if [[ "$${prefix}" -gt 16 ]]; then echo "Unexpected docker network, expecting a prefix of at most 16 bits"; exit 1; fi; \
		IFS=. read -ra BASE_ADDR_SPACE <<< "$$($(YQ) 'select(.kind == "IPAddressPool") | .spec.addresses[0]' $(KUMA_DIR)/mk/metallb-k3d-$(KIND_CLUSTER_NAME).yaml)"; \
		ADDR_SPACE="$${NETWORK_ADDR_SPACE[0]}.$${NETWORK_ADDR_SPACE[1]}.$${BASE_ADDR_SPACE[2]}.$${BASE_ADDR_SPACE[3]}" \
	      $(YQ) '(select(.kind == "IPAddressPool") | .spec.addresses[0]) = env(ADDR_SPACE)' $(KUMA_DIR)/mk/metallb-k3d-$(KIND_CLUSTER_NAME).yaml | \
		KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) apply -f -

.PHONY: k3d/wait
k3d/wait:
	@TIMES_TRIED=0; \
	MAX_ALLOWED_TRIES=30; \
	until KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) wait -n kube-system --timeout=5s --for condition=Ready --all pods; do \
		echo "Waiting for the cluster to come up" && sleep 1; \
		TIMES_TRIED=$$((TIMES_TRIED+1)); \
		if [[ $$TIMES_TRIED -ge $$MAX_ALLOWED_TRIES ]]; then \
			KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) get pods -n kube-system -o name | while read pod; do \
				echo "=== Describe $$pod ==="; \
				KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) -n kube-system describe $$pod; \
				echo "\n=== Logs for $$pod ==="; \
				KUBECONFIG=$(KIND_KUBECONFIG) $(KUBECTL) -n kube-system logs $$pod || true; \
			done; \
			exit 1; \
		fi; \
	done

.PHONY: k3d/stop
k3d/stop: k3d/cleanup-docker-credentials
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
ifndef K3D_DONT_LOAD
	$(MAKE) images
	$(MAKE) docker/tag
	$(MAKE) k3d/load/images
endif

k3d/deploy/%: export KUBECONFIG = $(KIND_KUBECONFIG)
k3d/deploy/%: KUMACTL := $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl

.PHONY: k3d/deploy/wait/%
k3d/deploy/wait/%: CONDITION = $(word 1,$(subst /, ,$*))
k3d/deploy/wait/%: KIND = $(word 2,$(subst /, ,$*))
k3d/deploy/wait/%: APP = $(word 3,$(subst /, ,$*))
k3d/deploy/wait/%:
	@echo "Waiting for $(KIND) with app=$(APP) to be $(CONDITION)..."
	@$(KUBECTL) wait \
		--namespace $(KUMA_NAMESPACE) \
		--timeout=60s \
		--for=condition=$(CONDITION) \
		--selector app=$(APP) \
		$(KIND) >/dev/null

.PHONY: k3d/deploy/wait/mesh
k3d/deploy/wait/mesh:
	@echo "Waiting for 'default' mesh to be created..."
	@until $(KUBECTL) get mesh default >/dev/null 2>&1; do \
		echo "  Still waiting..."; \
		sleep 1; \
	done

.PHONY: k3d/deploy/wait/cp
k3d/deploy/wait/cp: \
  k3d/deploy/wait/Available/deployments/$(PROJECT_NAME)-control-plane \
  k3d/deploy/wait/Ready/pods/$(PROJECT_NAME)-control-plane \
  k3d/deploy/wait/mesh

.PHONY: k3d/deploy/kumactl/install
k3d/deploy/kumactl/install:
	@$(KUMACTL) install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) \
		| $(KUBECTL) apply -f- \
		| grep --invert-match 'unchanged'
	@echo

.PHONY: k3d/deploy/kumactl/clean
k3d/deploy/kumactl/clean:
	@$(KUMACTL) install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) \
		| $(KUBECTL) delete -f-

.PHONY: k3d/deploy/kumactl
k3d/deploy/kumactl: build/kumactl k3d/load k3d/deploy/kumactl/install k3d/deploy/wait/cp

.PHONY: k3d/restart/kumactl
k3d/restart/kumactl: k3d/stop k3d/start k3d/deploy/kumactl k3d/deploy/demo

### Helm

.PHONY: k3d/deploy/helm/clean/release
k3d/deploy/helm/clean/release:
ifeq (,$(or $(K3D_DEPLOY_HELM_DONT_CLEAN),$(K3D_DEPLOY_HELM_DONT_CLEAN_RELEASE)))
	@echo "Deleting Helm release '$(PROJECT_NAME)'..."
	@helm delete --wait --ignore-not-found --namespace $(KUMA_NAMESPACE) $(PROJECT_NAME) >/dev/null 2>&1

	@echo "Waiting for control plane pods to terminate..."
	@until $(KUBECTL) get pods --namespace $(KUMA_NAMESPACE) --selector app=$(PROJECT_NAME)-control-plane >/dev/null 2>&1; do \
		echo "  Still terminating..."; \
		sleep 1; \
	done
endif

.PHONY: k3d/deploy/helm/clean/ns
k3d/deploy/helm/clean/ns:
ifeq (,$(or $(K3D_DEPLOY_HELM_DONT_CLEAN),$(K3D_DEPLOY_HELM_DONT_CLEAN_NS)))
	@echo "Deleting namespace '$(KUMA_NAMESPACE)'..."
	@$(KUBECTL) delete namespace $(KUMA_NAMESPACE) --force --wait --ignore-not-found >/dev/null 2>&1

	@echo "Waiting for namespace to be fully removed..."
	@until ! $(KUBECTL) get namespace $(KUMA_NAMESPACE) >/dev/null 2>&1; do \
		echo "  Still deleting..."; \
		sleep 1; \
	done
endif

.PHONY: k3d/deploy/helm/clean
k3d/deploy/helm/clean: k3d/deploy/helm/clean/release k3d/deploy/helm/clean/ns

.PHONY: k3d/deploy/helm/upgrade
k3d/deploy/helm/upgrade:
	@echo "Upgrading or installing Helm release '$(PROJECT_NAME)'..."
	@helm upgrade $(PROJECT_NAME) ./deployments/charts/$(PROJECT_NAME) \
		--install \
		--create-namespace \
		--namespace $(KUMA_NAMESPACE) \
		$(strip $(K3D_HELM_DEPLOY_OPTS))

.PHONY: k3d/deploy/helm
k3d/deploy/helm: k3d/load k3d/deploy/helm/clean k3d/deploy/helm/upgrade k3d/deploy/wait/cp

.PHONY: k3d/deploy/demo
k3d/deploy/demo: build/kumactl
	@$(KUMACTL) install demo | $(KUBECTL) apply -f -
	@$(KUBECTL) wait --timeout=60s --for=condition=Ready -n kuma-demo --all pods

# Renamed targets

.PHONY: k3d/deploy/kuma
k3d/deploy/kuma:
	@>&2 echo
	@>&2 echo " #######################################"
	@>&2 echo " #         Target was renamed          #"
	@>&2 echo " #       Use: k3d/deploy/kumactl       #"
	@>&2 echo " #######################################"
	@>&2 echo
	@exit 1

.PHONY: k3d/restart
k3d/restart:
	@>&2 echo
	@>&2 echo " #######################################"
	@>&2 echo " #         Target was renamed          #"
	@>&2 echo " #      Use: k3d/restart/kumactl       #"
	@>&2 echo " #######################################"
	@>&2 echo
	@exit 1
