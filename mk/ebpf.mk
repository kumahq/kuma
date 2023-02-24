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

RELEASE_TAG ?= main-dd8b1946a31a8ce03009a2743b18ebcc716cda61
RELEASE_REPO ?= https://github.com/kumahq/merbridge
# You can provide your own url if the tarball with all ebpf programs should be
# fetched from somewhere else
TARBALL_URL ?= $(RELEASE_REPO)/releases/download/$(RELEASE_TAG)
# Where should be placed directory with $(RELEASE_TAG) as name placed. We don't
# allow to overwrite the final $(BUILD_OUTPUT) variable,  because without $(RELEASE_TAG) and $(GOARCH), it may result in situation
# where without realizing, ebpf programs are not re-fetched when $(RELEASE_TAG) or $(GOARCH) changes
BASE_BUILD_OUTPUT = build/ebpf/$(RELEASE_TAG)
# Path where ebpf programs should be placed, to be compiled in when building kumactl
COMPILE_IN_PATH ?= ./pkg/transparentproxy/ebpf/programs

# We are placing ebpf programs inside $(BUILD_OUTPUT_WITH_TAG) directory first,
# as by default it contains $(RELEASE_TAG) in the path, which means
# if the tag changes, we will re-fetch programs

define make_ebpf_targets
$(1)/mb_*: | $(1)
	curl --progress-bar --location $(TARBALL_URL)/all-$(3).tar.gz | tar -C $$(@D) -xz

$(2)/mb_*: | $(2) $(1)/mb_*
	cp $(1)/mb_* $(2)

# Make $(2) $(1) directories if they don't
# exist
$(1) $(2):
	mkdir -p $$@
endef

EBF_ARCH:=amd64 arm64
$(foreach elt,$(EBF_ARCH),$(eval $(call make_ebpf_targets,$(BASE_BUILD_OUTPUT)/$(elt),$(COMPILE_IN_PATH)/$(elt),$(elt))))

EBF_TARGETS:=$(foreach elt,$(EBF_ARCH),$(BASE_BUILD_OUTPUT)/$(elt)/mb_* $(COMPILE_IN_PATH)/$(elt)/mb_*)
.PHONY: build/ebpf
build/ebpf: $(EBF_TARGETS)

.PHONY: clean/ebpf
clean/ebpf :
	-rm -rf $(BUILD_DIR)/ebpf
	find $(COMPILE_IN_PATH) -type f -name 'mb_*' -exec rm {} \;
