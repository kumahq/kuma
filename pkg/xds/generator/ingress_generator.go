package generator

import (
	"fmt"
	"sort"
	"strings"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"

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

	destinationsPerService := i.destinations(proxy.TrafficRouteList)

	listener, err := i.generateLDS(proxy.Dataplane, destinationsPerService)
	if err != nil {
		return nil, err
	}
	resources.Add(&model.Resource{
		Name:     listener.Name,
		Origin:   OriginIngress,
		Resource: listener,
	})

	services := i.services(proxy)

	cdsResources, err := i.generateCDS(services, destinationsPerService)
	if err != nil {
		return nil, err
	}
	resources.Add(cdsResources...)

	edsResources := i.generateEDS(proxy, services)
	resources.Add(edsResources...)

	return resources, nil
}

// generateLDS generates one Ingress Listener
// Ingress Listener assumes that mTLS is on. Using TLSInspector we sniff SNI value.
// SNI value has service name and tag values specified with the following format: "backend{cluster=2,version=1}"
// We take all possible destinations from TrafficRoutes and generate FilterChainsMatcher for each unique destination.
// Traffic is NOT decrypted here, therefore we don't need certificates and mTLS settings
func (i IngressGenerator) generateLDS(ingress *core_mesh.DataplaneResource, destinationsPerService map[string][]envoy_common.Tags) (*envoy_api_v2.Listener, error) {
	inbound := ingress.Spec.Networking.Inbound[0]
	inboundListenerName := envoy_names.GetInboundListenerName(ingress.Spec.GetNetworking().GetAddress(), inbound.Port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder().
		Configure(envoy_listeners.InboundListener(inboundListenerName, ingress.Spec.GetNetworking().GetAddress(), inbound.Port)).
		Configure(envoy_listeners.TLSInspector())

	if !ingress.Spec.HasAvailableServices() {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder()))
	}

	sniUsed := map[string]bool{}

	for _, inbound := range ingress.Spec.GetNetworking().GetIngress().GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		destinations := destinationsPerService[service]
		destinations = append(destinations, destinationsPerService[mesh_proto.MatchAllTag]...)

		for _, destination := range destinations {
			destination := destination.
				WithoutTag(mesh_proto.ServiceTag).
				WithTags("mesh", inbound.GetMesh())
			sni := i.sni(service, destination)
			if sniUsed[sni] {
				continue
			}
			sniUsed[sni] = true
			inboundListenerBuilder = inboundListenerBuilder.
				Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder().
					Configure(envoy_listeners.FilterChainMatch(i.sni(service, destination))).
					Configure(envoy_listeners.TcpProxy(service, envoy_common.ClusterSubset{
						ClusterName: service,
						Tags:        destination,
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

func (_ IngressGenerator) sni(service string, tags envoy_common.Tags) string {
	pairs := []string{}
	for k, v := range tags {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return fmt.Sprintf("%s{%s}", service, strings.Join(pairs, ","))
}

func (_ IngressGenerator) services(proxy *model.Proxy) []string {
	var services []string
	for service := range proxy.OutboundTargets {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func (i IngressGenerator) generateCDS(services []string, destinationsPerService map[string][]envoy_common.Tags) (resources []*model.Resource, _ error) {
	for _, service := range services {
		edsCluster, err := envoy_clusters.NewClusterBuilder().
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

func (_ IngressGenerator) generateEDS(proxy *model.Proxy, services []string) (resources []*model.Resource) {
	for _, service := range services {
		endpoints := proxy.OutboundTargets[service]
		resources = append(resources, &model.Resource{
			Name:     service,
			Origin:   OriginIngress,
			Resource: envoy_endpoints.CreateClusterLoadAssignment(service, endpoints),
		})
	}
	return
}
