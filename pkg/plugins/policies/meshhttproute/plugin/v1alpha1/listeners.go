package v1alpha1

import (
	"reflect"

	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
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

func generateListeners(
	proxy *core_xds.Proxy,
	rules rules.Rules,
	servicesAcc envoy_common.ServicesAccumulator,
	meshCtx xds_context.MeshContext,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of clusters.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[common_api.TargetRefHash]string{}

	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbound() {
		serviceName := outbound.GetService()
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)

		listenerBuilder := envoy_listeners.NewOutboundListenerBuilder(proxy.APIVersion, oface.DataplaneIP, oface.DataplanePort, core_xds.SocketAddressProtocolTCP).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.GetTags()).WithoutTags(mesh_proto.MeshTag)))

		filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.AddFilterChainConfigurer(&envoy_listeners_v3.HttpConnectionManagerConfigurer{
				StatsName:                serviceName,
				ForwardClientCertDetails: false,
				NormalizePath:            true,
			}))

		protocol := meshCtx.GetServiceProtocol(serviceName)
		var routes []xds.OutboundRoute
		for _, route := range prepareRoutes(rules, serviceName, protocol, outbound.GetTags()) {
			split := meshroute_xds.MakeHTTPSplit(clusterCache, servicesAcc, route.BackendRefs, meshCtx)
			if split == nil {
				continue
			}
			for _, filter := range route.Filters {
				if filter.Type == api.RequestMirrorType {
					// we need to create a split for the mirror backend
					_ = meshroute_xds.MakeHTTPSplit(
						clusterCache, servicesAcc,
						[]common_api.BackendRef{{
							TargetRef: filter.RequestMirror.BackendRef,
							Weight:    pointer.To[uint](1), // any non-zero value
						}},
						meshCtx,
					)
				}
			}
			routes = append(routes, xds.OutboundRoute{
				Matches:                 route.Matches,
				Filters:                 route.Filters,
				Split:                   split,
				BackendRefToClusterName: clusterCache,
			})
		}

		if len(routes) == 0 {
			continue
		}

		outboundRouteConfigurer := &xds.HttpOutboundRouteConfigurer{
			Service: serviceName,
			Routes:  routes,
			DpTags:  proxy.Dataplane.Spec.TagSet(),
		}

		filterChainBuilder.
			Configure(envoy_listeners.AddFilterChainConfigurer(outboundRouteConfigurer))

		// TODO: https://github.com/kumahq/kuma/issues/3325
		switch protocol {
		case core_mesh.ProtocolGRPC:
			filterChainBuilder.Configure(envoy_listeners.GrpcStats())
		}
		listenerBuilder.Configure(envoy_listeners.FilterChain(filterChainBuilder))
		listener, err := listenerBuilder.Build()
		if err != nil {
			return nil, err
		}
		resources.Add(&core_xds.Resource{
			Name:     listener.GetName(),
			Origin:   generator.OriginOutbound,
			Resource: listener,
		})
	}

	return resources, nil
}

// prepareRoutes handles the always present, catch all default route
func prepareRoutes(
	toRules rules.Rules,
	serviceName string,
	protocol core_mesh.Protocol,
	tags map[string]string,
) []Route {
	conf := rules.ComputeConf[api.PolicyDefault](toRules, core_rules.MeshServiceElement(serviceName))

	var apiRules []api.Rule
	if conf != nil {
		apiRules = conf.Rules
	}

	if len(apiRules) == 0 {
		switch protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		default:
			return nil
		}
	}

	catchAllPathMatch := api.PathMatch{Value: "/", Type: api.PathPrefix}
	catchAllMatch := []api.Match{
		{Path: pointer.To(catchAllPathMatch)},
	}

	noCatchAll := slices.IndexFunc(apiRules, func(rule api.Rule) bool {
		return reflect.DeepEqual(rule.Matches, catchAllMatch)
	}) == -1

	if noCatchAll {
		apiRules = append(apiRules, api.Rule{
			Matches: catchAllMatch,
		})
	}

	var routes []Route
	for _, rule := range apiRules {
		var matches []api.Match

		for _, match := range rule.Matches {
			if match.Path == nil {
				// According to Envoy docs, match must have precisely one of
				// prefix, path, safe_regex, connect_matcher,
				// path_separated_prefix, path_match_policy set, so when policy
				// doesn't specify explicit type of matching, we are assuming
				// "catch all" path (any path starting with "/").
				match.Path = pointer.To(catchAllPathMatch)
			}

			matches = append(matches, match)
		}

		route := Route{
			Matches: matches,
		}

		if rule.Default.BackendRefs != nil {
			route.BackendRefs = *rule.Default.BackendRefs
		} else {
			route.BackendRefs = []common_api.BackendRef{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.MeshService,
						Name: serviceName,
						Tags: tags,
					},
					Weight: pointer.To(uint(100)),
				},
			}
		}
		if rule.Default.Filters != nil {
			route.Filters = *rule.Default.Filters
		}
		routes = append(routes, route)
	}

	return routes
}
