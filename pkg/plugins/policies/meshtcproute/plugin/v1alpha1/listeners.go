package v1alpha1

import (
	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator/metadata"
)

func GenerateOutboundListener(
	proxy *core_xds.Proxy,
	svc meshroute_xds.DestinationService,
	splits []envoy_common.Split,
) (*core_xds.Resource, error) {
	unifiedNaming := proxy.Metadata.HasFeature(types.FeatureUnifiedResourceNaming)
	bindOutbounds := proxy.Metadata.HasFeature(types.FeatureBindOutbounds)
	transparentProxy := !bindOutbounds && proxy.GetTransparentProxy().Enabled()

	address := svc.Outbound.GetAddressWithFallback("127.0.0.1")
	port := svc.Outbound.GetPort()

	listenerName := envoy_names.GetOutboundListenerName(address, port)
	listenerStatPrefix := ""
	tcpProxyStatPrefix := svc.KumaServiceTagValue
	if id, ok := svc.Outbound.AssociatedServiceResource(); ok && unifiedNaming {
		listenerName = id.String()
		listenerStatPrefix = listenerName
		tcpProxyStatPrefix = listenerName
	}

	tags := envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag)

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.TCPProxy(tcpProxyStatPrefix, splits...)).
		ConfigureIf(svc.Protocol == mesh.ProtocolKafka, envoy_listeners.Kafka(tcpProxyStatPrefix))

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.StatPrefix(listenerStatPrefix)).
		Configure(envoy_listeners.OutboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.TransparentProxying(transparentProxy)).
		Configure(envoy_listeners.TagsMetadata(tags)).
		Configure(envoy_listeners.FilterChain(filterChain))

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}

	return &core_xds.Resource{
		Name:           resource.GetName(),
		Origin:         metadata.OriginOutbound,
		Resource:       resource,
		ResourceOrigin: svc.Outbound.Resource,
		Protocol:       svc.Protocol,
	}, nil
}

func generateFromService(
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
	clusterCache map[common_api.BackendRefHash]string,
	servicesAccumulator envoy_common.ServicesAccumulator,
	toRulesTCP rules.ToRules,
	svc meshroute_xds.DestinationService,
) (*core_xds.ResourceSet, error) {
	toRulesHTTP := proxy.Policies.Dynamic[meshhttproute_api.MeshHTTPRouteType].ToRules
	unifiedNaming := proxy.Metadata.HasFeature(types.FeatureUnifiedResourceNaming)

	backendRefs := getBackendRefs(toRulesTCP, toRulesHTTP, svc, meshCtx)
	if len(backendRefs) == 0 {
		return nil, nil
	}

	splits := meshroute_xds.MakeTCPSplit(clusterCache, servicesAccumulator, backendRefs, meshCtx, unifiedNaming)

	listener, err := GenerateOutboundListener(proxy, svc, splits)
	if err != nil {
		return nil, err
	}
	return core_xds.NewResourceSet().Add(listener), nil
}

func generateListeners(
	proxy *core_xds.Proxy,
	toRulesTCP rules.ToRules,
	servicesAccumulator envoy_common.ServicesAccumulator,
	meshCtx xds_context.MeshContext,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	// Cluster cache protects us from creating excessive amount of clusters.
	// For one outbound we pick one traffic route, so LB and Timeout are
	// the same.
	clusterCache := map[common_api.BackendRefHash]string{}
	for _, svc := range meshroute_xds.CollectServices(proxy, meshCtx) {
		rs, err := generateFromService(
			meshCtx,
			proxy,
			clusterCache,
			servicesAccumulator,
			toRulesTCP,
			svc,
		)
		if err != nil {
			return nil, err
		}
		resources.AddSet(rs)
	}

	return resources, nil
}
