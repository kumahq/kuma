package v1alpha1

import (
	"fmt"
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
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func GenerateOutboundListener(
	proxy *core_xds.Proxy,
	svc meshroute_xds.DestinationService,
	routes []xds.OutboundRoute,
	originDPPTags mesh_proto.MultiValueTagSet,
) (*core_xds.Resource, error) {
	unifiedNamingEnabled := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)
	transparentProxyEnabled := !proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) && proxy.GetTransparentProxy().Enabled()

	address := svc.Outbound.GetAddressWithFallback("127.0.0.1")
	port := svc.Outbound.GetPort()

	legacyRouteConfigName := envoy_names.GetOutboundRouteName(svc.KumaServiceTagValue)
	legacyListenerName := envoy_names.GetOutboundListenerName(address, port)

	routeConfigName := svc.ConditionallyResolveKRIWithFallback(true, legacyRouteConfigName)
	virtualHostName := svc.ConditionallyResolveKRIWithFallback(unifiedNamingEnabled, svc.KumaServiceTagValue)
	listenerStatPrefix := svc.ConditionallyResolveKRIWithFallback(unifiedNamingEnabled, "")
	listenerName := svc.ConditionallyResolveKRIWithFallback(unifiedNamingEnabled, legacyListenerName)

	route := &xds.HttpOutboundRouteConfigurer{
		RouteConfigName: routeConfigName,
		VirtualHostName: virtualHostName,
		Routes:          routes,
		DpTags:          originDPPTags,
	}

	hcm := &envoy_listeners_v3.HttpConnectionManagerConfigurer{
		StatsName:                virtualHostName,
		ForwardClientCertDetails: false,
		NormalizePath:            true,
		InternalAddresses:        proxy.InternalAddresses,
	}

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.AddFilterChainConfigurer(hcm)).
		Configure(envoy_listeners.AddFilterChainConfigurer(route)).
		ConfigureIf(svc.Protocol == core_mesh.ProtocolGRPC, envoy_listeners.GrpcStats()) // TODO: https://github.com/kumahq/kuma/issues/3325

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.StatPrefix(listenerStatPrefix)).
		Configure(envoy_listeners.OutboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.TransparentProxying(transparentProxyEnabled)).
		Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(svc.Outbound.TagsOrNil()).WithoutTags(mesh_proto.MeshTag))).
		Configure(envoy_listeners.FilterChain(filterChain))

	resource, err := listener.Build()
	if err != nil {
		return nil, err
	}

	return &core_xds.Resource{
		Name:           resource.GetName(),
		Origin:         generator.OriginOutbound,
		Resource:       resource,
		ResourceOrigin: svc.Outbound.Resource,
		Protocol:       svc.Protocol,
	}, nil
}

func generateFromService(
	meshCtx xds_context.MeshContext,
	proxy *core_xds.Proxy,
	clusterCache map[common_api.BackendRefHash]string,
	servicesAcc envoy_common.ServicesAccumulator,
	rules rules.ToRules,
	svc meshroute_xds.DestinationService,
) (*core_xds.ResourceSet, error) {
	var routes []xds.OutboundRoute

	unifiedNaming := proxy.Metadata.HasFeature(xds_types.FeatureUnifiedResourceNaming)

	for _, route := range prepareRoutes(rules, svc, meshCtx, unifiedNaming) {
		split := meshroute_xds.MakeHTTPSplit(clusterCache, servicesAcc, route.BackendRefs, meshCtx, unifiedNaming)
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
					unifiedNaming,
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

	var dpTags mesh_proto.MultiValueTagSet
	if meshCtx.IsXKumaTagsUsed() {
		dpTags = proxy.Dataplane.Spec.TagSet()
	}

	listener, err := GenerateOutboundListener(proxy, svc, routes, dpTags)
	if err != nil {
		return nil, err
	}
	return core_xds.NewResourceSet().Add(listener), nil
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

func ComputeHTTPRouteConf(
	toRules rules.ToRules,
	svc meshroute_xds.DestinationService,
	meshCtx xds_context.MeshContext,
) (*api.PolicyDefault, map[common_api.MatchesHash]common.Origin) {
	// check if there is configuration for real MeshService and prioritize it
	if r, ok := svc.Outbound.AssociatedServiceResource(); ok {
		if rule := toRules.ResourceRules.Compute(r, meshCtx.Resources); rule != nil && len(rule.Conf) > 0 {
			return pointer.To(rule.Conf[0].(api.PolicyDefault)), rule.OriginByMatches
		}
	}

	// compute for old MeshService
	if rule := toRules.Rules.Compute(subsetutils.KumaServiceTagElement(svc.KumaServiceTagValue)); rule != nil {
		return pointer.To(rule.Conf.(api.PolicyDefault)), util_maps.MapValues(
			rule.OriginByMatches,
			func(_ common_api.MatchesHash, o core_model.ResourceMeta) common.Origin {
				return common.Origin{Resource: o}
			},
		)
	}

	return nil, make(map[common_api.MatchesHash]common.Origin)
}

// prepareRoutes handles the always present, catch all default route
func prepareRoutes(
	toRules rules.ToRules,
	svc meshroute_xds.DestinationService,
	meshCtx xds_context.MeshContext,
	unifiedNaming bool,
) []api.Route {
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

	var routes []api.Route

	for _, rule := range apiRules {
		filters := pointer.Deref(rule.Default.Filters)
		backendRefs := pointer.Deref(rule.Default.BackendRefs)
		matchesHash := api.HashMatches(rule.Matches)
		routeName := string(matchesHash)
		origin := originByMatches[matchesHash]

		originID := kri.FromResourceMeta(origin.Resource, api.MeshHTTPRouteType, "")
		if unifiedNaming {
			originID = kri.WithSectionName(originID, fmt.Sprintf("rule_%d", origin.RuleIndex))
		}

		if _, ok := svc.Outbound.AssociatedServiceResource(); ok {
			routeName = originID.String()
		}

		for _, match := range rule.Matches {
			var refs []resolve.ResolvedBackendRef

			for _, br := range backendRefs {
				if rbr, ok := resolve.BackendRef(&originID, br, meshCtx.ResolveResourceIdentifier); ok {
					refs = append(refs, rbr)
				}
			}

			routes = append(
				routes,
				api.Route{
					Name:        routeName,
					Match:       match,
					Filters:     filters,
					BackendRefs: refs,
				},
			)
		}
	}

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
