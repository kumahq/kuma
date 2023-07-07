package v1alpha1

import (
	"github.com/pkg/errors"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func generateListeners(
	proxy *core_xds.Proxy,
	toRulesTCP rules.Rules,
	servicesAccumulator envoy_common.ServicesAccumulator,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	// Cluster cache protects us from creating excessive amount of clusters.
	// For one outbound we pick one traffic route, so LB and Timeout are
	// the same.
	clusterCache := map[common_api.TargetRefHash]string{}
	sc := &meshroute_xds.SplitCounter{}
	networking := proxy.Dataplane.Spec.GetNetworking()
	routing := proxy.Routing
	toRulesHTTP := proxy.Policies.Dynamic[meshhttproute_api.MeshHTTPRouteType].
		ToRules.Rules

	for _, outbound := range networking.GetOutbound() {
		serviceName := outbound.GetServiceName()
		protocol := plugins_xds.InferProtocol(routing, serviceName)

		backendRefs := getBackendRefs(toRulesTCP, toRulesHTTP, serviceName,
			protocol)
		if len(backendRefs) == 0 {
			continue
		}

		splits := meshroute_xds.MakeTCPSplit(proxy, clusterCache, sc, servicesAccumulator, backendRefs)
		filterChain := buildFilterChain(proxy, serviceName, splits)

		listener, err := buildOutboundListener(proxy, outbound, filterChain)
		if err != nil {
			return nil, errors.Wrap(err, "cannot build listener")
		}

		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   generator.OriginOutbound,
			Resource: listener,
		})
	}

	return resources, nil
}

func buildOutboundListener(
	proxy *core_xds.Proxy,
	outbound *mesh_proto.Dataplane_Networking_Outbound,
	opts ...envoy_listeners.ListenerBuilderOpt,
) (envoy_common.NamedResource, error) {
	oface := proxy.Dataplane.Spec.GetNetworking().ToOutboundInterface(outbound)
	tags := outbound.GetTagsIncludingLegacy()

	// build listener name in format: "outbound:[IP]:[Port]"
	// i.e. "outbound:240.0.0.0:80"
	builder := envoy_listeners.NewOutboundListenerBuilder(
		proxy.APIVersion,
		oface.DataplaneIP,
		oface.DataplanePort,
		core_xds.SocketAddressProtocolTCP,
	)

	tproxy := envoy_listeners.TransparentProxying(
		proxy.Dataplane.Spec.GetNetworking().GetTransparentProxying(),
	)

	tagsMetadata := envoy_listeners.TagsMetadata(
		envoy_tags.Tags(tags).WithoutTags(mesh_proto.MeshTag),
	)

	return builder.Configure(
		tproxy,
		tagsMetadata,
	).Configure(opts...).Build()
}
