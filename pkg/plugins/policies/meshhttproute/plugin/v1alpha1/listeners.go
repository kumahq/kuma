package v1alpha1

import (
	"reflect"
	"slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/pkg/core/xds/types"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/resolve"
	"github.com/kumahq/kuma/pkg/plugins/policies/core/rules/subsetutils"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/util/pointer"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
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
	listenerBuilder := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, svc.Outbound.GetAddress(), svc.Outbound.GetPort(), core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag)))

	if !proxy.Metadata.HasFeature(xds_types.FeatureDynamicLoopbackOutbounds) {
		listenerBuilder.Configure(envoy_listeners.TransparentProxying(proxy))
	}

	resourceName := svc.ServiceName

	filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.AddFilterChainConfigurer(&envoy_listeners_v3.HttpConnectionManagerConfigurer{
			StatsName:                resourceName,
			ForwardClientCertDetails: false,
			NormalizePath:            true,
			InternalAddresses:        proxy.InternalAddresses,
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
					[]resolve.ResolvedBackendRef{*resolve.NewResolvedBackendRef(pointer.To(resolve.LegacyBackendRef(filter.RequestMirror.BackendRef)))},
					meshCtx,
				)
			}
		}
		routes = append(routes, xds.OutboundRoute{
			Name:                    route.Name,
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
	if r, ok := svc.Outbound.AssociatedServiceResource(); ok {
		outboundRouteName = r.String()
	}
	var dpTags mesh_proto.MultiValueTagSet
	if meshCtx.IsXKumaTagsUsed() {
		dpTags = proxy.Dataplane.Spec.TagSet()
	}
	outboundRouteConfigurer := &xds.HttpOutboundRouteConfigurer{
		Name:    outboundRouteName,
		Service: svc.ServiceName,
		Routes:  routes,
		DpTags:  dpTags,
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

func ComputeHTTPRouteConf(toRules rules.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) (*api.PolicyDefault, map[common_api.MatchesHash]core_model.ResourceMeta) {
	// compute for old MeshService
	var conf *api.PolicyDefault
	originByMatches := map[common_api.MatchesHash]core_model.ResourceMeta{}

	ruleHTTP := toRules.Rules.Compute(subsetutils.MeshServiceElement(svc.ServiceName))
	if ruleHTTP != nil {
		conf = pointer.To(ruleHTTP.Conf.(api.PolicyDefault))
		originByMatches = ruleHTTP.OriginByMatches
	}
	// check if there is configuration for real MeshService and prioritize it
	if r, ok := svc.Outbound.AssociatedServiceResource(); ok {
		resourceConf := toRules.ResourceRules.Compute(r, meshCtx.Resources)
		if resourceConf != nil && len(resourceConf.Conf) != 0 {
			conf = pointer.To(resourceConf.Conf[0].(api.PolicyDefault))
			originByMatches = util_maps.MapValues(resourceConf.OriginByMatches, func(o common.Origin) core_model.ResourceMeta {
				return o.Resource
			})
		}
	}
	return conf, originByMatches
}

// prepareRoutes handles the always present, catch all default route
func prepareRoutes(toRules rules.ToRules, svc meshroute_xds.DestinationService, meshCtx xds_context.MeshContext) []api.Route {
	conf, originByMatches := ComputeHTTPRouteConf(toRules, svc, meshCtx)

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

	getOrigin := func(ms []api.Match) core_model.ResourceMeta {
		return originByMatches[api.HashMatches(ms)]
	}

	getRouteName := func(ms []api.Match) string {
		if _, ok := svc.Outbound.AssociatedServiceResource(); ok {
			return kri.FromResourceMeta(getOrigin(ms), api.MeshHTTPRouteType, "").String()
		}
		return string(api.HashMatches(ms))
	}

	routes := util_slices.FlatMap(apiRules, func(rule api.Rule) []api.Route {
		var routes []api.Route
		for _, match := range rule.Matches {
			routes = append(routes, api.Route{
				Name:    getRouteName(rule.Matches),
				Match:   match,
				Filters: pointer.Deref(rule.Default.Filters),
				BackendRefs: util_slices.FilterMap(pointer.Deref(rule.Default.BackendRefs), func(br common_api.BackendRef) (resolve.ResolvedBackendRef, bool) {
					return resolve.BackendRef(getOrigin(rule.Matches), br, meshCtx.ResolveResourceIdentifier)
				}),
			})
		}
		return routes
	})

	// sort rules before we add default prefix matches etc
	slices.SortStableFunc(routes, func(i, j api.Route) int {
		return api.CompareMatch(i.Match, j.Match)
	})

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
			Name:  string(api.HashMatches([]api.Match{catchAllMatch})),
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
			route.BackendRefs = []resolve.ResolvedBackendRef{
				*svc.DefaultBackendRef(),
			}
		}
	}

	return routes
}
