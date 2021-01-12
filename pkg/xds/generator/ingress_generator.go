package generator

import (
	"sort"

	"github.com/kumahq/kuma/pkg/xds/envoy/tls"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	model "github.com/kumahq/kuma/pkg/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
)

const (
	IngressProxy = "ingress-proxy"

	// OriginIngress is a marker to indicate by which ProxyGenerator resources were generated.
	OriginIngress = "ingress"
)

type IngressGenerator struct {
}

func (i IngressGenerator) Generate(ctx xds_context.Context, proxy *model.Proxy) (*model.ResourceSet, error) {
	resources := model.NewResourceSet()

	destinationsPerService := i.destinations(proxy.Routing.TrafficRouteList)

	listener, err := i.generateLDS(proxy.Dataplane, destinationsPerService, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.Add(&model.Resource{
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
// We take all possible destinations from TrafficRoutes and generate FilterChainsMatcher for each unique destination.
// This approach has a limitation: additional tags on outbound in Universal mode won't work across different zones.
// Traffic is NOT decrypted here, therefore we don't need certificates and mTLS settings
func (i IngressGenerator) generateLDS(
	ingress *core_mesh.DataplaneResource,
	destinationsPerService map[string][]envoy_common.Tags,
	apiVersion envoy_common.APIVersion,
) (envoy_common.NamedResource, error) {
	inbound := ingress.Spec.Networking.Inbound[0]
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), inbound.Port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), inbound.Port)).
		Configure(envoy_listeners.TLSInspector())

	if !ingress.Spec.HasAvailableServices() {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion)))
	}

	sniUsed := map[string]bool{}

	for _, inbound := range ingress.Spec.GetNetworking().GetIngress().GetAvailableServices() {
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
			inboundListenerBuilder = inboundListenerBuilder.
				Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion).
					Configure(envoy_listeners.FilterChainMatch(sni)).
					Configure(envoy_listeners.TcpProxy(service, envoy_common.ClusterSubset{
						ClusterName: service,
						Tags:        meshDestination.WithoutTag(mesh_proto.ServiceTag),
					}))))
		}
	}

	return inboundListenerBuilder.Build()
}

func (_ IngressGenerator) destinations(trs *core_mesh.TrafficRouteResourceList) map[string][]envoy_common.Tags {
	destinations := map[string][]envoy_common.Tags{}
	for _, tr := range trs.Items {
		for _, split := range tr.Spec.Conf.Split {
			service := split.Destination[mesh_proto.ServiceTag]
			destinations[service] = append(destinations[service], split.Destination)
		}
	}
	return destinations
}

func (_ IngressGenerator) services(proxy *model.Proxy) []string {
	var services []string
	for service := range proxy.Routing.OutboundTargets {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func (i IngressGenerator) generateCDS(
	services []string,
	destinationsPerService map[string][]envoy_common.Tags,
	apiVersion envoy_common.APIVersion,
) (resources []*model.Resource, _ error) {
	for _, service := range services {
		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.EdsCluster(service)).
			Configure(envoy_clusters.LbSubset(i.lbSubsets(service, destinationsPerService))).
			Build()
		if err != nil {
			return nil, err
		}
		resources = append(resources, &model.Resource{
			Name:     service,
			Origin:   OriginIngress,
			Resource: edsCluster,
		})
	}
	return
}
func (_ IngressGenerator) lbSubsets(service string, destinationsPerService map[string][]envoy_common.Tags) [][]string {
	selectors := [][]string{}
	destinations := destinationsPerService[service]
	destinations = append(destinations, destinationsPerService[mesh_proto.MatchAllTag]...)

	for _, destination := range destinations {
		keys := append(destination.WithoutTag(mesh_proto.ServiceTag).Keys(), "mesh")
		selectors = append(selectors, keys)
	}
	return selectors
}

func (_ IngressGenerator) generateEDS(
	proxy *model.Proxy,
	services []string,
	apiVersion envoy_common.APIVersion,
) (resources []*model.Resource, err error) {
	for _, service := range services {
		endpoints := proxy.Routing.OutboundTargets[service]
		cla, err := envoy_endpoints.CreateClusterLoadAssignment(service, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &model.Resource{
			Name:     service,
			Origin:   OriginIngress,
			Resource: cla,
		})
	}
	return
}
