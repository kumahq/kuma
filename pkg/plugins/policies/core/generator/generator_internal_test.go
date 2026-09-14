package generator

import (
	"encoding/json"

	envoy_listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	envoy_hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/test/resources/samples"
	xds_builders "github.com/kumahq/kuma/v3/pkg/test/xds/builders"
)

var _ = Describe("writeUnifiedOtelRoute", func() {
	It("should write the /otel route for a proxy that advertises no features", func() {
		// given
		socketPath := core_xds.OpenTelemetrySocketName("/tmp", "otel-backend")
		proxy := xds_builders.Proxy().
			WithID(*core_xds.BuildProxyId("default", "backend")).
			WithDataplane(samples.DataplaneBackendBuilder()).
			WithMetadata(&core_xds.DataplaneMetadata{WorkDir: "/tmp"}).
			Build()
		proxy.OtelPipeBackends = &core_xds.OtelPipeBackends{}
		proxy.OtelPipeBackends.AddSignal("otel-backend", core_xds.OtelPipeBackend{
			Name:       "otel-backend",
			SocketPath: socketPath,
			Endpoint:   "collector.mesh:4317",
		}, core_xds.OtelSignalTraces, core_xds.OtelSignalRuntimePlan{Enabled: true})
		rs := core_xds.NewResourceSet()

		// when
		Expect(writeUnifiedOtelRoute(rs, proxy)).To(Succeed())

		// then
		listeners := rs.Resources(envoy_resource.ListenerType)
		Expect(listeners).To(HaveLen(1))
		var listener *envoy_listener.Listener
		for _, res := range listeners {
			listener = res.Resource.(*envoy_listener.Listener)
		}
		hcm := &envoy_hcm.HttpConnectionManager{}
		filter := listener.GetFilterChains()[0].GetFilters()[0]
		Expect(filter.GetTypedConfig().UnmarshalTo(hcm)).To(Succeed())

		var body string
		for _, route := range hcm.GetRouteConfig().GetVirtualHosts()[0].GetRoutes() {
			if route.GetMatch().GetPath() == core_xds.OtelDynconfPath && route.GetDirectResponse().GetStatus() == 200 {
				body = route.GetDirectResponse().GetBody().GetInlineString()
			}
		}
		var dpConfig core_xds.OtelDpConfig
		Expect(json.Unmarshal([]byte(body), &dpConfig)).To(Succeed())
		Expect(dpConfig.Backends).To(HaveLen(1))
		Expect(dpConfig.Backends[0].SocketPath).To(Equal(socketPath))
		Expect(dpConfig.Backends[0].Endpoint).To(Equal("collector.mesh:4317"))
		Expect(dpConfig.Backends[0].Traces.Enabled).To(BeTrue())
	})
})
