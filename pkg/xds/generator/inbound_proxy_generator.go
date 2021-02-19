package generator

import (
	"github.com/pkg/errors"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/validators"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"

	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

// OriginInbound is a marker to indicate by which ProxyGenerator resources were generated.
const OriginInbound = "inbound"

type InboundProxyGenerator struct {
}

func (g InboundProxyGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	endpoints, err := proxy.Dataplane.Spec.Networking.GetInboundInterfaces()
	if err != nil {
		return nil, err
	}
	resources := model.NewResourceSet()
	for i, endpoint := range endpoints {
		// we do not create inbounds for serviceless
		if endpoint.IsServiceLess() {
			continue
		}

		iface := proxy.Dataplane.Spec.Networking.Inbound[i]
		protocol := mesh_core.ParseProtocol(iface.GetProtocol())

		// generate CDS resource
		localClusterName := envoy_names.GetLocalClusterName(endpoint.WorkloadPort)
		clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion).
			Configure(envoy_clusters.StaticCluster(localClusterName, endpoint.WorkloadIP, endpoint.WorkloadPort))
		switch protocol {
		case mesh_core.ProtocolHTTP2, mesh_core.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}
		cluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate cluster %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), localClusterName)
		}
		resources.Add(&model.Resource{
			Name:     localClusterName,
			Resource: cluster,
			Origin:   OriginInbound,
		})

		// generate LDS resource
		service := iface.GetService()
		inboundListenerName := envoy_names.GetInboundListenerName(endpoint.DataplaneIP, endpoint.DataplanePort)
		filterChainBuilder := func() *envoy_listeners.FilterChainBuilder {
			filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion)
			switch protocol {
			// configuration for HTTP case
			case mesh_core.ProtocolHTTP, mesh_core.ProtocolHTTP2:
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(localClusterName)).
					Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint])).
					Configure(envoy_listeners.Tracing(proxy.Policies.TracingBackend)).
					Configure(envoy_listeners.HttpInboundRoute(service, envoy_common.ClusterSubset{ClusterName: localClusterName}))
			case mesh_core.ProtocolGRPC:
				filterChainBuilder.
					Configure(envoy_listeners.HttpConnectionManager(localClusterName)).
					Configure(envoy_listeners.GrpcStats()).
					Configure(envoy_listeners.FaultInjection(proxy.Policies.FaultInjections[endpoint])).
					Configure(envoy_listeners.Tracing(proxy.Policies.TracingBackend)).
					Configure(envoy_listeners.HttpInboundRoute(service, envoy_common.ClusterSubset{ClusterName: localClusterName}))
			case mesh_core.ProtocolKafka:
				filterChainBuilder.
					Configure(envoy_listeners.Kafka(localClusterName)).
					Configure(envoy_listeners.TcpProxy(localClusterName, envoy_common.ClusterSubset{ClusterName: localClusterName}))
			case mesh_core.ProtocolTCP:
				fallthrough
			default:
				// configuration for non-HTTP cases
				filterChainBuilder.Configure(envoy_listeners.TcpProxy(localClusterName, envoy_common.ClusterSubset{ClusterName: localClusterName}))
			}
			return filterChainBuilder.
				Configure(envoy_listeners.ServerSideMTLS(ctx, proxy.Metadata)).
				Configure(envoy_listeners.NetworkRBAC(inboundListenerName, ctx.Mesh.Resource.MTLSEnabled(), proxy.Policies.TrafficPermissions[endpoint]))
		}()
		inboundListener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.InboundListener(inboundListenerName, endpoint.DataplaneIP, endpoint.DataplanePort, model.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.FilterChain(filterChainBuilder)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), inboundListenerName)
		}
		resources.Add(&model.Resource{
			Name:     inboundListenerName,
			Resource: inboundListener,
			Origin:   OriginInbound,
		})
	}
	return resources, nil
}
