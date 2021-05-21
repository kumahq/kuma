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
include mk/k3d.mk
include mk/e2e.mk
include mk/e2e.new.mk
