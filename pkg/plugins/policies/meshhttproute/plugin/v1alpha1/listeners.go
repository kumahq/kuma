package v1alpha1

import (
	"fmt"
	"slices"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/kri"
	core_meta "github.com/kumahq/kuma/v3/pkg/core/metadata"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	xds_types "github.com/kumahq/kuma/v3/pkg/core/xds/types"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/common"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/rules/resolve"
	meshroute_xds "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/v3/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/v3/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/v3/pkg/xds/generator/metadata"
)

func GenerateOutboundListener(
	proxy *core_xds.Proxy,
	svc meshroute_xds.DestinationService,
	routes []xds.OutboundRoute,
	originDPPTags mesh_proto.MultiValueTagSet,
) (*core_xds.Resource, error) {
	transparentProxyEnabled := !proxy.Metadata.HasFeature(xds_types.FeatureBindOutbounds) && proxy.GetTransparentProxy().Enabled()

	address := svc.Outbound.GetAddressWithFallback("127.0.0.1")
	port := svc.Outbound.GetPort()

	legacyRouteConfigName := envoy_names.GetOutboundRouteName(svc.KumaServiceTagValue)
	legacyListenerName := envoy_names.GetOutboundListenerName(address, port)

	routeConfigName := legacyRouteConfigName
	virtualHostName := svc.KumaServiceTagValue
	listenerStatPrefix := ""
	listenerName := legacyListenerName
	if svc.DestinationResource != "" {
		routeConfigName = svc.DestinationResource
		virtualHostName = svc.DestinationResource
		listenerStatPrefix = svc.DestinationResource
		listenerName = svc.DestinationResource
	}

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
		IPv6Enabled:              proxy.Metadata.GetIPv6Enabled(),
	}

	filterChain := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
		Configure(envoy_listeners.AddFilterChainConfigurer(hcm)).
		Configure(envoy_listeners.AddFilterChainConfigurer(route)).
		ConfigureIf(svc.Protocol == core_meta.ProtocolGRPC, envoy_listeners.GrpcStats()) // TODO: https://github.com/kumahq/kuma/issues/3325

	listener := envoy_listeners.NewListenerBuilder(proxy.APIVersion, listenerName).
		Configure(envoy_listeners.StatPrefix(listenerStatPrefix)).
		Configure(envoy_listeners.OutboundListener(address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.TransparentProxying(transparentProxyEnabled)).
		Configure(envoy_listeners.TagsMetadata(svc.OutboundListenerTags())).
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
	servicesAcc envoy_common.ServicesAccumulator,
	rules rules.ToRules,
	svc meshroute_xds.DestinationService,
) (*core_xds.ResourceSet, error) {
	var routes []xds.OutboundRoute

	for _, route := range prepareRoutes(rules, svc, meshCtx) {
		var split []xds.BackendRefSplit
		for _, backendRef := range route.BackendRefs {
			backendSplit := meshroute_xds.MakeHTTPSplit(
				clusterCache,
				servicesAcc,
				[]resolve.ResolvedBackendRef{backendRef.BackendRef},
				meshCtx,
			)
			if len(backendSplit) == 0 {
				continue
			}
			split = append(split, xds.BackendRefSplit{
				Split:   backendSplit[0],
				Filters: backendRef.Filters,
			})
		}
		if len(split) == 0 && !route.AllBackendRefsUnresolved && route.DirectResponseStatus == 0 {
			continue
		}
		// mirrored requests go to a cluster of their own, it has no weight so
		// keep the split only to know which cluster was created for the mirror
		mirrorSplits := map[int]envoy_common.Split{}
		for i := range route.Filters {
			ref, ok := route.MirrorBackendRefs[i]
			if !ok {
				continue
			}
			mirrorSplit := meshroute_xds.MakeHTTPSplit(
				clusterCache, servicesAcc,
				[]resolve.ResolvedBackendRef{ref},
				meshCtx,
			)
			if len(mirrorSplit) == 0 {
				continue
			}
			mirrorSplits[i] = mirrorSplit[0]
		}
		routes = append(routes, xds.OutboundRoute{
			Name:                        route.Name,
			Match:                       route.Match,
			Filters:                     route.Filters,
			DirectResponseStatus:        route.DirectResponseStatus,
			UnresolvedBackendRefsWeight: route.UnresolvedBackendRefsWeight,
			AllBackendRefsUnresolved:    route.AllBackendRefsUnresolved,
			MirrorSplits:                mirrorSplits,
			Split:                       split,
		})
	}

	if len(routes) == 0 {
		return nil, nil
	}

	var dpTags mesh_proto.MultiValueTagSet
	if meshCtx.IsXKumaTagsUsed() {
		data := map[string][]string{}
		for key, value := range proxy.Dataplane.GetMeta().GetLabels() {
			data[key] = []string{value}
		}
		dpTags = mesh_proto.MultiValueTagSetFrom(data)
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
	if r, ok := svc.Outbound.AssociatedServiceResource(); ok {
		if rule := toRules.ResourceRules.Compute(r, meshCtx.Resources); rule != nil && len(rule.Conf) > 0 {
			return pointer.To(rule.Conf[0].(api.PolicyDefault)), rule.OriginByMatches
		}
	}

	return nil, make(map[common_api.MatchesHash]common.Origin)
}

// prepareRoutes handles the always-present default backend route and the
// policy-generated catch-all 404 fallback when rules leave unmatched traffic.
func prepareRoutes(
	toRules rules.ToRules,
	svc meshroute_xds.DestinationService,
	meshCtx xds_context.MeshContext,
) []api.Route {
	conf, originByMatches := ComputeHTTPRouteConf(toRules, svc, meshCtx)

	var apiRules []api.Rule
	if conf != nil {
		apiRules = conf.Rules
	}

	if len(apiRules) == 0 && !core_meta.IsHTTPBased(svc.Protocol) {
		return nil
	}

	var routes []api.Route

	for _, rule := range apiRules {
		filters := pointer.Deref(rule.Default.Filters)
		backendRefs := pointer.Deref(rule.Default.BackendRefs)
		hasExplicitBackendRefs := rule.Default.BackendRefs != nil
		matchesHash := api.HashMatches(rule.Matches)
		routeName := string(matchesHash)
		origin := originByMatches[matchesHash]

		originID := kri.WithSectionName(kri.FromResourceMeta(origin.Resource, api.MeshHTTPRouteType), fmt.Sprintf("rule_%d", origin.RuleIndex))

		if _, ok := svc.Outbound.AssociatedServiceResource(); ok {
			routeName = originID.String()
		}

		mirrorRefs := map[int]resolve.ResolvedBackendRef{}

		for i, filter := range filters {
			if filter.Type != api.RequestMirrorType || filter.RequestMirror == nil {
				continue
			}
			if rbr, ok := resolve.BackendRef(originID, filter.RequestMirror.BackendRef, meshCtx.ResolveResourceIdentifier); ok {
				mirrorRefs[i] = rbr
			}
		}

		for _, match := range rule.Matches {
			var refs []api.RouteBackendRef
			var unresolvedWeight uint

			for _, br := range backendRefs {
				rbr, ok := resolve.BackendRef(originID, br.CommonBackendRef(), meshCtx.ResolveResourceIdentifier)
				if !ok {
					unresolvedWeight += pointer.DerefOr(br.Weight, 1)
					continue
				}
				if !backendRefProducesHTTPSplit(meshCtx, rbr) {
					if rr := rbr.RealResourceBackendRef(); rr != nil {
						unresolvedWeight += rr.Weight
					} else {
						unresolvedWeight += pointer.DerefOr(br.Weight, 1)
					}
					continue
				}
				refs = append(refs, api.RouteBackendRef{
					BackendRef: rbr,
					Filters:    pointer.Deref(br.Filters),
				})
			}

			routes = append(
				routes,
				api.Route{
					Name:                        routeName,
					Origin:                      originID,
					Match:                       match,
					Filters:                     filters,
					BackendRefs:                 refs,
					UnresolvedBackendRefsWeight: unresolvedWeight,
					AllBackendRefsUnresolved:    hasExplicitBackendRefs && len(refs) == 0,
					MirrorBackendRefs:           mirrorRefs,
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
		return matchIsCatchAll(route.Match)
	}) == -1

	if noCatchAll {
		fallbackRoute := api.Route{
			Match:  catchAllMatch,
			Name:   string(api.HashMatches([]api.Match{catchAllMatch})),
			Origin: svc.Outbound.Resource,
		}
		if len(routes) > 0 && core_meta.IsHTTPBased(svc.Protocol) {
			fallbackRoute.DirectResponseStatus = 404
		}
		routes = append(routes, fallbackRoute)
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

		if len(route.BackendRefs) == 0 && !route.AllBackendRefsUnresolved && route.DirectResponseStatus == 0 {
			route.BackendRefs = []api.RouteBackendRef{
				{
					BackendRef: *svc.DefaultBackendRef(),
				},
			}
		}
	}

	return routes
}

func matchIsCatchAll(match api.Match) bool {
	if match.Method != nil || len(pointer.Deref(match.Headers)) > 0 || len(pointer.Deref(match.QueryParams)) > 0 {
		return false
	}
	if match.Path == nil {
		return true
	}
	return match.Path.Type == api.PathPrefix && match.Path.Value == "/"
}

func backendRefProducesHTTPSplit(
	meshCtx xds_context.MeshContext,
	ref resolve.ResolvedBackendRef,
) bool {
	rr := ref.RealResourceBackendRef()
	if rr == nil || rr.Weight == 0 {
		return false
	}

	_, port, ok := meshroute_xds.DestinationPortFromRef(meshCtx, rr)
	return ok && core_meta.IsHTTPBased(port.GetProtocol())
}
