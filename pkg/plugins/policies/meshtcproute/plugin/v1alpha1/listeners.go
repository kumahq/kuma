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
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func GenerateOutboundListener(
	apiVersion core_xds.APIVersion,
	svc meshroute_xds.DestinationService,
	isTransparent bool,
	splits []envoy_common.Split,
) (*core_xds.Resource, error) {
	listener, err := envoy_listeners.NewOutboundListenerBuilder(apiVersion, svc.Outbound.GetAddress(), svc.Outbound.GetPort(), core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TransparentProxying(isTransparent)).
		Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag))).
		Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource).
			ConfigureIf(svc.Protocol == mesh.ProtocolKafka, envoy_listeners.Kafka(svc.KumaServiceTagValue)).
			Configure(envoy_listeners.TCPProxy(svc.KumaServiceTagValue, splits...)))).
		Build()
	if err != nil {
		return nil, err
	}

	return &core_xds.Resource{
		Name:           listener.GetName(),
		Origin:         generator.OriginOutbound,
		Resource:       listener,
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

	backendRefs := getBackendRefs(toRulesTCP, toRulesHTTP, svc, meshCtx)
	if len(backendRefs) == 0 {
		return nil, nil
	}

	splits := meshroute_xds.MakeTCPSplit(clusterCache, servicesAccumulator, backendRefs, meshCtx)

	isTransparent := !proxy.Metadata.HasFeature(types.FeatureBindOutbounds) && proxy.GetTransparentProxy().Enabled()

	listener, err := GenerateOutboundListener(proxy.APIVersion, svc, isTransparent, splits)
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
