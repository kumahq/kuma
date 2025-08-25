TEST_CONTAINER_TESTS_PATH ?= $(KUMA_DIR)/test/container-structure

define TEST_CONTAINER_STRUCTURES_BY_ARCH
.PHONY: test/container-structure/$(1)/$(2)
test/container-structure/$(1)/$(2): $(if $(CI),docker/load/$(1)/$(2),image/$(1)/$(2))
	$(CONTAINER_STRUCTURE_TEST) test \
		--config $(if $(RESOLVE_CONTAINER_TEST_FILE),$$(call RESOLVE_CONTAINER_TEST_FILE,$(1)),$(TEST_CONTAINER_TESTS_PATH)/$(1).yaml) \
		--platform $(2) \
		--image $(call build_image,$(1),$(2))
endef
$(foreach goarch,$(SUPPORTED_GOARCHES),$(foreach image,$(IMAGES_RELEASE),$(eval $(call TEST_CONTAINER_STRUCTURES_BY_ARCH,$(image),$(goarch)))))

define TEST_CONTAINER_STRUCTURES
.PHONY: test/container-structure/$(1)
test/container-structure/$(1): $(foreach goarch,$(ENABLED_GOARCHES),test/container-structure/$(1)/$(goarch))
endef
$(foreach image,$(IMAGES_RELEASE),$(eval $(call TEST_CONTAINER_STRUCTURES,$(image))))

.PHONY: test/container-structure
test/container-structure: $(foreach image,$(IMAGES_RELEASE),test/container-structure/$(image))
