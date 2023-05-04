TEST_CONTAINER_TESTS_PATH ?= $(KUMA_DIR)/test/container-structure

define TEST_CONTAINER_STRUCTURES
.PHONY: test/container-structure/$(1)/$(2)
test/container-structure/$(1)/$(2): $(if $(CI),docker/$(1)/$(2)/load,image/$(1)/$(2))
	$(CONTAINER_STRUCTURE_TEST) test \
		--config $(if $(RESOLVE_CONTAINER_TEST_FILE),$$(call RESOLVE_CONTAINER_TEST_FILE,$(1)),$(TEST_CONTAINER_TESTS_PATH)/$(1).yaml) \
		--image $(call build_image,$(1),$(2))
endef

$(foreach goarch,$(SUPPORTED_GOARCHES),$(foreach image,$(IMAGES_RELEASE),$(eval $(call TEST_CONTAINER_STRUCTURES,$(image),$(goarch)))))

.PHONY: test/container-structure
test/container-structure: $(patsubst %,test/container-structure/%,$(filter-out $(foreach goarch,$(SUPPORTED_GOARCHES),%kuma-init/$(goarch)),$(ALL_RELEASE_WITH_ARCH)))
