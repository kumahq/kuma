MERBRIDGE_RELEASE_TAG ?= main-20aa4b03cf256a261d1ebfa3af9f390186bec3ae
MERBRIDGE_RELEASE_REPO ?= https://github.com/kumahq/merbridge
TARBALL_NAME ?= all.tar.gz

PROGRAMS = $(KUMA_DIR)/pkg/transparentproxy/ebpf/programs/mb_*

.PHONY: build/ebpf
build/ebpf: $(PROGRAMS)

# It's worth noting that changing MERBRIDGE_RELEASE_TAG won't re-trigger fetch
# if ebpf programs are already fetched. Remember to call `make clean/ebpf`
# after changing MERBRIDGE_RELEASE_TAG
$(PROGRAMS):
	MERBRIDGE_TAG=$(MERBRIDGE_RELEASE_TAG) \
	OUTPUT_PATH=$(@D) \
	RELEASE_REPO=$(MERBRIDGE_RELEASE_REPO) \
	TARBALL_NAME=$(TARBALL_NAME) \
	$(KUMA_DIR)/tools/ebpf/fetch.sh

.PHONY: clean/ebpf
clean/ebpf :
	-rm -rf $(PROGRAMS)

