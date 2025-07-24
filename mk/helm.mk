HELM_ARGS ?=
HELM_DEV ?=

ifeq ($(HELM_DEV),true)
HELM_ARGS+= --dev
endif
HELM_PKG_EXTRA_CMD ?=

.PHONY: helm/update-version
helm/update-version:
	$(TOOLS_DIR)/releases/helm.sh $(HELM_ARGS) --update-version
	$(HELM_PKG_EXTRA_CMD)

.PHONY: helm/package
helm/package:
	$(TOOLS_DIR)/releases/helm.sh $(HELM_ARGS) --package

.PHONY: helm/release
helm/release:
	$(TOOLS_DIR)/releases/helm.sh $(HELM_ARGS) --release
