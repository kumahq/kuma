BUILD_DOCKER_IMAGES_DIR ?= $(BUILD_DIR)/docker-images-${GOARCH}
KUMA_VERSION ?= master

DOCKER_REPO ?= docker.io
DOCKER_REGISTRY ?= kumahq
DOCKER_USERNAME ?=
DOCKER_API_KEY ?=

define build_image
$(addsuffix :$(BUILD_INFO_VERSION)$(if $(2),-$(2)),$(addprefix $(DOCKER_REPO)/$(DOCKER_REGISTRY)/,$(1)))
endef

IMAGES_RELEASE += kuma-cp kuma-dp kumactl kuma-init kuma-cni
IMAGES_TEST = kuma-universal
KUMA_IMAGES = $(call build_image,$(IMAGES_RELEASE) $(IMAGES_TEST))

.PHONY: images/show
images/show: ## output all images that are built with the current configuration
	@echo $(KUMA_IMAGES)

# Always use Docker BuildKit, see
# https://docs.docker.com/develop/develop-images/build_enhancements/
export DOCKER_BUILDKIT := 1

ENVOY_ARTIFACT_EXT ?= opt
# add targets to build images for each arch
# $(1) - GOOS to build for
define IMAGE_TARGETS_BY_ARCH
.PHONY: image/static/$(1)
image/static/$(1): ## Dev: Rebuild `kuma-static` Docker image
	docker build -t kumahq/static-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.static .

.PHONY: image/base/$(1)
image/base/$(1): ## Dev: Rebuild `kuma-base` Docker image
	docker build -t kumahq/base-nossl-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.base .

.PHONY: image/base-root/$(1)
image/base-root/$(1): ## Dev: Rebuild `kuma-base-root` Docker image
	docker build -t kumahq/base-root-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.base-root .

.PHONY: image/envoy/$(1)
image/envoy/$(1): build/artifacts-linux-$(1)/envoy/$(ENVOY_VERSION)-alpine-$(ENVOY_ARTIFACT_EXT)/envoy ## Dev: Rebuild `envoy` Docker image
	docker build -t kumahq/envoy:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) --build-arg ENVOY_VERSION=$(ENVOY_VERSION)-alpine-$(ENVOY_ARTIFACT_EXT) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.envoy .

.PHONY: image/kuma-cp/$(1)
image/kuma-cp/$(1): image/static/$(1) build/artifacts-linux-$(1)/kuma-cp ## Dev: Rebuild `kuma-cp` Docker image
	docker build -t $$(call build_image,kuma-cp,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.kuma-cp .

.PHONY: image/kuma-dp/$(1)
image/kuma-dp/$(1): image/base/$(1) image/envoy/$(1) build/artifacts-linux-$(1)/kuma-dp build/artifacts-linux-$(1)/coredns ## Dev: Rebuild `kuma-dp` Docker image
	docker build -t $$(call build_image,kuma-dp,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.kuma-dp .

.PHONY: image/kumactl/$(1)
image/kumactl/$(1): image/base/$(1) build/artifacts-linux-$(1)/kumactl ## Dev: Rebuild `kumactl` Docker image
	docker build -t $$(call build_image,kumactl,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.kumactl .

.PHONY: image/kuma-init/$(1)
image/kuma-init/$(1): build/artifacts-linux-$(1)/kumactl ## Dev: Rebuild `kuma-init` Docker image
	docker build -t $$(call build_image,kuma-init,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.kuma-init .

.PHONY: image/kuma-cni/$(1)
image/kuma-cni/$(1): image/base-root/$(1) build/artifacts-linux-$(1)/kuma-cni build/artifacts-linux-$(1)/install-cni
	docker build -t $$(call build_image,kuma-cni,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/Dockerfile.kuma-cni .

.PHONY: image/kuma-universal/$(1)
image/kuma-universal/$(1): image/envoy/$(1) build/artifacts-linux-$(1)/kuma-cp build/artifacts-linux-$(1)/kuma-dp build/artifacts-linux-$(1)/kumactl build/artifacts-linux-$(1)/kumactl build/artifacts-linux-$(1)/test-server
	docker build -t $$(call build_image,kuma-universal,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(KUMA_DIR)/test/dockerfiles/Dockerfile.universal .
endef
$(foreach goarch,$(SUPPORTED_GOARCHES),$(eval $(call IMAGE_TARGETS_BY_ARCH,$(goarch))))

# add targets to generate docker/{save,load,tag,push} for each supported ARCH
# add targets to build images for each arch
# $(1) - GOOS to build for
# $(2) - GOARCH to build for
define DOCKER_TARGETS_BY_ARCH
.PHONY: docker/$(1)/$(2)/save
docker/$(1)/$(2)/save:
	@mkdir -p build/docker
	docker save --output build/docker/$(1)-$(2).tar $$(call build_image,$(1),$(2))

.PHONY: docker/$(1)/$(2)/load
docker/$(1)/$(2)/load:
	@docker load --quiet --input build/docker/$(1)-$(2).tar

# we only tag the image that has the same arch than the HOST (tag is meant to use the image just after so having the same arch makes sense)
.PHONY: docker/$(1)/$(2)/tag
docker/$(1)/$(2)/tag:
	$$(if $$(findstring $(GOARCH),$(2)),docker tag $$(call build_image,$(1),$(2)) $$(call build_image,$(1)),# Not tagging $(1) as $(2) is not host arch)

.PHONY: docker/$(1)/$(2)/push
docker/$(1)/$(2)/push:
	$$(call GATE_PUSH,docker push $$(call build_image,$(1),$(2)))
endef
$(foreach goarch, $(SUPPORTED_GOARCHES),$(foreach image, $(IMAGES_RELEASE) $(IMAGES_TEST),$(eval $(call DOCKER_TARGETS_BY_ARCH,$(image),$(goarch)))))

# create and push a manifest for each
docker/%/manifest:
	$(call GATE_PUSH,docker manifest create $(call build_image,$*) $(patsubst %,--amend $(call build_image,$*,%),$(ENABLED_GOARCHES)))
	$(call GATE_PUSH,docker manifest push $(call build_image,$*))

# add targets like `docker/save` with dependencies all `ENABLED_GOARCHES`
ALL_RELEASE_WITH_ARCH=$(foreach arch,$(ENABLED_GOARCHES),$(patsubst %,%/$(arch),$(IMAGES_RELEASE)))
ALL_TEST_WITH_ARCH=$(foreach arch,$(ENABLED_GOARCHES),$(patsubst %,%/$(arch),$(IMAGES_TEST)))
.PHONY: docker/save
docker/save: $(patsubst %,docker/%/save,$(ALL_RELEASE_WITH_ARCH) $(ALL_TEST_WITH_ARCH))
.PHONY: docker/load
docker/load: $(patsubst %,docker/%/load,$(ALL_RELEASE_WITH_ARCH) $(ALL_TEST_WITH_ARCH))
.PHONY: docker/tag
docker/tag: docker/tag/test docker/tag/release
.PHONY: docker/tag/release
docker/tag/release: $(patsubst %,docker/%/tag,$(ALL_RELEASE_WITH_ARCH))
.PHONY: docker/tag/test
docker/tag/test: $(patsubst %,docker/%/tag,$(ALL_TEST_WITH_ARCH))
.PHONY: docker/push
docker/push: $(patsubst %,docker/%/push,$(ALL_RELEASE_WITH_ARCH))
.PHONY: docker/manifest
docker/manifest: $(patsubst %,docker/%/manifest,$(IMAGES_RELEASE))
.PHONY: images
images: images/release images/test ## Dev: Rebuild release and test Docker images
.PHONY: images/release
images/release: $(addprefix image/,$(ALL_RELEASE_WITH_ARCH)) ## Dev: Rebuild release Docker images
.PHONY: images/test
images/test: $(addprefix image/,$(ALL_TEST_WITH_ARCH)) ## Dev: Rebuild test Docker images

.PHONY: docker/purge
docker/purge: ## Dev: Remove all Docker containers, images, networks and volumes
	for c in `docker ps -q`; do docker kill $$c; done
	docker system prune --all --volumes --force

.PHONY: docker/login
docker/login:
	$(call GATE_PUSH,docker login -u $(DOCKER_USERNAME) -p $(DOCKER_API_KEY) $(DOCKER_REPO))

.PHONY: docker/logout
docker/logout:
	$(call GATE_PUSH,docker logout $(DOCKER_REPO))
