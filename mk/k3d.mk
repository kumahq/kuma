# Auto-include dependencies (guarded, safe to include multiple times)
include $(dir $(lastword $(MAKEFILE_LIST)))common.mk
include $(dir $(lastword $(MAKEFILE_LIST)))k8s.mk

# --- K3s ---

K3S_VERSION ?= $(K8S_MIN_VERSION)
ifdef CI_K3S_VERSION
ifeq ($(origin K3S_VERSION), default)
K3S_VERSION := $(CI_K3S_VERSION)
endif
endif

K3S_IMAGE   ?= rancher/k3s:$(K3S_VERSION)

# --- MetalLB ---

# renovate: datasource=github-tags depName=metallb packageName=metallb/metallb versioning=semver
METALLB_VERSION ?= v0.15.3
METALLB_MANIFESTS ?= https://raw.githubusercontent.com/metallb/metallb/$(METALLB_VERSION)/config/manifests/metallb-native.yaml
METALLB_NAMESPACE ?= metallb-system

# --- Calico ---

# renovate: datasource=github-releases depName=projectcalico/tigera-operator packageName=projectcalico/calico versioning=semver
CALICO_VERSION ?= v3.31.4
CALICO_NAMESPACE ?= tigera-operator
CALICO_HELM_REPO_ADDR ?= https://docs.tigera.io/calico/charts
CALICO_HELM_REPO_NAME ?= projectcalico
CALICO_HELM_RELEASE ?= calico
CALICO_HELM_CHART ?= $(CALICO_HELM_REPO_NAME)/tigera-operator

# --- Cluster settings ---

PROJECT_NAME ?= kuma
KUMA_MODE ?= zone
KUMA_NAMESPACE ?= kuma-system
KUMA_SETTINGS_PREFIX ?=

K3D_REGISTRY_FILE ?= $(TMP_DIR_K8S)/k3d-registry.json

# --- CNI ---
# Validate and normalize K3D_CNI before it's used in K3D_CLUSTER_CREATE_OPTS.

K3D_CNI ?= flannel
ifdef K3D_NETWORK_CNI
ifeq ($(origin K3D_CNI), default)
K3D_CNI := $(K3D_NETWORK_CNI)
endif
endif

K3D_CNI := $(strip $(K3D_CNI))
ifeq ($(K3D_CNI),)
  override K3D_CNI := flannel
else ifneq ($(filter $(K3D_CNI),flannel calico),$(K3D_CNI))
  $(warning Unsupported K3D_CNI '$(K3D_CNI)', using flannel)
  override K3D_CNI := flannel
endif

# --- Component disable list ---

K3D_DISABLE_DEFAULT := traefik servicelb metrics-server
# Re-enable components: K3D_ENABLE="traefik metrics-server"
K3D_ENABLE ?=
K3D_DISABLE := $(filter-out $(strip $(K3D_ENABLE)),$(K3D_DISABLE_DEFAULT))

# --- Cluster create options ---

K3D_CLUSTER_CREATE_OPTS ?= \
	--image "$(K3S_IMAGE)" \
	--volume "$(subst @,\@,$(TOP)/$(KUMA_DIR))/test/framework/deployments:/tmp/deployments@server:0" \
	--network "$(DOCKER_NETWORK)" \
	--timeout "120s" \
	--k3s-arg "--kubelet-arg=image-gc-high-threshold=100@server:0" \
	--registry-config "$(K3D_REGISTRY_FILE)"

ifeq ($(K3D_CNI),calico)
	K3D_CLUSTER_CREATE_OPTS += \
		--k3s-arg "--flannel-backend=none@server:*" \
		--k3s-arg "--disable-network-policy@server:*"
endif

K3D_CLUSTER_CREATE_OPTS += $(foreach c,$(K3D_DISABLE),--k3s-arg "--disable=$(c)@server:0")

# Relax kubelet eviction thresholds for local/dev clusters.
# k3d runs k3s inside Docker with small loopback-backed filesystems, which can
# trigger DiskPressure evictions under default thresholds.
# Set K3D_TWEAK_FS_EVICTION_RULES=false to use k3s defaults.
# Ref: https://github.com/k3d-io/k3d/issues/133
K3D_TWEAK_FS_EVICTION_RULES ?= true

ifeq ($(K3D_TWEAK_FS_EVICTION_RULES),true)
	K3D_CLUSTER_CREATE_OPTS += \
		--k3s-arg "--kubelet-arg=eviction-hard=imagefs.available<1%,nodefs.available<1%@agent:*" \
		--k3s-arg "--kubelet-arg=eviction-minimum-reclaim=imagefs.available=1%,nodefs.available=1%@agent:*" \
		--k3s-arg "--kubelet-arg=eviction-hard=imagefs.available<1%,nodefs.available<1%@server:0" \
		--k3s-arg "--kubelet-arg=eviction-minimum-reclaim=imagefs.available=1%,nodefs.available=1%@server:0"
endif

# --- eBPF ---
# Mount bpffs inside k3d containers for eBPF-based transparent proxy.
# On macOS Docker Desktop the mount must happen post-create; on Linux
# it's done via a volume at cluster creation time.

ifeq ($(GOOS),linux)
K3D_CLUSTER_CREATE_OPTS += --volume "/sys/fs/bpf:/sys/fs/bpf:shared"
endif

# --- k3d-specific context ---

CLUSTER_KUBECONTEXT := k3d-$(CLUSTER_NAME)

# --- Target-specific variables ---

k3d/%: KUMACTL := $(BUILD_ARTIFACTS_DIR)/kumactl/kumactl
k3d/cluster/%: export KUBECONFIG  = $(K3D_CLUSTER_KUBECONFIG)
k3d/cluster/%: export KUBECONTEXT = $(CLUSTER_KUBECONTEXT)

# --- Diagnostics ---

.PHONY: k3d/cluster/diagnose
k3d/cluster/diagnose: NAMESPACE ?= kube-system
k3d/cluster/diagnose:
	$(Q)$(KUBECTL) --namespace $(NAMESPACE) get pods --output name | while read -r pod; do \
	  echo "### Describe $$pod ###"; $(KUBECTL) --namespace $(NAMESPACE) describe "$$pod" || true; \
	  echo "### Logs for $$pod ###"; $(KUBECTL) --namespace $(NAMESPACE) logs "$$pod" || true; \
	done || true

# --- Image loading ---
# Sequential $(MAKE) calls ensure images are built and tagged before import,
# which is required because `k3d image import` needs the tagged images to exist.

.PHONY: k3d/cluster/load
k3d/cluster/load:
ifndef K3D_DONT_LOAD
	$(Q)$(MAKE) images
	$(Q)$(MAKE) docker/tag
	$(Q)$(MAKE) k3d/cluster/load/images
endif

K3D_IMAGE_IMPORT_MODE    ?= direct
K3D_IMAGE_IMPORT_RETRIES ?= 5
K3D_IMAGE_IMPORT_VERBOSE ?= true
K3D_IMAGE_IMPORT_VERBOSE_FLAG := $(if $(filter true 1,$(K3D_IMAGE_IMPORT_VERBOSE)),--verbose,)

# The import command, factored out so the retry wrapper stays generic
_k3d_import = $(K3D) image import \
  --mode $(K3D_IMAGE_IMPORT_MODE) \
  --cluster $(CLUSTER_NAME) \
  $(K3D_IMAGE_IMPORT_VERBOSE_FLAG) \
  $(KUMA_IMAGES)

.PHONY: k3d/cluster/load/images
k3d/cluster/load/images:
	$(Q)$(call _retry,$(_k3d_import),$(K3D_IMAGE_IMPORT_RETRIES),k3d image import to $(CLUSTER_NAME))

# --- CNI setup ---

.PHONY: k3d/cluster/cni/setup
k3d/cluster/cni/setup: k3d/cluster/cni/setup/$(K3D_CNI) ; @

.PHONY: k3d/cluster/cni/setup/flannel
k3d/cluster/cni/setup/flannel: ; @

.PHONY: k3d/cluster/cni/setup/calico
k3d/cluster/cni/setup/calico:
	$(Q)$(HELM) repo add $(CALICO_HELM_REPO_NAME) $(CALICO_HELM_REPO_ADDR) >/dev/null
	$(Q)$(HELM) repo update $(CALICO_HELM_REPO_NAME) >/dev/null
	$(Q)$(HELM) upgrade $(CALICO_HELM_RELEASE) $(CALICO_HELM_CHART) \
		--install --create-namespace --wait \
		--kube-context $(KUBECONTEXT) \
		--namespace $(CALICO_NAMESPACE) \
		--version $(CALICO_VERSION) \
		--set apiServer.enabled=false \
		--set goldmane.enabled=false \
		--set whisker.enabled=false \
		--set defaultFelixConfiguration.enabled=false

# --- eBPF setup ---

.PHONY: k3d/cluster/ebpf/setup
k3d/cluster/ebpf/setup:
ifeq ($(GOOS),darwin)
	$(Q)docker exec k3d-$(CLUSTER_NAME)-server-0 mount bpffs /sys/fs/bpf -t bpf && \
	docker exec k3d-$(CLUSTER_NAME)-server-0 mount --make-shared /sys/fs/bpf
endif

# --- Helm deploy options ---

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

# --- Docker network ---
# DOCKER_NETWORK and DOCKER_NETWORK_OPTS are defined in k8s.mk (shared).

DOCKER_NETWORK_LOCK ?= $(BUILD_DIR)/docker_network_$(DOCKER_NETWORK).lock
K3D_PORT_PREFIX_LOCK_DIR ?= $(TMP_DIR_K8S)/k3d-port-prefix.$(DOCKER_NETWORK).lock

# _docker_network_exists: true when the named Docker network is present
_docker_network_exists = docker network inspect $(DOCKER_NETWORK) >/dev/null 2>&1

.PHONY: k3d/docker/network/create
k3d/docker/network/create: | $(dir $(DOCKER_NETWORK_LOCK))
	$(Q)$(call _flock,$(DOCKER_NETWORK_LOCK),\
	  if ! $(_docker_network_exists); then \
	    echo "Creating docker network: $(DOCKER_NETWORK)"; \
	    docker network create --driver bridge $(DOCKER_NETWORK_OPTS) $(DOCKER_NETWORK) >/dev/null; \
	  fi; \
	)

.PHONY: k3d/docker/network/destroy
k3d/docker/network/destroy: | $(dir $(DOCKER_NETWORK_LOCK))
	$(Q)$(call _flock,$(DOCKER_NETWORK_LOCK),\
	  if $(_docker_network_exists); then \
	    endpoints=$$(docker network inspect $(DOCKER_NETWORK) --format '{{len .Containers}}' 2>/dev/null || echo 0); \
	    if [ "$$endpoints" = "0" ]; then \
	      echo "Removing docker network: $(DOCKER_NETWORK)"; \
	      docker network rm $(DOCKER_NETWORK) >/dev/null; \
	    else \
	      echo "Skipping docker network removal: $(DOCKER_NETWORK) still has $$endpoints attached container(s)"; \
	    fi; \
	  fi; \
	)

# Order-only prerequisite: create the lock file directory on demand
$(dir $(DOCKER_NETWORK_LOCK)):
	$(Q)mkdir -p $@

# --- Docker registry credentials ---

DOCKERHUB_PULL_CREDENTIAL ?=

.PHONY: $(K3D_REGISTRY_FILE)
$(K3D_REGISTRY_FILE):
	$(Q)mkdir -p $(dir $@)
ifneq ($(DOCKERHUB_PULL_CREDENTIAL),)
	$(Q)cred='$(DOCKERHUB_PULL_CREDENTIAL)'; \
	$(JQ) -n --arg u "$${cred%%:*}" --arg p "$${cred#*:}" \
	  '{"configs":{"registry-1.docker.io":{"auth":{"username":$$u,"password":$$p}}}}' > $@
else
	$(Q)echo '{"configs":{}}' > $@
endif

.PHONY: k3d/docker/credentials/setup
k3d/docker/credentials/setup: $(K3D_REGISTRY_FILE)

.PHONY: k3d/docker/credentials/cleanup
k3d/docker/credentials/cleanup:
	$(Q)rm -f $(K3D_REGISTRY_FILE)

# --- Cluster lifecycle ---

K3D_CREATE_CLUSTER ?= $(KUMA_DIR)/mk/resources/k3d-create-cluster.sh

.PHONY: k3d/cluster/create
k3d/cluster/create: $(KUBECONFIG_DIR)
	$(Q)$(K3D_CREATE_CLUSTER) \
	  "$(DOCKER_NETWORK)" \
	  "$(K3D_PORT_PREFIX_LOCK_DIR)" \
	  -- \
	  $(K3D) cluster create $(CLUSTER_NAME) $(K3D_CLUSTER_CREATE_OPTS)

.PHONY: k3d/cluster/stop
k3d/cluster/stop:
	$(Q)$(K3D) cluster delete $(CLUSTER_NAME) && rm -f $(K3D_CLUSTER_KUBECONFIG)

# Orchestrated via sequential $(MAKE) calls to guarantee ordering under -j.
.PHONY: k3d/cluster/start
k3d/cluster/start:
	$(Q)$(MAKE) k3d/docker/network/create
	$(Q)$(MAKE) k3d/docker/credentials/setup
	$(Q)$(MAKE) k3d/cluster/create
	$(Q)$(MAKE) k3d/cluster/cni/setup
	$(Q)$(MAKE) k3d/cluster/ebpf/setup
	$(Q)$(MAKE) k3d/cluster/wait
	$(Q)$(MAKE) k3d/cluster/metallb/setup

# --- MetalLB ---

METALLB_RENDER_POOL    ?= $(KUMA_DIR)/mk/resources/render-metallb-pool.sh
METALLB_RESOURCES      ?= $(TMP_DIR_K8S)/metallb-$(CLUSTER_NAME).yaml

.PHONY: $(METALLB_RESOURCES)
$(METALLB_RESOURCES): $(METALLB_RENDER_POOL) k3d/docker/network/create
	$(Q)$(METALLB_RENDER_POOL) $(DOCKER_NETWORK) $(CLUSTER_NUMBER) $@ $(METALLB_NAMESPACE)

.PHONY: k3d/cluster/metallb/setup
k3d/cluster/metallb/setup: $(METALLB_RESOURCES)
	$(Q)$(KUBECTL) apply \
	  --filename $(METALLB_MANIFESTS)

	$(Q)$(KUBECTL) wait \
	  --namespace $(METALLB_NAMESPACE) \
	  --for condition=Ready \
	  --timeout 120s \
	  --all pods

	$(Q)$(KUBECTL) apply \
	  --namespace $(METALLB_NAMESPACE) \
	  --filename $(METALLB_RESOURCES)

# --- Cluster wait ---

.PHONY: k3d/cluster/wait
k3d/cluster/wait: NAMESPACE   ?= kube-system
k3d/cluster/wait: MAX_RETRIES ?= 30
k3d/cluster/wait:
	$(Q)tries=0; \
	until $(KUBECTL) wait pods --all \
	  --context $(KUBECONTEXT) \
	  --namespace $(NAMESPACE) \
	  --timeout 5s \
	  --for condition=Ready 2>/dev/null; do \
	  tries=$$((tries + 1)); \
	  if [ $$tries -ge $(MAX_RETRIES) ]; then \
	    $(MAKE) --silent --no-print-directory k3d/cluster/diagnose NAMESPACE='$(NAMESPACE)'; \
	    exit 1; \
	  fi; \
	  echo "Waiting for pods in $(NAMESPACE) ($$tries/$(MAX_RETRIES))..."; \
	  sleep 1; \
	done

# --- Multi-cluster and teardown ---

.PHONY: k3d/clusters/destroy
k3d/clusters/destroy:
	@for cluster in $$($(K3D) cluster list --output json | $(JQ) -r '.[].name'); do \
		echo "Deleting cluster $$cluster"; \
		$(K3D) cluster delete $$cluster; \
		rm -f $(KUBECONFIG_DIR)/k3d-$$cluster.yaml; \
	done

.PHONY: k3d/destroy
k3d/destroy: k3d/docker/credentials/cleanup k3d/clusters/destroy k3d/docker/network/destroy
	$(Q)echo "k3d environment cleaned up"

.PHONY: k3d/teardown k3d/nuke k3d/kill
k3d/teardown k3d/nuke k3d/kill: k3d/destroy

# --- Deploy: wait helpers ---

.PHONY: k3d/cluster/deploy/wait/cp
k3d/cluster/deploy/wait/cp: \
  k3d/cluster/deploy/wait/Available/deployments/$(PROJECT_NAME)-control-plane \
  k3d/cluster/deploy/wait/Ready/pods/$(PROJECT_NAME)-control-plane \
  k3d/cluster/deploy/wait/mesh

.PHONY: k3d/cluster/deploy/wait/%
k3d/cluster/deploy/wait/%: CONDITION = $(word 1,$(subst /, ,$*))
k3d/cluster/deploy/wait/%: KIND      = $(word 2,$(subst /, ,$*))
k3d/cluster/deploy/wait/%: APP       = $(word 3,$(subst /, ,$*))
k3d/cluster/deploy/wait/%:
	$(Q)echo "Waiting for $(KIND) with app=$(APP) to be $(CONDITION)..."
	$(Q)$(KUBECTL) wait \
		--timeout 60s \
		--namespace $(KUMA_NAMESPACE) \
		--for condition=$(CONDITION) \
		--selector app=$(APP) \
		$(KIND) >/dev/null

.PHONY: k3d/cluster/deploy/wait/mesh
k3d/cluster/deploy/wait/mesh:
	$(Q)echo "Waiting for 'default' mesh to be created..."; \
	until $(KUBECTL) get mesh default >/dev/null 2>&1; do \
	  echo "  Still waiting..."; \
	  sleep 1; \
	done

# --- Deploy: kumactl ---

.PHONY: k3d/cluster/deploy/kumactl/install
k3d/cluster/deploy/kumactl/install:
	$(Q)$(KUMACTL) install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) \
	  | $(KUBECTL) apply --filename - \
	  | grep --invert-match 'unchanged'
	$(Q)printf '\n'

.PHONY: k3d/cluster/deploy/kumactl/clean
k3d/cluster/deploy/kumactl/clean:
	@$(KUMACTL) install --mode $(KUMA_MODE) control-plane $(KUMACTL_INSTALL_CONTROL_PLANE_IMAGES) \
		| $(KUBECTL) delete --filename -

.PHONY: k3d/cluster/deploy/kumactl
k3d/cluster/deploy/kumactl:
	$(Q)$(MAKE) build/kumactl
	$(Q)$(MAKE) k3d/cluster/load
	$(Q)$(MAKE) k3d/cluster/deploy/kumactl/install
	$(Q)$(MAKE) k3d/cluster/deploy/wait/cp

.PHONY: k3d/cluster/restart/kumactl
k3d/cluster/restart/kumactl:
	$(Q)$(MAKE) k3d/cluster/stop
	$(Q)$(MAKE) k3d/cluster/start
	$(Q)$(MAKE) k3d/cluster/deploy/kumactl
	$(Q)$(MAKE) k3d/cluster/deploy/demo

# --- Deploy: Helm ---

.PHONY: k3d/cluster/deploy/helm/clean/release
k3d/cluster/deploy/helm/clean/release:
ifeq (,$(or $(K3D_DEPLOY_HELM_DONT_CLEAN),$(K3D_DEPLOY_HELM_DONT_CLEAN_RELEASE)))
	$(Q)echo "Deleting Helm release '$(PROJECT_NAME)'..."
	$(Q)$(HELM) delete --wait --ignore-not-found --namespace $(KUMA_NAMESPACE) $(PROJECT_NAME) >/dev/null 2>&1

	$(Q)echo "Waiting for control plane pods to terminate..."
	$(Q)until $(KUBECTL) get pods --namespace $(KUMA_NAMESPACE) --selector app=$(PROJECT_NAME)-control-plane >/dev/null 2>&1; do \
	  echo "  Still terminating..."; \
	  sleep 1; \
	done
endif

.PHONY: k3d/cluster/deploy/helm/clean/ns
k3d/cluster/deploy/helm/clean/ns:
ifeq (,$(or $(K3D_DEPLOY_HELM_DONT_CLEAN),$(K3D_DEPLOY_HELM_DONT_CLEAN_NS)))
	$(Q)echo "Deleting namespace '$(KUMA_NAMESPACE)'..."
	$(Q)$(KUBECTL) delete namespace $(KUMA_NAMESPACE) --force --wait --ignore-not-found >/dev/null 2>&1

	$(Q)echo "Waiting for namespace to be fully removed..."
	$(Q)until ! $(KUBECTL) get namespace $(KUMA_NAMESPACE) >/dev/null 2>&1; do \
	  echo "  Still deleting..."; \
	  sleep 1; \
	done
endif

.PHONY: k3d/cluster/deploy/helm/clean
k3d/cluster/deploy/helm/clean:
	$(Q)$(MAKE) k3d/cluster/deploy/helm/clean/release
	$(Q)$(MAKE) k3d/cluster/deploy/helm/clean/ns

.PHONY: k3d/cluster/deploy/helm/upgrade
k3d/cluster/deploy/helm/upgrade:
	$(Q)echo "Upgrading or installing Helm release '$(PROJECT_NAME)'..."
	$(Q)$(HELM) upgrade $(PROJECT_NAME) ./deployments/charts/$(PROJECT_NAME) \
		--install \
		--create-namespace \
		--namespace $(KUMA_NAMESPACE) \
		$(strip $(K3D_HELM_DEPLOY_OPTS))

.PHONY: k3d/cluster/deploy/helm
k3d/cluster/deploy/helm:
	$(Q)$(MAKE) k3d/cluster/load
	$(Q)$(MAKE) k3d/cluster/deploy/helm/clean
	$(Q)$(MAKE) k3d/cluster/deploy/helm/upgrade
	$(Q)$(MAKE) k3d/cluster/deploy/wait/cp

# --- Deploy: demo ---

.PHONY: k3d/cluster/deploy/demo
k3d/cluster/deploy/demo: build/kumactl
	$(Q)$(KUMACTL) install demo | $(KUBECTL) apply -f -
	$(Q)$(KUBECTL) wait --timeout=60s --for=condition=Ready --namespace kuma-demo --all pods

# --- Compatibility targets ---

LEGACY_TARGET_ALIASES := \
	k3d/start=k3d/cluster/start \
	k3d/stop=k3d/cluster/stop \
	k3d/stop/all=k3d/clusters/destroy \
	k3d/wait=k3d/cluster/wait \
	k3d/load/images=k3d/cluster/load/images \
	k3d/load=k3d/cluster/load \
	k3d/network/create=k3d/docker/network/create \
	k3d/network/destroy=k3d/docker/network/destroy \
	k3d/setup-docker-credentials=k3d/docker/credentials/setup \
	k3d/cleanup-docker-credentials=k3d/docker/credentials/cleanup \
	k3d/deploy/kumactl/install=k3d/cluster/deploy/kumactl/install \
	k3d/deploy/kumactl/clean=k3d/cluster/deploy/kumactl/clean \
	k3d/deploy/kumactl=k3d/cluster/deploy/kumactl \
	k3d/restart/kumactl=k3d/cluster/restart/kumactl \
	k3d/deploy/helm/clean/release=k3d/cluster/deploy/helm/clean/release \
	k3d/deploy/helm/clean/ns=k3d/cluster/deploy/helm/clean/ns \
	k3d/deploy/helm/clean=k3d/cluster/deploy/helm/clean \
	k3d/deploy/helm/upgrade=k3d/cluster/deploy/helm/upgrade \
	k3d/deploy/helm=k3d/cluster/deploy/helm \
	k3d/deploy/demo=k3d/cluster/deploy/demo

define _LEGACY_ALIAS
.PHONY: $(1)
$(1):
	@echo "[DEPRECATED] '$(1)' has been renamed to '$(2)'" >&2
	@$$(MAKE) --no-print-directory $(2)
endef

$(foreach m,$(LEGACY_TARGET_ALIASES),\
  $(eval $(call _LEGACY_ALIAS,$(word 1,$(subst =, ,$(m))),$(word 2,$(subst =, ,$(m))))))

k3d/deploy/wait/%:
	@echo "[DEPRECATED] 'k3d/deploy/wait/$*' has been renamed to 'k3d/cluster/deploy/wait/$*'" >&2
	@$(MAKE) --no-print-directory k3d/cluster/deploy/wait/$*

# --- Renamed targets ---

RENAMED_TARGETS := \
	k3d/deploy/kuma=k3d/cluster/deploy/kumactl \
	k3d/restart=k3d/cluster/restart/kumactl \
	k3d/cluster/deploy/kuma=k3d/cluster/deploy/kumactl \
	k3d/cluster/restart=k3d/cluster/restart/kumactl

define _RENAMED
.PHONY: $(1)
$(1):
	$$(error Target was renamed. Use: $(2))
endef

$(foreach m,$(RENAMED_TARGETS),\
  $(eval $(call _RENAMED,$(word 1,$(subst =, ,$(m))),$(word 2,$(subst =, ,$(m))))))
