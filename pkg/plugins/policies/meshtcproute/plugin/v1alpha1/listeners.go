package v1alpha1

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func generateFromService(
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
	clusterCache map[common_api.BackendRefHash]string,
	servicesAccumulator envoy_common.ServicesAccumulator,
	toRulesTCP rules.ToRules,
	svc meshroute_xds.DestinationService,
) (*core_xds.ResourceSet, error) {
	toRulesHTTP := proxy.Policies.Dynamic[meshhttproute_api.MeshHTTPRouteType].ToRules

	resources := core_xds.NewResourceSet()

	serviceName := svc.ServiceName
	protocol := svc.Protocol

	backendRefs := getBackendRefs(toRulesTCP, toRulesHTTP, svc, protocol, meshCtx)
	if len(backendRefs) == 0 {
		return nil, nil
	}

	splits := meshroute_xds.MakeTCPSplit(clusterCache, servicesAccumulator, backendRefs, meshCtx)
	filterChain := buildFilterChain(proxy, serviceName, splits, protocol)

	listener, err := buildOutboundListener(proxy, svc, filterChain)
	if err != nil {
		return nil, errors.Wrap(err, "cannot build listener")
	}

	resources.Add(&core_xds.Resource{
		Name:           listener.GetName(),
		Origin:         generator.OriginOutbound,
		Resource:       listener,
		ResourceOrigin: svc.Outbound.Resource,
		Protocol:       protocol,
	})
	return resources, nil
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

func buildOutboundListener(
	proxy *core_xds.Proxy,
	svc meshroute_xds.DestinationService,
	opts ...envoy_listeners.ListenerBuilderOpt,
) (envoy_common.NamedResource, error) {
	var tags envoy_tags.Tags
	if svc.Outbound.LegacyOutbound != nil {
		tags = svc.Outbound.LegacyOutbound.Tags
	}

	// build listener name in format: "outbound:[IP]:[Port]"
	// i.e. "outbound:240.0.0.0:80"
	builder := envoy_listeners.NewOutboundListenerBuilder(
		proxy.APIVersion,
		svc.Outbound.GetAddress(),
		svc.Outbound.GetPort(),
		core_xds.SocketAddressProtocolTCP,
	)

	tproxy := envoy_listeners.TransparentProxying(
		proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying(),
	)

	tagsMetadata := envoy_listeners.TagsMetadata(
		tags.WithoutTags(mesh_proto.MeshTag),
	)

	return builder.Configure(
		tproxy,
		tagsMetadata,
	).Configure(opts...).Build()
}
