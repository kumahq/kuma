DISTRIBUTION_NAME ?= kuma
DISTRIBUTION_LICENSE_PATH ?= tools/releases/templates
DISTRIBUTION_CONFIG_PATH ?= pkg/config/app/kuma-cp/kuma-cp.defaults.yaml
# A list of all distributions
# OS:ARCH:coredns:ENVOY_FLAVOUR:ENVOY_FLAVOUR
# The second ENVOY_FLAVOUR is optional
# If you don't want to include coredns just put an empty string
DISTRIBUTION_LIST ?= linux:amd64:coredns:alpine-opt:centos-opt linux:arm64:coredns:alpine-opt darwin:amd64:coredns:darwin-opt darwin:arm64:coredns:darwin-opt

PULP_HOST ?= "https://api.pulp.konnect-prod.konghq.com"
PULP_PACKAGE_TYPE ?= kuma
PULP_DIST_NAME ?= $(DISTRIBUTION_NAME)-$(shell echo $(BUILD_INFO_VERSION) | awk -F '.' '{ print $$1"."$$2"."$$3 }')

# This function dynamically builds targets for building distribution packages and uploading them to pulp with a set of parameters
define make_distributions_target
build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME): GOOS=$(1)
build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME): GOARCH=$(2)
build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME):
	rm -rf $$@
	mkdir -p $$@/bin $$@/conf
	$(MAKE) build/kumactl GOOS=$(1) GOARCH=$(2)
	cp build/artifacts-$(1)-$(2)/kumactl/kumactl $$@/bin
	$(MAKE) build/kuma-cp GOOS=$(1) GOARCH=$(2)
	cp build/artifacts-$(1)-$(2)/kuma-cp/kuma-cp $$@/bin
	$(MAKE) build/kuma-dp GOOS=$(1) GOARCH=$(2)
	cp build/artifacts-$(1)-$(2)/kuma-dp/kuma-dp $$@/bin
	cp $(DISTRIBUTION_LICENSE_PATH)/* $$@
	cp $(DISTRIBUTION_CONFIG_PATH) $$@/conf
# CoreDNS doesn't always need to be included
ifeq ($(3),coredns)
	$(MAKE) build/coredns GOOS=$(1) GOARCH=$(2)
	cp build/artifacts-$(1)-$(2)/coredns/coredns $$@/bin
endif
# A first possible envoy to package
ifneq ($(4),)
	$(MAKE) build/envoy/$(1)-$(2)/$(4)/envoy GOOS=$(1) GOARCH=$(2)
	cp build/envoy/$(1)-$(2)/$(4)/envoy $$@/bin
endif
# A second possible envoy to package
ifneq ($(5),)
	$(MAKE) build/envoy/$(1)-$(2)/$(4)/envoy GOOS=$(1) GOARCH=$(2)
	cp build/envoy/$(1)-$(2)/$(4)/envoy $$@/bin/envoy-$(5)
endif
	# Set permissions correctly
	find $$@ -type f | xargs chmod 555
	# Text files don't have executable access
	find $$@ -type f -exec grep -I -q '' {} \; -print | xargs chmod 444

build/distributions/out/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2).tar.gz: build/distributions/$(1)-$(2)/$(DISTRIBUTION_TARGET_NAME)
	mkdir -p build/distributions/out
	tar --strip-components 3 --numeric-owner -czvf $$@ $$<
	shasum -a 256 $$@ > $$@.sha256

.PHONY: publish/pulp/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2)
publish/pulp/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2):
	docker run --rm \
	  -e PULP_USERNAME="${PULP_USERNAME}" -e PULP_PASSWORD="${PULP_PASSWORD}" \
	  -e PULP_HOST=$(PULP_HOST) \
	  -v $(TOP)/build/distributions/out:/files:ro -it kong/release-script \
	  --file /files/$(DISTRIBUTION_TARGET_NAME)-$(1)-$(2).tar.gz \
	  --package-type $(PULP_PACKAGE_TYPE) --dist-name alpine
endef
# Removed --publish to see if internal publication works :)

# Call make_distribution_target with each combination
$(foreach elt,$(DISTRIBUTION_LIST),$(eval $(call make_distributions_target,$(word 1, $(subst :, ,$(elt))),$(word 2, $(subst :, ,$(elt))),$(word 3, $(subst :, ,$(elt))),$(word 4, $(subst :, ,$(elt))),$(word 5, $(subst :, ,$(elt))))))

# Create a main target which will call the tar.gz target for each distribution
DIST_TARGETS:=$(foreach elt, $(DISTRIBUTION_LIST), build/distributions/out/$(DISTRIBUTION_TARGET_NAME)-$(word 1, $(subst :, ,$(elt)))-$(word 2, $(subst :, ,$(elt))).tar.gz)
.PHONY: build/distributions
build/distributions: $(DIST_TARGETS)

# Create a main target which will publish to pulp each to the tar.gz built
PULP_PUBLISH_TARGETS:=$(foreach elt, $(DISTRIBUTION_LIST), publish/pulp/$(DISTRIBUTION_TARGET_NAME)-$(word 1, $(subst :, ,$(elt)))-$(word 2, $(subst :, ,$(elt))))
.PHONY: publish/pulp
publish/pulp: $(PULP_PUBLISH_TARGETS)

.PHONY: clean/distributions
clean/distributions:
	rm -rf build/distributions
