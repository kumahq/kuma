.PHONY: clean \
		local/package local/image \
		all/package all/image \
		download/binary download/all/binaries \
		build/image/konvoy-build-fpm \
		build/package build/image \
		upload/package upload/image \
		print/supported/target_os print/supported/base_images \
		foreach/build/package foreach/build/image

# Supported base images
ALL_BASE_IMAGES := ubuntu:14.04 ubuntu:16.04 ubuntu:18.04 \
					debian:7 debian:8 debian:9 \
					rhel:6 rhel:7 \
					centos:6 centos:7 \
					amazonlinux:2 \
					alpine:3.7 alpine:3.8 alpine:3.9
ALL_TARGET_OS := ubuntu:14.04 centos:7
# Target distribution
BASE_IMAGE ?= ubuntu:16.04
BASE_IMAGE_NAME ?= $(shell echo '$(BASE_IMAGE)' | awk -F: '{print $$1}')
BASE_IMAGE_TAG ?= $(shell echo '$(BASE_IMAGE)' | awk -F: '{print $$2}')
# Conditional variables
PARENT_IMAGE=$(BASE_IMAGE)
ifeq ($(BASE_IMAGE_NAME),debian)
  TARGET_OS = debian
  DOCKERFILE_TEMPLATE = deb
endif
ifeq ($(BASE_IMAGE_NAME),ubuntu)
  TARGET_OS = debian
  DOCKERFILE_TEMPLATE = deb
endif
ifeq ($(BASE_IMAGE_NAME),centos)
  TARGET_OS = centos
  DOCKERFILE_TEMPLATE = rpm
endif
ifeq ($(BASE_IMAGE_NAME),rhel)
  TARGET_OS = centos
  DOCKERFILE_TEMPLATE = rpm
  PARENT_IMAGE = registry.access.redhat.com/rhel${BASE_IMAGE_TAG}
endif
ifeq ($(BASE_IMAGE_NAME),amazonlinux)
  TARGET_OS = centos
  DOCKERFILE_TEMPLATE = rpm
endif
ifeq ($(BASE_IMAGE_NAME),alpine)
  TARGET_OS = debian
  DOCKERFILE_TEMPLATE = apk
  PARENT_IMAGE = frolvlad/alpine-glibc:alpine-$(BASE_IMAGE_TAG)
endif
# CI coordinates
COMMIT ?= 00000000000000000000
PREBUILT_TAR ?= konvoy-$(TARGET_OS).tar.gz
S3_CI_BUCKET ?= konvoy-ci
S3_KONVOY_BINARIES ?= $(COMMIT)/binaries
S3_KONVOY_BINARY ?= $(S3_KONVOY_BINARIES)/$(PREBUILT_TAR)
# Package info
PACKAGE_NAME ?= konvoy
# Debian version must start from a number
PACKAGE_VERSION ?= 0.x-$(COMMIT)
PACKAGE_ARCH := x86_64
PACKAGE_DESCRIPTION ?= "Konvoy distribution."
PACKAGE_URL ?= "https://konvoy.io"
PACKAGE_VENDOR ?= "Konvoy Project"
PACKAGE_MAINTAINER ?= "Konvoy Developers"
PACKAGE_LICENSE ?= "Apache 2.0"
DOCKER_IMAGE := konvoy/konvoy-$(BASE_IMAGE_NAME)$(BASE_IMAGE_TAG):$(PACKAGE_VERSION)
S3_KONVOY_PACKAGE := $(COMMIT)/packages/$(BASE_IMAGE)
S3_KONVOY_IMAGE := $(COMMIT)/images/$(BASE_IMAGE)
# Docker images (until open sourced, docker images will be stored in S3)
S3_DOCKER_IMAGES := images
DOCKER_IMAGE_BUILD_FPM := konvoy/konvoy-build-fpm:latest
# Temporary files
BUILD_DIR ?= $(PWD)/build
OUTPUT_DIR ?= $(BUILD_DIR)/output
PACKAGES_DIR ?= $(OUTPUT_DIR)/packages
IMAGES_DIR ?= $(OUTPUT_DIR)/images
# Travis CI build fails when a docker container attempts to write into a mounted folder with ':' in its name
PACKAGE_DIR ?= $(PACKAGES_DIR)/$(BASE_IMAGE_NAME)/$(BASE_IMAGE_TAG)
IMAGE_DIR ?= $(IMAGES_DIR)/$(BASE_IMAGE_NAME)/$(BASE_IMAGE_TAG)
DOWNLOADS_DIR ?= $(BUILD_DIR)/downloads
STAGING_DIR ?= $(BUILD_DIR)/staging/konvoy-$(TARGET_OS)

clean:
	rm -rf $(BUILD_DIR)

download/binary:
	@echo "Downloading pre-built binary into '$(DOWNLOADS_DIR)' ..."
	aws s3 cp s3://$(S3_CI_BUCKET)/$(S3_KONVOY_BINARY) $(DOWNLOADS_DIR)/
	if [ -d $(STAGING_DIR) ]; then rm -rf $(STAGING_DIR) ; fi
	mkdir -p $(STAGING_DIR)/usr
	@echo "Extracting pre-built binary into '$(STAGING_DIR)' ..."
	tar -xzf $(DOWNLOADS_DIR)/$(PREBUILT_TAR) -C $(STAGING_DIR)/usr ./bin/

require/binary:
	[ -d $(STAGING_DIR) ]

build/image/konvoy-build-fpm:
	docker inspect --type=image $(DOCKER_IMAGE_BUILD_FPM) >/dev/null || \
     ( echo "Downloading Docker image '$(DOCKER_IMAGE_BUILD_FPM)' ..." && \
	 aws s3 cp s3://$(S3_CI_BUCKET)/$(S3_DOCKER_IMAGES)/$(DOCKER_IMAGE_BUILD_FPM) $(DOWNLOADS_DIR)/images/$(DOCKER_IMAGE_BUILD_FPM) && docker load -i $(DOWNLOADS_DIR)/images/$(DOCKER_IMAGE_BUILD_FPM) ) || \
     ( echo "Building Docker image '$(DOCKER_IMAGE_BUILD_FPM)' ..." && \
	 docker build --file images/fpm/Dockerfile.fpm -t $(DOCKER_IMAGE_BUILD_FPM) images/fpm )

require/image/konvoy-build-fpm:
	docker inspect --type=image $(DOCKER_IMAGE_BUILD_FPM) >/dev/null

copy/etc:
	[ ! -d $(STAGING_DIR)/etc/konvoy ] && mkdir -p $(STAGING_DIR)/etc/konvoy || true
	cp configs/konvoy-demo.yaml $(STAGING_DIR)/etc/konvoy/

build/package: require/binary require/image/konvoy-build-fpm copy/etc
	[ ! -d $(PACKAGE_DIR) ] && mkdir -p $(PACKAGE_DIR) || true
	docker run -t --rm \
	 -v $(STAGING_DIR):/source \
	 -v $(PWD)/images/fpm/entrypoint.sh:/tmp/entrypoint.sh:ro \
	 -v $(PACKAGE_DIR):/output \
	 -e SOURCE_DIR=/source \
	 -e OUTPUT_DIR=/output \
	 -e PACKAGE_NAME=$(PACKAGE_NAME) \
	 -e PACKAGE_VERSION=$(PACKAGE_VERSION) \
	 -e PACKAGE_ARCH=$(PACKAGE_ARCH) \
	 -e PACKAGE_DESCRIPTION=$(PACKAGE_DESCRIPTION) \
	 -e PACKAGE_URL=$(PACKAGE_URL) \
	 -e PACKAGE_VENDOR=$(PACKAGE_VENDOR) \
	 -e PACKAGE_MAINTAINER=$(PACKAGE_MAINTAINER) \
	 -e PACKAGE_LICENSE=$(PACKAGE_LICENSE) \
	 -e BASE_IMAGE_NAME=$(BASE_IMAGE_NAME) \
	 -e BASE_IMAGE_TAG=$(BASE_IMAGE_TAG) \
	 --workdir /tmp \
	 --user $(shell id -u):$(shell id -g) \
	 --entrypoint /tmp/entrypoint.sh \
	 $(DOCKER_IMAGE_BUILD_FPM)

require/package:
	[ $(shell find $(PACKAGE_DIR) -maxdepth 1 -type f | wc -l) -eq 1 ]

upload/package: require/package
	find $(PACKAGE_DIR) -maxdepth 1 -type f -exec sh -c 'aws s3 cp {} s3://$(S3_CI_BUCKET)/$(S3_KONVOY_PACKAGE)/$$(basename {})' \;

local/package: download/binary build/image/konvoy-build-fpm build/package

all/package: local/package upload/package

build/image: require/binary copy/etc
	[ ! -d $(STAGING_DIR)/docker ] && mkdir -p $(STAGING_DIR)/docker || true
	cp images/konvoy/entrypoint.sh $(STAGING_DIR)/docker/
	docker build \
	 --file images/konvoy/Dockerfile.$(DOCKERFILE_TEMPLATE) \
	 --build-arg PARENT_IMAGE=$(PARENT_IMAGE) \
	 -t $(DOCKER_IMAGE) \
	 $(STAGING_DIR)
	# sanity check
	docker run $(DOCKER_IMAGE) --version

require/image:
	docker inspect --type=image $(DOCKER_IMAGE) >/dev/null

upload/image: require/image
	[ ! -d $(IMAGE_DIR)/$(shell dirname $(DOCKER_IMAGE)) ] && mkdir -p $(IMAGE_DIR)/$(shell dirname $(DOCKER_IMAGE)) || true
	docker save $(DOCKER_IMAGE) -o $(IMAGE_DIR)/$(DOCKER_IMAGE)
	aws s3 cp $(IMAGE_DIR)/$(DOCKER_IMAGE) s3://$(S3_CI_BUCKET)/$(S3_KONVOY_IMAGE)/$(DOCKER_IMAGE)

local/image: download/binary build/image

all/image: local/image upload/image

# Targets for local development

print/supported/target_os:
	@echo $(ALL_TARGET_OS)

print/supported/base_images:
	@echo $(ALL_BASE_IMAGES)

download/all/binaries:
	for IMAGE in $(ALL_TARGET_OS) ; do \
		make download/binary BASE_IMAGE=$${IMAGE} ; \
	done

require/all/binaries:
	for IMAGE in $(ALL_TARGET_OS) ; do \
		make require/binary BASE_IMAGE=$${IMAGE} ; \
	done

foreach/build/package: require/all/binaries require/image/konvoy-build-fpm
	for IMAGE in $(ALL_BASE_IMAGES) ; do \
		make build/package BASE_IMAGE=$${IMAGE} ; \
	done
	@echo "All built packages:"
	find $(PACKAGES_DIR) -type f -exec file {} \;

foreach/build/image: require/all/binaries
	for IMAGE in $(ALL_BASE_IMAGES) ; do \
		make local/image BASE_IMAGE=$${IMAGE} ; \
	done
