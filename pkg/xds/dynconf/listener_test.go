package dynconf_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/pkg/test/xds/builders"
	"github.com/kumahq/kuma/pkg/xds/dynconf"
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
		err := dynconf.AddConfigRoute(proxy, rs, "/meshmetric", body)
		Expect(err).ToNot(HaveOccurred())

		// then
		listeners := rs.Resources(envoy_resource.ListenerType)
		Expect(listeners).ToNot(BeEmpty())

		// when
		var listener *envoy_listener.Listener
		for _, res := range rs.Resources(envoy_resource.ListenerType) {
			if res.Origin == dynconf.Origin {
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

	It("configure mesh metric and embedded dns dynamic configurations", func() {
		// when
		rs := core_xds.NewResourceSet()
		body := make([]byte, 1024)
		err := dynconf.AddConfigRoute(proxy, rs, "/meshmetric", body)
		Expect(err).ToNot(HaveOccurred())

		// again with a different body size
		newBody := make([]byte, 2048)
		err = dynconf.AddConfigRoute(proxy, rs, "/dns", newBody)
		Expect(err).ToNot(HaveOccurred())

		// then
		listeners := rs.Resources(envoy_resource.ListenerType)
		Expect(listeners).ToNot(BeEmpty())

		// when
		var listener *envoy_listener.Listener
		for _, res := range rs.Resources(envoy_resource.ListenerType) {
			if res.Origin == dynconf.Origin {
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
		Expect(hcm.GetRouteConfig().MaxDirectResponseBodySizeBytes.GetValue()).To(Equal(uint32(2048)))
	})
})
