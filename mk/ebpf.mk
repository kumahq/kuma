MERBRIDGE_RELEASE_TAG ?= main-20aa4b03cf256a261d1ebfa3af9f390186bec3ae
MERBRIDGE_RELEASE_REPO ?= https://github.com/kumahq/merbridge
TARBALL_NAME ?= all.tar.gz

BUILD_OUTPUT ?= $(KUMA_DIR)/build/ebpf/$(MERBRIDGE_RELEASE_TAG)

# Path where ebpf programs should be placed, to be compiled in when building
# kumactl
COMPILE_IN_PATH ?= $(KUMA_DIR)/pkg/transparentproxy/ebpf/programs

# We are placing ebpf programs inside $(BUILD_OUTPUT) directory first,
# as by default it contains $(MERBRIDGE_RELEASE_TAG) in the path, which means
# if the tag changes, we will re-fetch programs
.PHONY: build/ebpf
build/ebpf: $(BUILD_OUTPUT)/mb_* $(COMPILE_IN_PATH)/mb_*

# "| $(DIRECTORY_NAME)" is an order-only prerequisite, which means prerequisite
# that is never checked when determining if the target is out of date;
# even order-only prerequisites marked as phony will not cause the target
# to be rebuilt. In our case it's used to make $(DIRECTORY_NAME) directory, if
# it doesn't exist
# ref. https://www.gnu.org/software/make/manual/html_node/Prerequisite-Types.html

$(BUILD_OUTPUT)/mb_*: | $(BUILD_OUTPUT)
	MERBRIDGE_TAG=$(MERBRIDGE_RELEASE_TAG) \
	OUTPUT_PATH=$(@D) \
	RELEASE_REPO=$(MERBRIDGE_RELEASE_REPO) \
	TARBALL_NAME=$(TARBALL_NAME) \
	$(KUMA_DIR)/tools/ebpf/fetch.sh

$(COMPILE_IN_PATH)/mb_*: | $(COMPILE_IN_PATH)
	cp $(BUILD_OUTPUT)/mb_* $(COMPILE_IN_PATH)

# Make (COMPILE_IN_PATH) $(BUILD_OUTPUT) directories if they don't exist
$(COMPILE_IN_PATH) $(BUILD_OUTPUT):
	mkdir -p $@

.PHONY: clean/ebpf
clean/ebpf :
	-rm -rf $(BUILD_OUTPUT) $(COMPILE_IN_PATH)

