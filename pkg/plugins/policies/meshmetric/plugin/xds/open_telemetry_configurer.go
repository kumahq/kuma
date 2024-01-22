package xds

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
)

type OpenTelemetryConfigurer struct {
	Endpoint     *core_xds.Endpoint
	ListenerName string
	ClusterName  string
}

func (oc *OpenTelemetryConfigurer) ConfigureCluster(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	return envoy_clusters.NewClusterBuilder(proxy.APIVersion, oc.ClusterName).
		Configure(envoy_clusters.Http2()).
		Configure(envoy_clusters.ProvidedEndpointCluster(proxy.Dataplane.IsIPv6(), *oc.Endpoint)).
		Configure(envoy_clusters.ClientSideTLS([]core_xds.Endpoint{*oc.Endpoint})).
		Configure(envoy_clusters.DefaultTimeout()).
		Build()
}

func (oc *OpenTelemetryConfigurer) ConfigureListener(proxy *core_xds.Proxy) (envoy_common.NamedResource, error) {
	return envoy_listeners.NewListenerBuilder(proxy.APIVersion, oc.ListenerName).
		Configure(envoy_listeners.PipeListener(core_xds.OpenTelemetrySocketName(proxy.Metadata.WorkDir))).
		Configure(envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
				Configure(envoy_listeners.StaticEndpoints(oc.ListenerName, []*envoy_common.StaticEndpointPath{
					{
						ClusterName: oc.ClusterName,
						Path:        "/",
					},
				})).
				Configure(envoy_listeners.GrpcStats()),
		)).
		Build()
}
