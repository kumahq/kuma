.PHONY: help
help: ## Display this help screen
	@# Display top-level targets since they are the ones most developes will need.
	@grep -h -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort -k1 | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@# Now show hierarchical targets in separate sections.
	@grep -h -E '^[a-zA-Z0-9_-]+/[a-zA-Z0-9/_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk '{print $$1}' | \
		awk -F/ '{print $$1}' | \
		sort -u | \
	while read section ; do \
		echo; \
		grep -h -E "^$$section/[^:]+:.*?## .*$$" $(MAKEFILE_LIST) | sort -k1 | \
			awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' ; \
	done

include mk/dev.mk

include mk/api.mk
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
include mk/docs.mk
include mk/envoy.mk
