TAG ?= $(shell git rev-parse --verify HEAD)

build:
	docker build -t kuma/install-cni:$(TAG) \
		 -f build/Dockerfile \
		 .