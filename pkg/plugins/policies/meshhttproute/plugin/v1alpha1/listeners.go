package v1alpha1

import (
	"reflect"

	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func generateFromService(
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	rules rules.ToRules,
	svc meshroute_xds.DestinationService,
) (*core_xds.ResourceSet, error) {
	var tags envoy_tags.Tags
	if svc.Outbound.LegacyOutbound != nil {
		tags = svc.Outbound.LegacyOutbound.GetTags()
	}
	listenerBuilder := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, svc.Outbound.GetAddress(), svc.Outbound.GetPort(), core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
		Configure(envoy_listeners.TagsMetadata(tags.WithoutTags(mesh_proto.MeshTag)))

	resourceName := svc.ServiceName

	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.AddFilterChainConfigurer(&envoy_listeners_v3.HttpConnectionManagerConfigurer{
			StatsName:                resourceName,
			ForwardClientCertDetails: false,
			NormalizePath:            true,
		}))

	var routes []xds.OutboundRoute
	for _, route := range prepareRoutes(rules, svc, meshCtx) {
		split := meshroute_xds.MakeHTTPSplit(clusterCache, servicesAcc, route.BackendRefs, meshCtx)
		if split == nil {
			continue
		}
		for _, filter := range route.Filters {
			if filter.Type == api.RequestMirrorType {
				// we need to create a split for the mirror backend
				_ = meshroute_xds.MakeHTTPSplit(
					clusterCache, servicesAcc,
					[]core_model.ResolvedBackendRef{*core_model.NewResolvedBackendRef(pointer.To(core_model.LegacyBackendRef(filter.RequestMirror.BackendRef)))},
					meshCtx,
				)
			}
		}
		routes = append(routes, xds.OutboundRoute{
			Hash:                    route.Hash,
			Match:                   route.Match,
			Filters:                 route.Filters,
			Split:                   split,
			BackendRefToClusterName: clusterCache,
		})
	}

	if len(routes) == 0 {
		return nil, nil
	}

	var outboundRouteName string
	if svc.Outbound.Resource != nil && svc.Outbound.Resource.ResourceType == core_model.ResourceType(common_api.MeshExternalService) {
		outboundRouteName = resourceName
	}
	outboundRouteConfigurer := &xds.HttpOutboundRouteConfigurer{
		Name:    outboundRouteName,
		Service: svc.ServiceName,
		Routes:  routes,
		DpTags:  proxy.Dataplane.Spec.TagSet(),
	}

	filterChainBuilder.
		Configure(envoy_listeners.AddFilterChainConfigurer(outboundRouteConfigurer))

	// TODO: https://github.com/kumahq/kuma/issues/3325
	switch svc.Protocol {
	case core_mesh.ProtocolGRPC:
		filterChainBuilder.Configure(envoy_listeners.GrpcStats())
	}
	listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
	listener, err := listenerBuilder.Build()
	if err != nil {
		return nil, err
	}

	resources := core_xds.NewResourceSet().Add(
		&core_xds.Resource{
			Name:           listener.GetName(),
			Origin:         generator.OriginOutbound,
			Resource:       listener,
			ResourceOrigin: svc.Outbound.Resource,
			Protocol:       svc.Protocol,
		})

	return resources, nil
}

func generateListeners(
	proxy *core_xds.Proxy,
	rules rules.ToRules,
	servicesAcc envoy_common.ServicesAccumulator,
	meshCtx xds_context.MeshContext,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of clusters.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[common_api.BackendRefHash]string{}

	for _, svc := range meshroute_xds.CollectServices(proxy, meshCtx) {
		rs, err := generateFromService(
			meshCtx,
			proxy,
			clusterCache,
			servicesAcc,
			rules,
			svc,
		)
		if err != nil {
			return nil, err
		}
		resources.AddSet(rs)
	}

	return resources, nil
}

func ComputeHTTPRouteConf(toRules rules.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) (*api.PolicyDefault, map[string]core_model.ResourceMeta) {
	// compute for old MeshService
	var conf *api.PolicyDefault
	backendRefOrigin := map[string]core_model.ResourceMeta{}

	ruleHTTP := toRules.Rules.Compute(core_rules.MeshService(svc.ServiceName))
	if ruleHTTP != nil {
		conf = pointer.To(ruleHTTP.Conf.(api.PolicyDefault))
		for hash := range ruleHTTP.BackendRefOriginIndex {
			if origin, ok := ruleHTTP.GetBackendRefOrigin(hash); ok {
				backendRefOrigin[string(hash)] = origin
			}
		}
	}
	// check if there is configuration for real MeshService and prioritize it
	if svc.Outbound.Resource != nil {
		resourceConf := toRules.ResourceRules.Compute(*svc.Outbound.Resource, meshCtx.Resources)
		if resourceConf != nil && len(resourceConf.Conf) != 0 {
			conf = pointer.To(resourceConf.Conf[0].(api.PolicyDefault))
			for hash := range resourceConf.BackendRefOriginIndex {
				if origin, ok := resourceConf.GetBackendRefOrigin(hash); ok {
					backendRefOrigin[string(hash)] = origin
				}
			}
		}
	}
	return conf, backendRefOrigin
}

// prepareRoutes handles the always present, catch all default route
func prepareRoutes(toRules rules.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) []api.Route {
	conf, backendRefToOrigin := ComputeHTTPRouteConf(toRules, svc, meshCtx)

	var apiRules []api.Rule
	if conf != nil {
		apiRules = conf.Rules
	}

	if len(apiRules) == 0 {
		switch svc.Protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		default:
			return nil
		}
	}

	// sort rules before we add default prefix matches etc
	routes := api.SortRules(apiRules, backendRefToOrigin, meshCtx.ResolveResourceIdentifier)

	catchAllPathMatch := api.PathMatch{Value: "/", Type: api.PathPrefix}
	catchAllMatch := api.Match{
		Path: pointer.To(catchAllPathMatch),
	}

	noCatchAll := slices.IndexFunc(routes, func(route api.Route) bool {
		return reflect.DeepEqual(route.Match, catchAllMatch)
	}) == -1

	if noCatchAll {
		routes = append(routes, api.Route{
			Match: catchAllMatch,
			Hash:  api.HashMatches([]api.Match{catchAllMatch}),
		})
	}

	for i := range routes {
		route := &routes[i]
		if route.Match.Path == nil {
			// According to Envoy docs, match must have precisely one of
			// prefix, path, safe_regex, connect_matcher,
			// path_separated_prefix, path_match_policy set, so when policy
			// doesn't specify explicit type of matching, we are assuming
			// "catch all" path (any path starting with "/").
			route.Match.Path = pointer.To(catchAllPathMatch)
		}

		if len(route.BackendRefs) == 0 {
			route.BackendRefs = []core_model.ResolvedBackendRef{
				*svc.DefaultBackendRef(),
			}
		}
	}

	return routes
}
