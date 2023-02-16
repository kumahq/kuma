package generator

import (
	"reflect"
	"sort"

	"golang.org/x/exp/slices"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	meshhttproute_api "github.com/kumahq/kuma/pkg/plugins/policies/meshhttproute/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

const (
	IngressProxy = "ingress-proxy"

	// OriginIngress is a marker to indicate by which ProxyGenerator resources were generated.
	OriginIngress = "ingress"
)

type IngressGenerator struct {
}

func (i IngressGenerator) Generate(ctx xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	destinationsPerService := i.destinations(proxy.ZoneIngressProxy)

	listener, err := i.generateLDS(proxy, proxy.ZoneIngress, destinationsPerService, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginIngress,
		Resource: listener,
	})

	services := i.services(proxy)

	cdsResources, err := i.generateCDS(services, destinationsPerService, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.Add(cdsResources...)

	edsResources, err := i.generateEDS(proxy, services, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.Add(edsResources...)

	return resources, nil
}

// generateLDS generates one Ingress Listener
// Ingress Listener assumes that mTLS is on. Using TLSInspector we sniff SNI value.
// SNI value has service name and tag values specified with the following format: "backend{cluster=2,version=1}"
// We take all possible destinations from TrafficRoutes + GatewayRoutes and generate FilterChainsMatcher for each unique destination.
// This approach has a limitation: additional tags on outbound in Universal mode won't work across different zones.
// Traffic is NOT decrypted here, therefore we don't need certificates and mTLS settings
func (i IngressGenerator) generateLDS(
	proxy *core_xds.Proxy,
	ingress *core_mesh.ZoneIngressResource,
	destinationsPerService map[string][]tags.Tags,
	apiVersion core_xds.APIVersion,
) (envoy_common.NamedResource, error) {
	inboundListenerName := envoy_names.GetInboundListenerName(proxy.ZoneIngress.Spec.GetNetworking().GetAddress(), proxy.ZoneIngress.Spec.GetNetworking().GetPort())
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), ingress.Spec.GetNetworking().GetPort(), core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.TLSInspector())

	if len(proxy.ZoneIngress.Spec.AvailableServices) == 0 {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion)))
	}

	sniUsed := map[string]bool{}

	for _, inbound := range proxy.ZoneIngress.Spec.GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		destinations := destinationsPerService[service]
		destinations = append(destinations, destinationsPerService[mesh_proto.MatchAllTag]...)

		for _, destination := range destinations {
			meshDestination := destination.
				WithTags(mesh_proto.ServiceTag, service).
				WithTags("mesh", inbound.GetMesh())
			sni := tls.SNIFromTags(meshDestination)
			if sniUsed[sni] {
				continue
			}
			sniUsed[sni] = true
			inboundListenerBuilder = inboundListenerBuilder.Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(apiVersion).Configure(
					envoy_listeners.MatchTransportProtocol("tls"),
					envoy_listeners.MatchServerNames(sni),
					envoy_listeners.TcpProxyWithMetadata(service, envoy_common.NewCluster(
						envoy_common.WithService(service),
						envoy_common.WithTags(meshDestination.WithoutTags(mesh_proto.ServiceTag)),
					)),
				),
			))
		}
	}

	return inboundListenerBuilder.Build()
}

func tagsFromTargetRef(targetRef common_api.TargetRef) (tags.Tags, bool) {
	var service string
	var tags tags.Tags

	switch targetRef.Kind {
	case common_api.MeshService:
		service = targetRef.Name
	case common_api.MeshServiceSubset:
		service = targetRef.Name
		tags = targetRef.Tags
	case common_api.Mesh:
		service = mesh_proto.MatchAllTag
	case common_api.MeshSubset:
		service = mesh_proto.MatchAllTag
		tags = targetRef.Tags
	default:
		return nil, false
	}

	return tags.WithTags(mesh_proto.ServiceTag, service), true
}

func (_ IngressGenerator) destinations(
	ingressProxy *core_xds.ZoneIngressProxy,
) map[string][]tags.Tags {
	destinations := map[string][]tags.Tags{}
	for _, tr := range ingressProxy.PolicyResources[core_mesh.TrafficRouteType].(*core_mesh.TrafficRouteResourceList).Items {
		for _, split := range tr.Spec.Conf.GetSplitWithDestination() {
			service := split.Destination[mesh_proto.ServiceTag]
			destinations[service] = append(destinations[service], split.Destination)
		}
		for _, http := range tr.Spec.Conf.Http {
			for _, split := range http.GetSplitWithDestination() {
				service := split.Destination[mesh_proto.ServiceTag]
				destinations[service] = append(destinations[service], split.Destination)
			}
		}
	}

	if len(ingressProxy.PolicyResources[meshhttproute_api.MeshHTTPRouteType].GetItems()) > 0 {
		// We need to add a destination to route any service to any instance of
		// that service
		matchAllTags := tags.Tags{
			mesh_proto.ServiceTag: mesh_proto.MatchAllTag,
		}
		matchAllDestinations := destinations[mesh_proto.MatchAllTag]
		foundAllServicesDestination := slices.ContainsFunc(matchAllDestinations, func(tagsElem tags.Tags) bool {
			return reflect.DeepEqual(tagsElem, matchAllTags)
		})
		if !foundAllServicesDestination {
			matchAllDestinations = append(matchAllDestinations, matchAllTags)
		}
		destinations[mesh_proto.MatchAllTag] = matchAllDestinations
	}

	// Note that we're not merging these resources, but that's OK because the
	// set of destinations after merging is a subset of the set we get here by
	// iterating through them.
	for _, route := range ingressProxy.PolicyResources[meshhttproute_api.MeshHTTPRouteType].(*meshhttproute_api.MeshHTTPRouteResourceList).Items {
		for _, to := range route.Spec.To {
			toTags, ok := tagsFromTargetRef(to.TargetRef)
			if !ok {
				continue
			}

			service := toTags[mesh_proto.ServiceTag]

			for _, rule := range to.Rules {
				if rule.Default.BackendRefs == nil {
					destinations[service] = append(destinations[service], toTags)
				}
				for _, backendRef := range pointer.Deref(rule.Default.BackendRefs) {
					backendTags, ok := tagsFromTargetRef(backendRef.TargetRef)
					if !ok {
						continue
					}
					destinations[service] = append(destinations[service], backendTags)
				}
			}
		}
	}

	var backends []*mesh_proto.MeshGatewayRoute_Backend

	for _, route := range ingressProxy.GatewayRoutes.Items {
		for _, rule := range route.Spec.GetConf().GetHttp().GetRules() {
			backends = append(backends, rule.Backends...)
		}
		for _, rule := range route.Spec.GetConf().GetTcp().GetRules() {
			backends = append(backends, rule.Backends...)
		}
	}

	for _, backend := range backends {
		service := backend.Destination[mesh_proto.ServiceTag]
		destinations[service] = append(destinations[service], backend.Destination)
	}

	for _, gateway := range ingressProxy.MeshGateways.Items {
		for _, selector := range gateway.Selectors() {
			service := selector.GetMatch()[mesh_proto.ServiceTag]
			for _, listener := range gateway.Spec.GetConf().GetListeners() {
				if !listener.CrossMesh {
					continue
				}
				destinations[service] = append(
					destinations[service],
					tags.Tags(mesh_proto.Merge(selector.GetMatch(), gateway.Spec.GetTags(), listener.GetTags())),
				)
			}
		}
	}

	return destinations
}

func (_ IngressGenerator) services(proxy *core_xds.Proxy) []string {
	var services []string
	for service := range proxy.Routing.OutboundTargets {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func (i IngressGenerator) generateCDS(
	services []string,
	destinationsPerService map[string][]tags.Tags,
	apiVersion core_xds.APIVersion,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		tagSlice := tags.TagsSlice(append(destinationsPerService[service], destinationsPerService[mesh_proto.MatchAllTag]...))
		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(tags.Without(mesh_proto.ServiceTag), tags.With("mesh"))
		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.EdsCluster(service)).
			Configure(envoy_clusters.LbSubset(tagKeySlice)).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     service,
			Origin:   OriginIngress,
			Resource: edsCluster,
		})
	}
	return resources, nil
}

func (_ IngressGenerator) generateEDS(
	proxy *core_xds.Proxy,
	services []string,
	apiVersion core_xds.APIVersion,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		endpoints := proxy.Routing.OutboundTargets[service]
		cla, err := envoy_endpoints.CreateClusterLoadAssignment(service, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     service,
			Origin:   OriginIngress,
			Resource: cla,
		})
	}
	return resources, nil
}
