package generator

import (
	"context"

	core_xds "github.com/kumahq/kuma/v2/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/v2/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/v2/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v2/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v2/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v2/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v2/pkg/xds/generator/metadata"
)

const (
	httpbinClusterName  = "httpbin_cluster"
	httpbinListenerName = "httpbin_listener"
	httpbinListenerPort = 10001

	// "outbound" that routes to the fake egress on the same pod
	egressClusterName  = "egress_cluster"
	egressListenerName = "egress_outbound_listener"
	egressListenerPort = 10002
)

// HttpbinProxyGenerator generates a hardcoded listener on port 10001
// that routes traffic to httpbin.org via STRICT_DNS cluster.
type HttpbinProxyGenerator struct{}

func (g HttpbinProxyGenerator) Generate(_ context.Context, _ *core_xds.ResourceSet, _ xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	// Create STRICT_DNS cluster pointing to httpbin.org
	// Using hostname (not IP) triggers STRICT_DNS mode
	clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, httpbinClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(
			proxy.Metadata.GetIPv6Enabled(),
			core_xds.Endpoint{Target: "httpbin.org", Port: 80},
		)).
		Configure(envoy_clusters.DefaultTimeout())

	envoyCluster, err := clusterBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     httpbinClusterName,
		Resource: envoyCluster,
		Origin:   metadata.OriginInbound,
	})

	// Create listener on port 10001 with TCP proxy to httpbin cluster
	cluster := plugins_xds.NewClusterBuilder().WithName(httpbinClusterName).Build()

	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.TcpProxyDeprecated(httpbinClusterName, cluster))

	listenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion, httpbinListenerName).
		Configure(envoy_listeners.InboundListener(proxy.Dataplane.Spec.GetNetworking().GetAddress(), httpbinListenerPort, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(filterChainBuilder))

	envoyListener, err := listenerBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     httpbinListenerName,
		Resource: envoyListener,
		Origin:   metadata.OriginInbound,
	})

	// Create cluster pointing to the fake egress listener on this same pod (pod_ip:10001)
	podIP := proxy.Dataplane.Spec.GetNetworking().GetAddress()
	egressClusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, egressClusterName).
		Configure(envoy_clusters.ProvidedEndpointCluster(
			proxy.Metadata.GetIPv6Enabled(),
			core_xds.Endpoint{Target: podIP, Port: httpbinListenerPort},
		)).
		Configure(envoy_clusters.DefaultTimeout())

	egressCluster, err := egressClusterBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     egressClusterName,
		Resource: egressCluster,
		Origin:   metadata.OriginOutbound,
	})

	// Create "outbound" listener on port 10002 that routes to the egress cluster
	egressClusterRef := plugins_xds.NewClusterBuilder().WithName(egressClusterName).Build()

	egressFilterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.TcpProxyDeprecated(egressClusterName, egressClusterRef))

	egressListenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion, egressListenerName).
		Configure(envoy_listeners.InboundListener("127.0.0.1", egressListenerPort, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.FilterChain(egressFilterChainBuilder))

	egressListener, err := egressListenerBuilder.Build()
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     egressListenerName,
		Resource: egressListener,
		Origin:   metadata.OriginOutbound,
	})

	return resources, nil
}
