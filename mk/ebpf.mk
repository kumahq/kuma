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
$(1)/.fetched: | $(1)/
	curl --progress-bar --location $(TARBALL_URL)/all-$(3).tar.gz | tar -C $(1) -xz
	touch $$@

$(2)/.copied: $(1)/.fetched | $(2)/
	command cp $(1)/mb_* $(2)
	touch $$@

# Make $(2) $(1) directories if they don't
# exist
$(1)/ $(2)/:
	mkdir -p $$@
endef

EBPF_ARCH:=amd64 arm64
$(foreach elt,$(EBPF_ARCH),$(eval $(call make_ebpf_targets,$(BASE_BUILD_OUTPUT)/$(elt),$(COMPILE_IN_PATH)/$(elt),$(elt))))

EBPF_TARGETS:=$(foreach elt,$(EBPF_ARCH),$(COMPILE_IN_PATH)/$(elt)/.copied)
.PHONY: build/ebpf
build/ebpf: $(EBPF_TARGETS)

.PHONY: clean/ebpf
clean/ebpf :
	-rm -rf $(BUILD_DIR)/ebpf
	rm -f $(foreach elt,$(EBPF_ARCH),$(COMPILE_IN_PATH)/$(elt)/.copied)
	find $(COMPILE_IN_PATH) -type f -name 'mb_*' -exec rm {} \;
