TEST_CONFIGS_PATH_ROOT ?= tools/container-structure-test

define TEST_CONTAINER_STRUCTURES
TEST_CONFIGS_PATH_FILENAME_$(1) ?= $(1)
TEST_CONFIGS_PATH_$(1) ?= $(KUMA_DIR)/$(TEST_CONFIGS_PATH_ROOT)/$$(TEST_CONFIGS_PATH_FILENAME_$(1)).yaml

.PHONY: test/container-structure/$(1)/$(2)
test/container-structure/$(1)/$(2): docker/$(1)/$(2)/load
	$(CONTAINER_STRUCTURE_TEST) test \
		--config $$(TEST_CONFIGS_PATH_$(1)) \
		--image $(call build_image,$(1),$(2))
endef

$(foreach goarch,$(SUPPORTED_GOARCHES),$(foreach image,$(IMAGES_RELEASE),$(eval $(call TEST_CONTAINER_STRUCTURES,$(image),$(goarch)))))

.PHONY: test/container-structure
test/container-structure: $(patsubst %,test/container-structure/%,$(ALL_RELEASE_WITH_ARCH))
