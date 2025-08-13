package dynconf_test

import (
	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	"github.com/kumahq/kuma/pkg/xds/dynconf"
	"github.com/kumahq/kuma/pkg/xds/dynconf/metadata"
)

var _ = Describe("AddConfigRoute", func() {
	var proxy *core_xds.Proxy

	BeforeEach(func() {
		proxy = xds_builders.Proxy().
			WithID(*core_xds.BuildProxyId("default", "backend")).
			WithDataplane(samples.DataplaneBackendBuilder()).
			WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
			Build()
	})

	It("configure mesh metric dynamic configuration", func() {
		// when
		rs := core_xds.NewResourceSet()
		body := make([]byte, 1024)
		err := dynconf.AddConfigRoute(proxy, rs, "/meshmetric", "/meshmetric", body)
		Expect(err).ToNot(HaveOccurred())

		// then
		listeners := rs.Resources(envoy_resource.ListenerType)
		Expect(listeners).ToNot(BeEmpty())

		// when
		var listener *envoy_listener.Listener
		for _, res := range rs.Resources(envoy_resource.ListenerType) {
			if res.Origin == metadata.OriginDynamicConfig {
				listener = res.Resource.(*envoy_listener.Listener)
				break
			}
		}

		// then
		Expect(listener).ToNot(BeNil())
		Expect(listener.GetName()).To(Equal(dynconf.ListenerName))

		filterChains := listener.GetFilterChains()
		Expect(filterChains).To(HaveLen(1))
		Expect(filterChains[0].GetFilters()).To(HaveLen(1))

		filter := filterChains[0].GetFilters()[0]
		filterConfig, err := filter.GetTypedConfig().UnmarshalNew()
		Expect(err).ToNot(HaveOccurred())

		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		Expect(ok).To(BeTrue())
		Expect(hcm.GetRouteConfig().MaxDirectResponseBodySizeBytes.GetValue()).To(Equal(uint32(1024)))
	})

	It("configure mesh metric and embedded dns dynamic configurations with unified naming", func() {
		// when
		proxy.Metadata.Features = map[string]bool{
			xds_types.FeatureUnifiedResourceNaming: true,
		}
		rs := core_xds.NewResourceSet()
		body := make([]byte, 1024)
		err := dynconf.AddConfigRoute(proxy, rs, "meshmetric", "/meshmetric", body)
		Expect(err).ToNot(HaveOccurred())

		// again with a different body size
		newBody := make([]byte, 2048)
		err = dynconf.AddConfigRoute(proxy, rs, "dns", "/dns", newBody)
		Expect(err).ToNot(HaveOccurred())

		// then
		listeners := rs.Resources(envoy_resource.ListenerType)
		Expect(listeners).ToNot(BeEmpty())

		// when
		var listener *envoy_listener.Listener
		for _, res := range rs.Resources(envoy_resource.ListenerType) {
			if res.Origin == metadata.OriginDynamicConfig {
				listener = res.Resource.(*envoy_listener.Listener)
				break
			}
		}

		// then
		Expect(listener).ToNot(BeNil())
		Expect(listener.GetName()).To(Equal("system_dynamicconfig"))

		filterChains := listener.GetFilterChains()
		Expect(filterChains).To(HaveLen(1))
		Expect(filterChains[0].GetFilters()).To(HaveLen(1))

		filter := filterChains[0].GetFilters()[0]
		filterConfig, err := filter.GetTypedConfig().UnmarshalNew()
		Expect(err).ToNot(HaveOccurred())

		hcm, ok := filterConfig.(*envoy_hcm.HttpConnectionManager)
		Expect(ok).To(BeTrue())
		routes := hcm.GetRouteConfig().GetVirtualHosts()[0].Routes
		names := util_slices.Map(routes, func(route *envoy_route.Route) string {
			return route.Name
		})
		Expect(names).To(ConsistOf("system_dynamicconfig_meshmetric_not_modified", "system_dynamicconfig_meshmetric", "system_dynamicconfig_dns_not_modified", "system_dynamicconfig_dns"))
		Expect(hcm.GetRouteConfig().MaxDirectResponseBodySizeBytes.GetValue()).To(Equal(uint32(2048)))
	})
})
