# Explanation of used GNU Make concepts used in the file:
#
# "| $(DIRECTORY_NAME)" is an order-only prerequisite, which means prerequisite
#   that is never checked when determining if the target is out of date;
#   even order-only prerequisites marked as phony will not cause the target
#   to be rebuilt. In our case it's used to make $(DIRECTORY_NAME) directory,
#   if it doesn't exist.
# ref. https://www.gnu.org/software/make/manual/html_node/Prerequisite-Types.html
#
# "$(@D)" is the directory part of the file name of the target, with the trailing
#   slash removed. If the value of $(@) is "dir/foo.o" then $(@D) will be
#   equal "dir". This value will be equal "." if $(@) does not contain a slash.
# ref. https://www.gnu.org/software/make/manual/html_node/Automatic-Variables.html#Automatic-Variables

RELEASE_TAG ?= main-20aa4b03cf256a261d1ebfa3af9f390186bec3ae
RELEASE_REPO ?= https://github.com/kumahq/merbridge
TARBALL_NAME ?= all.tar.gz
# You can provide your own url if the tarball with all ebpf programs should be
# fetched from somewhere else
TARBALL_URL ?= $(RELEASE_REPO)/releases/download/$(RELEASE_TAG)/$(TARBALL_NAME)
# Where should be placed directory with $(RELEASE_TAG) as name placed. We don't
# allow to overwrite the final $(BUILD_OUTPUT_WITH_TAG) variable, as if someone
# by mistake remove $(RELEASE_TAG) from the path, it may result in situation
# where without realizing, ebpf programs are not re-fetched when $(RELEASE_TAG)
# changes
BUILD_OUTPUT ?= build/ebpf
BUILD_OUTPUT_WITH_TAG = $(BUILD_OUTPUT)/$(RELEASE_TAG)
# Path where ebpf programs should be placed, to be compiled in when building
# kumactl
COMPILE_IN_PATH ?= pkg/transparentproxy/ebpf/programs

# We are placing ebpf programs inside $(BUILD_OUTPUT_WITH_TAG) directory first,
# as by default it contains $(RELEASE_TAG) in the path, which means
# if the tag changes, we will re-fetch programs
.PHONY: build/ebpf
build/ebpf: $(BUILD_OUTPUT_WITH_TAG)/mb_* $(COMPILE_IN_PATH)/mb_*

$(BUILD_OUTPUT_WITH_TAG)/mb_*: | $(BUILD_OUTPUT_WITH_TAG)
	curl --progress-bar --location $(TARBALL_URL) | tar -C $(@D) -xz

$(COMPILE_IN_PATH)/mb_*: | $(COMPILE_IN_PATH)
	cp $(BUILD_OUTPUT_WITH_TAG)/mb_* $(COMPILE_IN_PATH)

# Make $(COMPILE_IN_PATH) $(BUILD_OUTPUT_WITH_TAG) directories if they don't
# exist
$(COMPILE_IN_PATH) $(BUILD_OUTPUT_WITH_TAG):
	mkdir -p $@

.PHONY: clean/ebpf
clean/ebpf :
	-rm -rf $(BUILD_OUTPUT_WITH_TAG) $(COMPILE_IN_PATH)/mb_*

