BUILD_DOCKER_IMAGES_DIR ?= $(BUILD_DIR)/docker-images
KUMA_VERSION ?= master

DOCKER_REGISTRY ?= docker.io/kumahq
DOCKER_USERNAME ?=
DOCKER_API_KEY ?=

KUMACTL_INSTALL_USE_LOCAL_IMAGES?=true
ifeq ($(KUMACTL_INSTALL_USE_LOCAL_IMAGES),true)
	DOCKER_REGISTRY = kumahq
endif

KUMA_CP_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-cp
KUMA_DP_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-dp
KUMACTL_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kumactl
KUMA_INIT_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-init
KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-prometheus-sd

export KUMA_CP_DOCKER_IMAGE ?= $(KUMA_CP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_DP_DOCKER_IMAGE ?= $(KUMA_DP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMACTL_DOCKER_IMAGE ?= $(KUMACTL_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_INIT_DOCKER_IMAGE ?= $(KUMA_INIT_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_PROMETHEUS_SD_DOCKER_IMAGE ?= $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_UNIVERSAL_DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/kuma-universal:$(BUILD_INFO_VERSION)
KUMA_IMAGES ?= $(KUMA_CP_DOCKER_IMAGE) $(KUMA_DP_DOCKER_IMAGE) $(KUMACTL_DOCKER_IMAGE) $(KUMA_INIT_DOCKER_IMAGE) $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) $(KUMA_UNIVERSAL_DOCKER_IMAGE)

IMAGES_TARGETS ?= images/release images/test
DOCKER_SAVE_TARGETS ?= docker/save/release docker/save/test
DOCKER_LOAD_TARGETS ?= docker/load/release docker/load/test

# Always use Docker BuildKit, see
# https://docs.docker.com/develop/develop-images/build_enhancements/
export DOCKER_BUILDKIT := 1

.PHONY: image/kuma-cp
image/kuma-cp: build/kuma-cp/linux-amd64 ## Dev: Rebuild `kuma-cp` Docker image
	docker build -t $(KUMA_CP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-cp .

.PHONY: image/kuma-dp
image/kuma-dp: build/kuma-dp/linux-amd64 build/coredns/linux-amd64 build/artifacts-linux-amd64/envoy/envoy ## Dev: Rebuild `kuma-dp` Docker image
	docker build -t $(KUMA_DP_DOCKER_IMAGE) --build-arg ENVOY_VERSION=${ENVOY_VERSION} -f tools/releases/dockerfiles/Dockerfile.kuma-dp .

.PHONY: image/kumactl
image/kumactl: build/kumactl/linux-amd64 ## Dev: Rebuild `kumactl` Docker image
	docker build -t $(KUMACTL_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kumactl .

.PHONY: image/kuma-init
image/kuma-init: build/kumactl/linux-amd64 ## Dev: Rebuild `kuma-init` Docker image
	docker build -t $(KUMA_INIT_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-init .

.PHONY: image/kuma-prometheus-sd
image/kuma-prometheus-sd: build/kuma-prometheus-sd/linux-amd64 ## Dev: Rebuild `kuma-prometheus-sd` Docker image
	docker build -t $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-prometheus-sd .

.PHONY: image/kuma-universal
image/kuma-universal: build/linux-amd64
	docker build -t kuma-universal --build-arg ENVOY_VERSION=${ENVOY_VERSION}  -f test/dockerfiles/Dockerfile.universal .
	docker tag kuma-universal $(KUMA_UNIVERSAL_DOCKER_IMAGE)

.PHONY: images
images: $(IMAGES_TARGETS) ## Dev: Rebuild release and tesst Docker images

.PHONY: images/release
images/release: image/kuma-cp image/kuma-dp image/kumactl image/kuma-init image/kuma-prometheus-sd ## Dev: Rebuild release Docker images

.PHONY: images/test
images/test: image/kuma-universal ## Dev: Rebuild test Docker images

${BUILD_DOCKER_IMAGES_DIR}:
	mkdir -p ${BUILD_DOCKER_IMAGES_DIR}

.PHONY: docker/save
docker/save: $(DOCKER_SAVE_TARGETS)

.PHONY: docker/save/release
docker/save/release: docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-init docker/save/kuma-prometheus-sd

.PHONY: docker/save/test
docker/save/test: docker/save/kuma-universal

.PHONY: docker/save/kuma-cp
docker/save/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar $(KUMA_CP_DOCKER_IMAGE)

.PHONY: docker/save/kuma-dp
docker/save/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar $(KUMA_DP_DOCKER_IMAGE)

.PHONY: docker/save/kumactl
docker/save/kumactl: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar $(KUMACTL_DOCKER_IMAGE)

.PHONY: docker/save/kuma-init
docker/save/kuma-init: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar $(KUMA_INIT_DOCKER_IMAGE)

.PHONY: docker/save/kuma-prometheus-sd
docker/save/kuma-prometheus-sd: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE)

.PHONY: docker/save/kuma-universal
docker/save/kuma-universal: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-universal.tar $(KUMA_UNIVERSAL_DOCKER_IMAGE)

.PHONY: docker/load
docker/load: $(DOCKER_LOAD_TARGETS)

.PHONY: docker/load/release
docker/load/release: docker/load/kuma-cp docker/load/kuma-dp docker/load/kumactl docker/load/kuma-init docker/load/kuma-prometheus-sd

.PHONY: docker/load/test
docker/load/test: docker/load/kuma-universal

.PHONY: docker/load/kuma-cp
docker/load/kuma-cp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-cp.tar

.PHONY: docker/load/kuma-dp
docker/load/kuma-dp: ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-dp.tar

.PHONY: docker/load/kumactl
docker/load/kumactl: ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kumactl.tar

.PHONY: docker/load/kuma-init
docker/load/kuma-init: ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-init.tar

.PHONY: docker/load/kuma-prometheus-sd
docker/load/kuma-prometheus-sd: ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar

.PHONY: docker/load/kuma-universal
docker/load/kuma-universal: ${BUILD_DOCKER_IMAGES_DIR}/kuma-universal.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-universal.tar

.PHONY: docker/tag/kuma-cp
docker/tag/kuma-cp:
	docker tag $(KUMA_CP_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-cp:$(KUMA_VERSION)

.PHONY: docker/tag/kuma-dp
docker/tag/kuma-dp:
	docker tag $(KUMA_DP_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-dp:$(KUMA_VERSION)

.PHONY: docker/tag/kumactl
docker/tag/kumactl:
	docker tag $(KUMACTL_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kumactl:$(KUMA_VERSION)

.PHONY: docker/tag/kuma-init
docker/tag/kuma-init:
	docker tag $(KUMA_INIT_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-init:$(KUMA_VERSION)

.PHONY: docker/tag/kuma-universal
docker/tag/kuma-universal:
	docker tag $(KUMA_UNIVERSAL_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-universal:$(KUMA_VERSION)

.PHONY: docker/purge
docker/purge: ## Dev: Remove all Docker containers, images, networks and volumes
	for c in `docker ps -q`; do docker kill $$c; done
	docker system prune --all --volumes --force

.PHONY: image/kuma-cp/push
image/kuma-cp/push: image/kuma-cp
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_API_KEY) $(DOCKER_REGISTRY)
	docker tag $(KUMA_CP_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-cp:$(KUMA_VERSION)
	docker push $(DOCKER_REGISTRY)/kuma-cp:$(KUMA_VERSION)
	docker logout $(DOCKER_REGISTRY)

.PHONY: image/kuma-dp/push
image/kuma-dp/push: image/kuma-dp
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_API_KEY) $(DOCKER_REGISTRY)
	docker tag $(KUMA_DP_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-dp:$(KUMA_VERSION)
	docker push $(DOCKER_REGISTRY)/kuma-dp:$(KUMA_VERSION)
	docker logout $(DOCKER_REGISTRY)

.PHONY: image/kumactl/push
image/kumactl/push: image/kumactl
	docker login -u $(DOCKER_USERNAME) -p $(DOCKER_API_KEY) $(DOCKER_REGISTRY)
	docker tag $(KUMACTL_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kumactl:$(KUMA_VERSION)
	docker push $(DOCKER_REGISTRY)/kumactl:$(KUMA_VERSION)
	docker logout $(DOCKER_REGISTRY)

.PHONY: images/push
images/push: image/kuma-cp/push image/kuma-dp/push image/kumactl/push
