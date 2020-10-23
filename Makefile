.PHONY: help
help: ## Display this help screen
	@grep -h -E '^[a-zA-Z0-9_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

include mk/dev.mk
include mk/build.mk
include mk/check.mk
include mk/test.mk
include mk/generate.mk
include mk/docker.mk
include mk/run.mk
include mk/kind.mk
include mk/e2e.mk
include mk/e2e.new.mk

tag=1.0.0-rc1-36-g6bea2fb2
push:
	docker push lobkovilya/kuma-cp:$(tag)
	docker push lobkovilya/kuma-dp:$(tag)
	docker push lobkovilya/kuma-init:$(tag)
	docker push lobkovilya/kuma-prometheus-sd:$(tag)