package generator

import (
	"context"
	"maps"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	"github.com/kumahq/kuma/v3/pkg/core/naming"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	defaults_mesh "github.com/kumahq/kuma/v3/pkg/defaults/mesh"
	plugins_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds"
	"github.com/kumahq/kuma/v3/pkg/util/net"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/v3/pkg/xds/envoy/clusters"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

type InboundProxyGenerator struct{}

func (g InboundProxyGenerator) Generate(_ context.Context, _ *core_xds.ResourceSet, _ xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	for i, endpoint := range proxy.Dataplane.Spec.Networking.GetInboundInterfaces() {
		// we do not create inbounds for serviceless
		if endpoint.IsServiceLess() {
			continue
		}

		iface := proxy.Dataplane.Spec.Networking.Inbound[i]
		inboundProtocol := iface.GetProtocol()
		protocol := core_meta.ParseProtocol(inboundProtocol)
		// the cluster, the listener and their stat prefixes all share this name
		contextualName := naming.MustContextualInboundName(proxy.Dataplane, endpoint.InboundName)

		// generate CDS resource
		clusterBuilder := envoy_clusters.NewClusterBuilder(proxy.APIVersion, contextualName).
			Configure(envoy_clusters.ProvidedEndpointCluster(false, core_xds.Endpoint{Target: endpoint.WorkloadIP, Port: endpoint.WorkloadPort})).
			Configure(envoy_clusters.Timeout(defaults_mesh.DefaultInboundTimeout(), protocol))
		// localhost traffic is routed dirrectly to the application, in case of other interface we are going to set source address to
		// 127.0.0.6 to avoid redirections and thanks to first iptables rule just return fast
		if proxy.GetTransparentProxy().Enabled() && (endpoint.WorkloadIP != core_meta.LoopbackIPv4.String() && endpoint.WorkloadIP != core_meta.LoopbackIPv6.String()) {
			bindIP := metadata.TransparentInPassThroughIPv4
			if net.IsAddressIPv6(endpoint.WorkloadIP) {
				bindIP = metadata.TransparentInPassThroughIPv6
			}
			clusterBuilder.Configure(envoy_clusters.UpstreamBindConfig(bindIP, 0))
		}

		switch protocol {
		case core_meta.ProtocolHTTP:
			clusterBuilder.Configure(envoy_clusters.Http())
		case core_meta.ProtocolHTTP2, core_meta.ProtocolGRPC:
			clusterBuilder.Configure(envoy_clusters.Http2())
		}
		envoyCluster, err := clusterBuilder.Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate cluster %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), contextualName)
		}
		resources.Add(&core_xds.Resource{
			Name:     contextualName,
			Resource: envoyCluster,
			Origin:   metadata.OriginInbound,
		})

		// generate LDS resource
		listenerTags := maps.Clone(proxy.Dataplane.GetMeta().GetLabels())
		if listenerTags == nil {
			listenerTags = map[string]string{}
		}
		if inboundProtocol != "" {
			listenerTags[mesh_proto.ProtocolTag] = inboundProtocol
		}

		// the plain, non-TLS shape of the listener. When the proxy has an identity,
		// the MeshTLS plugin replaces this listener with the Strict or Permissive
		// topology - it is the sole owner of that decision.
		inboundListener, err := envoy_listeners.NewListenerBuilder(proxy.APIVersion, contextualName).
			Configure(envoy_listeners.InboundListener(endpoint.DataplaneIP, endpoint.DataplanePort, core_xds.SocketAddressProtocolTCP, proxy.Metadata.HasFeature(xds_types.FeatureReusePort))).
			Configure(envoy_listeners.StatPrefix(contextualName)).
			Configure(envoy_listeners.TransparentProxying(proxy)).
			Configure(envoy_listeners.TagsMetadata(InboundListenerTags(listenerTags, contextualName))).
			Configure(envoy_listeners.FilterChain(FilterChainBuilder(protocol, proxy, endpoint))).
			Build()
		if err != nil {
			return nil, errors.Wrapf(err, "%s: could not generate listener %s", validators.RootedAt("dataplane").Field("networking").Field("inbound").Index(i), contextualName)
		}
		resources.Add(&core_xds.Resource{
			Name:     contextualName,
			Resource: inboundListener,
			Origin:   metadata.OriginInbound,
		})
	}
	return resources, nil
}

// FilterChainBuilder builds the plaintext filter chain of an inbound. Callers
// that want the chain protected by Kuma's TLS - only the MeshTLS plugin does -
// configure the transport socket on the returned builder.
func FilterChainBuilder(
	protocol core_meta.Protocol,
	proxy *core_xds.Proxy,
	endpoint mesh_proto.InboundInterface,
) *envoy_listeners.FilterChainBuilder {
	contextualName := naming.MustContextualInboundName(proxy.Dataplane, endpoint.InboundName)

	cluster := plugins_xds.NewClusterBuilder().WithName(contextualName).Build()

	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource)

	switch protocol {
	// configuration for HTTP case
	case core_meta.ProtocolHTTP, core_meta.ProtocolHTTP2:
		filterChainBuilder.
			Configure(envoy_listeners.HttpConnectionManager(contextualName, true, proxy.InternalAddresses, proxy.Metadata.GetIPv6Enabled())).
			Configure(envoy_listeners.HttpInboundRoute(contextualName, contextualName, cluster))
	case core_meta.ProtocolGRPC:
		filterChainBuilder.
			Configure(envoy_listeners.HttpConnectionManager(contextualName, true, proxy.InternalAddresses, proxy.Metadata.GetIPv6Enabled())).
			Configure(envoy_listeners.GrpcStats()).
			Configure(envoy_listeners.HttpInboundRoute(contextualName, contextualName, cluster))
	default:
		// configuration for non-HTTP cases
		filterChainBuilder.Configure(envoy_listeners.TcpProxyDeprecated(contextualName, cluster))
	}

	return filterChainBuilder.
		Configure(envoy_listeners.Timeout(defaults_mesh.DefaultInboundTimeout(), protocol))
}

// InboundListenerTags keeps the inbound's own tags, or when they are empty
// falls back to the contextual name (self_inbound_dp_<sectionName>) under
// kuma.io/unified-name so the listener stays selectable. The name carries no
// Dataplane identity, so it survives Pod churn that a Dataplane KRI would not.
func InboundListenerTags(tags map[string]string, contextualName string) map[string]string {
	if len(tags) > 0 {
		return tags
	}
	return map[string]string{mesh_proto.UnifiedNameTag: contextualName}
}
