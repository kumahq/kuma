TEST_CONFIGS_PATH ?= tools/container-structure-test

define TEST_CONTAINER_STRUCTURES
.PHONY: test/container-structure/$(1)/$(2)
test/container-structure/$(1)/$(2): docker/$(1)/$(2)/load
	$(CONTAINER_STRUCTURE_TEST) test \
		--config $(KUMA_DIR)/$(TEST_CONFIGS_PATH)/$(1).yaml \
		--image $(call build_image,$(1),$(2))
endef

$(foreach goarch,$(SUPPORTED_GOARCHES),$(foreach image,$(IMAGES_RELEASE),$(eval $(call TEST_CONTAINER_STRUCTURES,$(image),$(goarch)))))

.PHONY: test/container-structure
test/container-structure: $(patsubst %,test/container-structure/%,$(ALL_RELEASE_WITH_ARCH))
