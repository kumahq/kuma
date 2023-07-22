JSONNET_DIR=$(KUMA_DIR)/app/kumactl/data/install/k8s/metrics/grafana/jsonnet
JSONNET_OUTPUT_DIR=$(KUMA_DIR)/app/kumactl/data/install/k8s/metrics/grafana/dashboards
JSONNET_BUNDLER_CACHE_DIR=build/jsonnet

$(JSONNET_BUNDLER_CACHE_DIR):
	$(JSONNET_BUNDLER) update --jsonnetpkg-home=$(JSONNET_BUNDLER_CACHE_DIR)

.PHONY: fmt/jsonnet
fmt/jsonnet:
	$(JSONNETFMT) -i $(JSONNET_DIR)/*.jsonnet $(JSONNET_DIR)/lib/*.jsonnet

.PHONY: generate/jsonnet
generate/jsonnet: $(JSONNET_BUNDLER_CACHE_DIR)
	$(JSONNET) -J $(JSONNET_BUNDLER_CACHE_DIR) -m $(JSONNET_OUTPUT_DIR) $(JSONNET_DIR)/main.jsonnet

.PHONY: clean/jsonnet
clean/jsonnet:
	rm -rf $(JSONNET_BUNDLER_CACHE_DIR) $(JSONNET_OUTPUT_DIR)/*
