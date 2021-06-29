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
KUMA_CNI_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-cni
KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME ?= $(DOCKER_REGISTRY)/kuma-prometheus-sd

export KUMA_CP_DOCKER_IMAGE ?= $(KUMA_CP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_DP_DOCKER_IMAGE ?= $(KUMA_DP_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMACTL_DOCKER_IMAGE ?= $(KUMACTL_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_INIT_DOCKER_IMAGE ?= $(KUMA_INIT_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_CNI_DOCKER_IMAGE ?= $(KUMA_CNI_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_PROMETHEUS_SD_DOCKER_IMAGE ?= $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE_NAME):$(BUILD_INFO_VERSION)
export KUMA_UNIVERSAL_DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/kuma-universal:$(BUILD_INFO_VERSION)

# Always use Docker BuildKit, see
# https://docs.docker.com/develop/develop-images/build_enhancements/
export DOCKER_BUILDKIT := 1

.PHONY: docker/build
docker/build: docker/build/release docker/build/test

.PHONY: docker/build/release
docker/build/release: docker/build/kuma-cp docker/build/kuma-dp docker/build/kumactl docker/build/kuma-init docker/build/kuma-cni docker/build/kuma-prometheus-sd ## Dev: Build all Docker images using existing artifacts from build

.PHONY: docker/build/test
docker/build/test: docker/build/kuma-universal

.PHONY: docker/build/kuma-cp
docker/build/kuma-cp: build/artifacts-linux-amd64/kuma-cp/kuma-cp ## Dev: Build `kuma-cp` Docker image using existing artifact
	docker build -t $(KUMA_CP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-cp .

.PHONY: docker/build/kuma-dp
docker/build/kuma-dp: build/artifacts-linux-amd64/kuma-dp/kuma-dp build/artifacts-linux-amd64/coredns/coredns ## Dev: Build `kuma-dp` Docker image using existing artifact
	docker build -t $(KUMA_DP_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-dp .

.PHONY: docker/build/kumactl
docker/build/kumactl: build/artifacts-linux-amd64/kumactl/kumactl ## Dev: Build `kumactl` Docker image using existing artifact
	docker build -t $(KUMACTL_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kumactl .

.PHONY: docker/build/kuma-init
docker/build/kuma-init: ## Dev: Build `kuma-init` Docker image using existing artifact
	docker build -t $(KUMA_INIT_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-init .

.PHONY: docker/build/kuma-cni
docker/build/kuma-cni:
	docker build -t $(KUMA_CNI_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-cni .

.PHONY: docker/build/kuma-prometheus-sd
docker/build/kuma-prometheus-sd: build/artifacts-linux-amd64/kuma-prometheus-sd/kuma-prometheus-sd ## Dev: Build `kuma-prometheus-sd` Docker image using existing artifact
	docker build -t $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE) -f tools/releases/dockerfiles/Dockerfile.kuma-prometheus-sd .

.PHONY: docker/build/kuma-universal
docker/build/kuma-universal: ## Dev: Build `kuma-universal` Docker image using existing artifact
docker/build/kuma-universal: build/artifacts-linux-amd64/kuma-cp/kuma-cp build/artifacts-linux-amd64/kuma-dp/kuma-dp build/artifacts-linux-amd64/kumactl/kumactl build/artifacts-linux-amd64/test-server/test-server
	docker build -t kuma-universal -f test/dockerfiles/Dockerfile.universal .
	docker tag kuma-universal $(KUMA_UNIVERSAL_DOCKER_IMAGE)

.PHONY: image/kuma-cp
image/kuma-cp: build/kuma-cp/linux-amd64 docker/build/kuma-cp ## Dev: Rebuild `kuma-cp` Docker image

.PHONY: image/kuma-dp
image/kuma-dp: build/kuma-dp/linux-amd64 build/coredns/linux-amd64 docker/build/kuma-dp ## Dev: Rebuild `kuma-dp` Docker image

.PHONY: image/kumactl
image/kumactl: build/kumactl/linux-amd64 docker/build/kumactl ## Dev: Rebuild `kumactl` Docker image

.PHONY: image/kuma-init
image/kuma-init: docker/build/kuma-init ## Dev: Rebuild `kuma-init` Docker image

.PHONY: image/kuma-cni
image/kuma-cni: build/kuma-cni/linux-amd64 build/kuma-cni-install/linux-amd64 docker/build/kuma-cni

.PHONY: image/kuma-prometheus-sd
image/kuma-prometheus-sd: build/kuma-prometheus-sd/linux-amd64 docker/build/kuma-prometheus-sd ## Dev: Rebuild `kuma-prometheus-sd` Docker image

.PHONY: image/kuma-universal
image/kuma-universal: build/linux-amd64 docker/build/kuma-universal

.PHONY: images
images: images/release images/test ## Dev: Rebuild release and tesst Docker images

.PHONY: images/release
images/release: image/kuma-cp image/kuma-dp image/kumactl image/kuma-init image/kuma-cni image/kuma-prometheus-sd ## Dev: Rebuild release Docker images

.PHONY: images/test
images/test: image/kuma-universal ## Dev: Rebuild test Docker images

${BUILD_DOCKER_IMAGES_DIR}:
	mkdir -p ${BUILD_DOCKER_IMAGES_DIR}

.PHONY: docker/save
docker/save: docker/save/release docker/save/test

.PHONY: docker/save/release
docker/save/release: docker/save/kuma-cp docker/save/kuma-dp docker/save/kumactl docker/save/kuma-init docker/save/kuma-cni docker/save/kuma-prometheus-sd

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

.PHONY: docker/save/kuma-cni
docker/save/kuma-cni: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-cni.tar $(KUMA_CNI_DOCKER_IMAGE)

.PHONY: docker/save/kuma-prometheus-sd
docker/save/kuma-prometheus-sd: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-prometheus-sd.tar $(KUMA_PROMETHEUS_SD_DOCKER_IMAGE)

.PHONY: docker/save/kuma-universal
docker/save/kuma-universal: ${BUILD_DOCKER_IMAGES_DIR}
	docker save --output ${BUILD_DOCKER_IMAGES_DIR}/kuma-universal.tar $(KUMA_UNIVERSAL_DOCKER_IMAGE)

.PHONY: docker/load
docker/load: docker/load/release docker/load/test

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

.PHONY: docker/load/kuma-cni
docker/load/kuma-cni: ${BUILD_DOCKER_IMAGES_DIR}/kuma-cni.tar
	docker load --input ${BUILD_DOCKER_IMAGES_DIR}/kuma-cni.tar

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

.PHONY: docker/tag/kuma-cni
docker/tag/kuma-cni:
	docker tag $(KUMA_CNI_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-cni:$(KUMA_VERSION)

.PHONY: docker/tag/kuma-universal
docker/tag/kuma-universal:
	docker tag $(KUMA_UNIVERSAL_DOCKER_IMAGE) $(DOCKER_REGISTRY)/kuma-universal:$(KUMA_VERSION)

.PHONY: docker/docker
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
