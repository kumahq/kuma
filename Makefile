.PHONY: help
help: ## Display this help screen
	@# Display top-level targets since they are the ones most developes will need.
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo
	@# Now show hierarchical targets
	@grep -h -E '^[a-zA-Z0-9_-]+/[a-zA-Z0-9/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

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
