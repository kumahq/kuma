BUILD_DOCKER_IMAGES_DIR ?= $(BUILD_DIR)/docker-images-${GOARCH}
KUMA_VERSION ?= master

DOCKER_SERVER ?= docker.io
DOCKER_REGISTRY ?= $(DOCKER_SERVER)/kumahq
DOCKER_DEV_IMAGE_SUFFIX ?=
DOCKER_USERNAME ?=
DOCKER_API_KEY ?=

ifdef PRE_RELEASE_REPOSITORY_SUFFIX
ifneq ($(strip $(PRE_RELEASE_REPOSITORY_SUFFIX)),)
DOCKER_PRE_RELEASE_IMAGE_SUFFIX = -$(PRE_RELEASE_REPOSITORY_SUFFIX)
endif
endif

define build_image
$(addsuffix :$(BUILD_INFO_VERSION)$(if $(2),-$(2)),$(addprefix $(DOCKER_REGISTRY)/,$(1)$(DOCKER_DEV_IMAGE_SUFFIX)))
endef

IMAGES_RELEASE += kuma-cp kuma-dp kumactl kuma-init kuma-cni
IMAGES_TEST += kuma-universal
KUMA_IMAGES = $(call build_image,$(IMAGES_RELEASE) $(IMAGES_TEST))

# Always use Docker BuildKit, see
# https://docs.docker.com/develop/develop-images/build_enhancements/
export DOCKER_BUILDKIT := 1

# add targets to build images for each arch
# $(1) - GOARCH to build for

define IMAGE_TARGETS_BY_ARCH
.PHONY: image/static/$(1)
image/static/$(1): ## Dev: Rebuild `kuma-static` Docker image
	docker build -t kumahq/static-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/static.Dockerfile .

.PHONY: image/base/$(1)
image/base/$(1): ## Dev: Rebuild `kuma-base` Docker image
	docker build -t kumahq/base-nossl-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/base.Dockerfile .

.PHONY: image/base-root/$(1)
image/base-root/$(1): ## Dev: Rebuild `kuma-base-root` Docker image
	docker build -t kumahq/base-root-debian11:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/base-root.Dockerfile .

.PHONY: image/envoy/$(1)
image/envoy/$(1): build/artifacts-linux-$(1)/envoy ## Dev: Rebuild `envoy` Docker image
	docker build -t kumahq/envoy:no-push-$(1) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/envoy.Dockerfile .

.PHONY: image/kuma-cp/$(1)
image/kuma-cp/$(1): image/static/$(1) build/artifacts-linux-$(1)/kuma-cp ## Dev: Rebuild `kuma-cp` Docker image
	docker build -t $$(call build_image,kuma-cp,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/kuma-cp.Dockerfile .

.PHONY: image/kuma-dp/$(1)
image/kuma-dp/$(1): image/base/$(1) image/envoy/$(1) build/artifacts-linux-$(1)/kuma-dp build/artifacts-linux-$(1)/coredns ## Dev: Rebuild `kuma-dp` Docker image
	docker build -t $$(call build_image,kuma-dp,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/kuma-dp.Dockerfile .

.PHONY: image/kumactl/$(1)
image/kumactl/$(1): image/base/$(1) build/artifacts-linux-$(1)/kumactl ## Dev: Rebuild `kumactl` Docker image
	docker build -t $$(call build_image,kumactl,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/kumactl.Dockerfile .

.PHONY: image/kuma-init/$(1)
image/kuma-init/$(1): build/artifacts-linux-$(1)/kumactl ## Dev: Rebuild `kuma-init` Docker image
	docker build -t $$(call build_image,kuma-init,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/kuma-init.Dockerfile .

.PHONY: image/kuma-cni/$(1)
image/kuma-cni/$(1): image/base-root/$(1) build/artifacts-linux-$(1)/kuma-cni build/artifacts-linux-$(1)/install-cni
	docker build -t $$(call build_image,kuma-cni,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(TOOLS_DIR)/releases/dockerfiles/kuma-cni.Dockerfile .

.PHONY: image/kuma-universal/$(1)
image/kuma-universal/$(1): image/envoy/$(1) build/artifacts-linux-$(1)/kuma-cp build/artifacts-linux-$(1)/kuma-dp build/artifacts-linux-$(1)/kumactl build/artifacts-linux-$(1)/kumactl build/artifacts-linux-$(1)/test-server build/artifacts-linux-$(1)/coredns
	docker build -t $$(call build_image,kuma-universal,$(1)) --build-arg ARCH=$(1) --platform=linux/$(1) -f $(KUMA_DIR)/test/dockerfiles/universal.Dockerfile .
endef
$(foreach goarch,$(SUPPORTED_GOARCHES),$(eval $(call IMAGE_TARGETS_BY_ARCH,$(goarch))))

# add targets to generate docker/{save,load,tag,push} for each supported ARCH
# add targets to build images for each arch
# $(1) - Image Name to build for
# $(2) - GOARCH to build for
# (TODO): Support image platform in output file names
define DOCKER_TARGETS_BY_ARCH
.PHONY: docker/save/$(1)/$(2)
docker/save/$(1)/$(2):
	@mkdir -p build/docker
	docker save --output build/docker/$(1)-$(2).tar $$(call build_image,$(1),$(2))

.PHONY: docker/$(1)/$(2)
docker/load/$(1)/$(2):
	@docker load --quiet --input build/docker/$(1)-$(2).tar

# we only tag the image that has the same arch than the HOST (tag is meant to use the image just after so having the same arch makes sense)
.PHONY: docker/tag/$(1)/$(2)
docker/tag/$(1)/$(2):
	$$(if $$(findstring $(GOARCH),$(2)),docker tag $$(call build_image,$(1),$(2)) $$(call build_image,$(1)),# Not tagging $(1) as $(2) is not host arch)

.PHONY: docker/push/$(1)/$(2)
docker/push/$(1)/$(2):
	$$(call GATE_PUSH,docker push $$(call build_image,$(1),$(2)))
endef
$(foreach goarch, $(SUPPORTED_GOARCHES),$(foreach image, $(IMAGES_RELEASE) $(IMAGES_TEST),$(eval $(call DOCKER_TARGETS_BY_ARCH,$(image),$(goarch)))))

# add targets to generate docker/{save,load,tag,push} for all supported ARCH
# $(1) - image name
define DOCKER_TARGETS_BY_IMAGE
.PHONY: docker/save/$(1)
docker/save/$(1): $(foreach arch,$(ENABLED_GOARCHES),docker/save/$(1)/$(arch)) ## Dev: Save `$(1)` Docker image

.PHONY: docker/load/$(1)
docker/load/$(1): $(foreach arch,$(ENABLED_GOARCHES),docker/load/$(1)/$(arch)) ## Dev: Load `$(1)` Docker image

.PHONY: docker/tag/$(1)
docker/tag/$(1): $(foreach arch,$(ENABLED_GOARCHES),docker/tag/$(1)/$(arch)) ## Dev: Tag `$(1)` Docker image

.PHONY: docker/push/$(1)
docker/push/$(1): $(foreach arch,$(ENABLED_GOARCHES),docker/push/$(1)/$(arch)) ## Dev: Push `$(1)` Docker image

.PHONY: docker/manifest/$(1)
docker/manifest/$(1): ## Dev: Create and push a manifest that groups all arches for `$(1)`
	$(call GATE_PUSH,docker manifest create $(call build_image,$(1)) $(patsubst %,--amend $(call build_image,$(1),%),$(ENABLED_GOARCHES)))
	$(call GATE_PUSH,docker manifest push $(call build_image,$(1)))

.PHONY: images/$(1)
images/$1: $(addprefix image/$(1)/,$(ENABLED_GOARCHES)) ## Dev: Rebuild `$(1)` Docker image for all arches
endef
$(foreach image, $(IMAGES_RELEASE) $(IMAGES_TEST),$(eval $(call DOCKER_TARGETS_BY_IMAGE,$(image))))

# add targets like `docker/save` with dependencies all `ENABLED_GOARCHES`
.PHONY: docker/save
docker/save: $(patsubst %,docker/save/%,$(IMAGES_RELEASE))
.PHONY: docker/load
docker/load: $(patsubst %,docker/load/%,$(IMAGES_RELEASE))
.PHONY: docker/tag
docker/tag: docker/tag/test docker/tag/release ## Tag local arch containers with the version with the arch (this is mostly to use non multi-arch images as if they were released images in e2e tests)
.PHONY: docker/tag/release
docker/tag/release: $(patsubst %,docker/tag/%,$(IMAGES_RELEASE))
.PHONY: docker/tag/test
docker/tag/test: $(patsubst %,docker/tag/%,$(IMAGES_TEST))
.PHONY: docker/push
docker/push: $(patsubst %,docker/push/%,$(IMAGES_RELEASE)) ## Publish all docker images with arch specific tags
.PHONY: docker/manifest
docker/manifest: $(patsubst %,docker/manifest/%,$(IMAGES_RELEASE)) ## Publish all manifests (images need to be pushed already
.PHONY: images
images: images/release images/test ## Dev: Rebuild release and test Docker images
.PHONY: images/release
images/release: $(addprefix images/,$(IMAGES_RELEASE)) ## Dev: Rebuild release Docker images
.PHONY: images/test
images/test: $(addprefix images/,$(IMAGES_TEST)) ## Dev: Rebuild test Docker images

.PHONY: docker/info/registry
docker/info/registry: ## Output the Docker registry
	@echo $(DOCKER_REGISTRY)

# The awk command is ok because we're passing a list of container image names which won't contain ' ' or '"'
# This outputs something like: ["docker.io/kumahq/kuma-cp:0.0.0-preview.vlocal-build","docker.io/kumahq/kuma-dp:0.0.0-preview.vlocal-build","docker.io/kumahq/kumactl:0.0.0-preview.vlocal-build","docker.io/kumahq/kuma-init:0.0.0-preview.vlocal-build","docker.io/kumahq/kuma-cni:0.0.0-preview.vlocal-build"]
.PHONY: manifests/json/release
manifests/json/release: ## output all release manifests in a json array
	@echo $(call build_image,$(IMAGES_RELEASE)) | awk 'BEGIN{FS=" "; printf("[")}{for(i=1;i<=NF;i++)  printf("\"%s\"%s", $$i, i!=NF ? "," : "")} END{printf("]")}'

.PHONY: images/info/release/json
images/info/release/json:
	@echo $(IMAGES_RELEASE) | awk 'BEGIN{FS=" "; printf("[")}{for(i=1;i<=NF;i++)  printf("\"%s\"%s", $$i, i!=NF ? "," : "")} END{printf("]")}'

.PHONY: docker/purge
docker/purge: ## Dev: Remove all Docker containers, images, networks and volumes
	for c in `docker ps -q`; do docker kill $$c; done
	docker system prune --all --volumes --force

.PHONY: docker/login
docker/login:
	$(call GATE_PUSH,docker login -u $(DOCKER_USERNAME) -p $(DOCKER_API_KEY) $(DOCKER_SERVER))

.PHONY: docker/logout
docker/logout:
	$(call GATE_PUSH,docker logout $(DOCKER_SERVER))
