package v1alpha1

import (
	"fmt"
	"reflect"

	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	xds "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
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
	splitCounter := &splitCounter{}
	// ClusterCache (cluster hash -> cluster name) protects us from creating excessive amount of caches.
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
			Configure(envoy_listeners.HttpConnectionManager(serviceName, false))

		var routes []xds.OutboundRoute
		for _, route := range prepareRoutes(rules, serviceName) {
			clusters := makeClusters(proxy, clusterCache, splitCounter, route.BackendRefs)
			servicesAcc.Add(clusters...)

			routes = append(routes, xds.OutboundRoute{
				Matches:  route.Matches,
				Filters:  route.Filters,
				Clusters: clusters,
			})
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
) []Route {
	var rules []api.Rule

	for _, toRule := range toRules {
		if toRule.Subset.IsSubset(core_xds.MeshService(serviceName)) {
			rules = toRule.Rules
		}
	}

	catchAllMatch := []api.Match{{
		Path: api.PathMatch{
			Prefix: "/",
		},
	}}

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
		route := Route{
			Matches: rule.Matches,
		}

		if rule.Default.BackendRefs != nil {
			route.BackendRefs = *rule.Default.BackendRefs
		} else {
			route.BackendRefs = []api.BackendRef{{
				TargetRef: common_api.TargetRef{
					Kind: common_api.MeshService,
					Name: serviceName,
				},
				Weight: 100,
			}}
		}
		if rule.Default.Filters != nil {
			route.Filters = *rule.Default.Filters
		}
		routes = append(routes, route)
	}

	return routes
}

// Whenever `split` is specified in the TrafficRoute which has more than kuma.io/service tag
// We generate a separate Envoy cluster with _X_ suffix. SplitCounter ensures that we have different X for every split in one Dataplane
// Each split is distinct for the whole Dataplane so we can avoid accidental cluster overrides.
type splitCounter struct {
	counter int
}

func (s *splitCounter) getAndIncrement() int {
	counter := s.counter
	s.counter++
	return counter
}

func makeClusters(
	proxy *core_xds.Proxy,
	clusterCache map[string]string,
	splitCounter *splitCounter,
	refs []api.BackendRef,
) []envoy.Cluster {
	var clusters []envoy.Cluster

	for _, ref := range refs {
		switch ref.Kind {
		case common_api.MeshService, common_api.MeshServiceSubset:
		default:
			continue
		}

		service := ref.Name
		if ref.Weight == 0 {
			continue
		}

		name := service

		if len(ref.Tags) > 0 {
			name = envoy_names.GetSplitClusterName(service, splitCounter.getAndIncrement())
		}

		// The mesh tag is present here if this destination is generated
		// from a cross-mesh MeshGateway listener virtual outbound.
		// It is not part of the service tags.
		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			// The name should be distinct to the service & mesh combination
			name = fmt.Sprintf("%s_%s", name, mesh)
		}

		// We assume that all the targets are either ExternalServices or not
		// therefore we check only the first one
		var isExternalService bool
		if endpoints := proxy.Routing.OutboundTargets[service]; len(endpoints) > 0 {
			isExternalService = endpoints[0].IsExternalService()
		}
		if endpoints := proxy.Routing.ExternalServiceOutboundTargets[service]; len(endpoints) > 0 {
			isExternalService = true
		}

		allTags := envoy_tags.Tags(ref.Tags).WithTags(mesh_proto.ServiceTag, ref.Name)
		cluster := envoy_common.NewCluster(
			envoy_common.WithService(service),
			envoy_common.WithName(name),
			envoy_common.WithWeight(uint32(ref.Weight)),
			envoy_common.WithTags(allTags.WithoutTags(mesh_proto.MeshTag)),
			envoy_common.WithExternalService(isExternalService),
		)

		if mesh, ok := ref.Tags[mesh_proto.MeshTag]; ok {
			cluster.SetMesh(mesh)
		}

		if name, ok := clusterCache[allTags.String()]; ok {
			cluster.SetName(name)
		} else {
			clusterCache[allTags.String()] = cluster.Name()
		}

		clusters = append(clusters, cluster)
	}

	return clusters
}
