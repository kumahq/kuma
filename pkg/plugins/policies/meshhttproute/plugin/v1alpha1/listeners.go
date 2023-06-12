package v1alpha1

import (
	"reflect"

	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	core_rules "github.com/kumahq/kuma/pkg/plugins/policies/core/rules"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	meshroute_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds/meshroute"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_listeners_v3 "github.com/kumahq/kuma/pkg/xds/envoy/listeners/v3"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func generateListeners(
	proxy *core_xds.Proxy,
	rules []ToRouteRule,
	servicesAcc envoy_common.ServicesAccumulator,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()
	splitCounter := &meshroute_xds.SplitCounter{}
	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of clusters.
	// For one outbound we pick one traffic route so LB and Timeout are the same.
	// If we have same split in many HTTP matches we can use the same cluster with different weight
	clusterCache := map[string]string{}

	for _, outbound := range proxy.Dataplane.Spec.GetNetworking().GetOutbound() {
		serviceName := outbound.GetTagsIncludingLegacy()[mesh_proto.ServiceTag]
		oface := proxy.Dataplane.Spec.Networking.ToOutboundInterface(outbound)
		outboundListenerName := envoy_names.GetOutboundListenerName(oface.DataplaneIP, oface.DataplanePort)

		listenerBuilder := envoy_listeners.NewListenerBuilder(proxy.APIVersion).
			Configure(envoy_listeners.OutboundListener(outboundListenerName, oface.DataplaneIP, oface.DataplanePort, core_xds.SocketAddressProtocolTCP)).
			Configure(envoy_listeners.TransparentProxying(proxy.Dataplane.Spec.Networking.GetTransparentProxying())).
			Configure(envoy_listeners.TagsMetadata(envoy_tags.Tags(outbound.GetTagsIncludingLegacy()).WithoutTags(mesh_proto.MeshTag)))

		filterChainBuilder := envoy_listeners.NewFilterChainBuilder(proxy.APIVersion).
			Configure(envoy_listeners.AddFilterChainConfigurer(&envoy_listeners_v3.HttpConnectionManagerConfigurer{
				StatsName:                serviceName,
				ForwardClientCertDetails: false,
				NormalizePath:            true,
			}))

		protocol := plugins_xds.InferProtocol(proxy.Routing, serviceName)
		var routes []xds.OutboundRoute
		for _, route := range prepareRoutes(rules, serviceName, protocol) {
			split := meshroute_xds.MakeHTTPSplit(proxy, clusterCache, splitCounter, servicesAcc, route.BackendRefs)
			if split == nil {
				continue
			}
			for _, filter := range route.Filters {
				if filter.Type == api.RequestMirrorType {
					// we need to create a split for the mirror backend
					_ = meshroute_xds.MakeHTTPSplit(proxy, clusterCache, splitCounter, servicesAcc,
						[]common_api.BackendRef{{
							TargetRef: filter.RequestMirror.BackendRef,
							Weight:    pointer.To[uint](1), // any non-zero value
						}})
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
	toRules []ToRouteRule,
	serviceName string,
	protocol core_mesh.Protocol,
) []Route {
	var rules []api.Rule

	for _, toRule := range toRules {
		if toRule.Subset.IsSubset(core_rules.MeshService(serviceName)) {
			rules = toRule.Rules
		}
	}

	if len(rules) == 0 {
		switch protocol {
		case core_mesh.ProtocolHTTP, core_mesh.ProtocolHTTP2, core_mesh.ProtocolGRPC:
		default:
			return nil
		}
	}

	catchAllPathMatch := api.PathMatch{Value: "/", Type: api.PathPrefix}
	catchAllMatch := []api.Match{{Path: pointer.To(catchAllPathMatch)}}

	noCatchAll := slices.IndexFunc(rules, func(rule api.Rule) bool {
		return reflect.DeepEqual(rule.Matches, catchAllMatch)
	}) == -1

	if noCatchAll {
		rules = append(rules, api.Rule{
			Matches: catchAllMatch,
		})
	}

	var routes []Route
	for _, rule := range rules {
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
			route.BackendRefs = []common_api.BackendRef{{
				TargetRef: common_api.TargetRef{
					Kind: common_api.MeshService,
					Name: serviceName,
				},
				Weight: pointer.To(uint(100)),
			}}
		}
		if rule.Default.Filters != nil {
			route.Filters = *rule.Default.Filters
		}
		routes = append(routes, route)
	}

	return routes
}
