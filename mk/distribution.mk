DISTRIBUTION_LICENSE_PATH ?= tools/releases/templates
DISTRIBUTION_CONFIG_PATH ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml
# A list of all distributions
# OS:ARCH:COREDNS:ENVOY_FLAVOUR
# COREDNS is always coredns(CORDNS_EXT)
# If you don't want to include just put skip
DISTRIBUTION_LIST ?= linux:amd64:coredns:envoy linux:arm64:coredns:envoy darwin:amd64:coredns:envoy darwin:arm64:coredns:envoy

PULP_HOST ?= "https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE ?= $(PROJECT_NAME)
PULP_RELEASE_IMAGE ?= kong/release-script
DISTRIBUTION_TARGET_NAME ?= $(PROJECT_NAME)-$(BUILD_INFO_VERSION)
PULP_DIST_VERSION ?= release
ifneq (,$(findstring preview,$(BUILD_INFO_VERSION)))
	PULP_DIST_VERSION=preview
endif
DISTRIBUTION_FOLDER=build/distributions/$(GOOS)-$(GOARCH)/$(DISTRIBUTION_TARGET_NAME)
TAR_EXCLUDES=--exclude=passwd --exclude=group

# This function dynamically builds targets for building distribution packages and uploading them to pulp with a set of parameters
# $(1) - GOOS to build for
# $(2) - GOARCH to build for
# $(3) - coredns extension to use (or `skip` if we shouldn't include COREDNS)
# $(4) - primary envoy to use in the distribution (the binary that will be called `envoy`)
define make_distributions_target
build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME): build/artifacts-$(1)-$(2)/kumactl build/artifacts-$(1)-$(2)/kuma-cp build/artifacts-$(1)-$(2)/kuma-dp
	rm -rf $$@
	mkdir -p $$@/bin $$@/conf
	cp build/artifacts-$(1)-$(2)/kumactl/kumactl $$@/bin
	cp build/artifacts-$(1)-$(2)/kuma-cp/kuma-cp $$@/bin
	cp build/artifacts-$(1)-$(2)/kuma-dp/kuma-dp $$@/bin
	cp $(DISTRIBUTION_LICENSE_PATH)/* $$@
	cp $(DISTRIBUTION_CONFIG_PATH) $$@/conf
# CoreDNS is not included when the value is `skip` otherwise it's used as the COREDNS_EXT (which is most commonly just coredns)
ifneq ($(3),skip)
	$(MAKE) build/artifacts-$(1)-$(2)/coredns COREDNS_EXT=$(subst coredns,,$(3))
	cp build/artifacts-$(1)-$(2)/coredns/coredns $$@/bin
endif

# Package envoy
	$(MAKE) build/artifacts-$(1)-$(2)/envoy ENVOY_EXT_$(1)_$(2)=$(subst envoy,,$(4))
	cp -r build/artifacts-$(1)-$(2)/envoy/* $$@/bin

	# Set permissions correctly
	find $$@ -type f | xargs chmod 555
	# Text files don't have executable access
	find $$@ -type f -exec grep -I -q '' {} \; -print | xargs chmod 444
# Rename all executables to `.exe` when building the windows distrib
ifeq ($(1),windows)
        # find on darwin doesn't support `-executable` so we use `-perm +111` which we can't use on Linux
	find $$@/bin $(if $(findstring Darwin,$(shell uname)),-perm +111,-executable) -type f -exec mv {} {}.exe \;
endif

build/distributions/out/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2).tar.gz: build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME)
	mkdir -p $$(@D)
	# Create a tar with group and owner 0 and mtime of 0 (this makes builds reproducible).
	# Have the tar be just the `kuma-version` folder and nothing else at root
	# tar is different between darwin and Linux so executre different commands
ifeq ($(shell uname),Darwin)
	tar $(TAR_EXCLUDES) --strip-components 3 --numeric-owner -czvf $$@ $$<
else
	tar $(TAR_EXCLUDES) --mtime='1970-01-01 00:00:00' -C $$(dir $$<) --sort=name --owner=root:0 --group=root:0 --numeric-owner -czvf $$@ $$(notdir $$<)
endif
	cd $$(@D) && shasum -a 256 $$(notdir $$@) > $$(notdir $$@).sha256

.PHONY: publish/pulp/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2)
publish/pulp/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2):
	$$(call GATE_PUSH,docker run --rm \
		-e PULP_USERNAME="${PULP_USERNAME}" \
		-e PULP_PASSWORD="${PULP_PASSWORD}" \
		-e PULP_HOST=$(PULP_HOST) \
		-e CLOUDSMITH_API_KEY='$(CLOUDSMITH_API_KEY)' \
		-e CLOUDSMITH_DRY_RUN='' \
		-e IGNORE_CLOUDSMITH_FAILURES=x \
		-e USE_CLOUDSMITH=x \
		-e USE_PULP=x \
		-v $(TOP)/build/distributions/out:/files:ro \
		$(PULP_RELEASE_IMAGE) \
		release \
		--file=/files/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2).tar.gz \
		--package-type=$(PULP_PACKAGE_TYPE) \
		--dist-name=binaries \
		--dist-version=$(PULP_DIST_VERSION) \
		--publish)
endef

# These are meant to be used inside foreach
dist_os = $(word 1, $(subst :, ,$(elt)))
dist_arch = $(word 2, $(subst :, ,$(elt)))
dist_coredns = $(word 3, $(subst :, ,$(elt)))
dist_envoy = $(word 4, $(subst :, ,$(elt)))
dist_envoy_alt = $(word 5, $(subst :, ,$(elt)))
dist_name = $(dist_os)-$(dist_arch)
# Call make_distribution_target with each combination
$(foreach elt,$(DISTRIBUTION_LIST),$(eval $(call make_distributions_target,$(dist_os),$(dist_arch),$(dist_coredns),$(dist_envoy),$(dist_envoy_alt))))
ENABLED_DIST_NAMES=$(filter $(addprefix %,$(ENABLED_ARCH_OS)),$(foreach elt,$(DISTRIBUTION_LIST),$(dist_name)))

# Create a main target which will call the tar.gz target for each distribution
.PHONY: build/distributions ## Build tar.gz for each enabled distribution
build/distributions: build/distributions/out
ifeq ($(shell uname),Darwin)
	cat $</$(DISTRIBUTION_TARGET_NAME).sha256 | base64 > $@/artifact_digest_file.text
else
	cat $</$(DISTRIBUTION_TARGET_NAME).sha256 | base64 -w0 > $@/artifact_digest_file.text
endif

build/distributions/out: $(patsubst %,build/distributions/out/$(DISTRIBUTION_TARGET_NAME)-%.tar.gz,$(ENABLED_DIST_NAMES))
	cd $@; sha256sum *.tar.gz > $(DISTRIBUTION_TARGET_NAME).sha256

.PHONY: build/info/distribution/repo
build/info/cloudsmith_repository:
	@echo $(PULP_PACKAGE_TYPE)-binaries-$(PULP_DIST_VERSION)

# Create a main target which will publish to pulp each to the tar.gz built
.PHONY: publish/pulp ## Publish to pulp all enabled distributions
publish/pulp: $(addprefix publish/pulp/$(DISTRIBUTION_TARGET_NAME)-,$(ENABLED_DIST_NAMES))

.PHONY: clean/distributions
clean/distributions:
	rm -rf build/distributions
