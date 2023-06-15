package generator

import (
	"sort"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	plugins_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

const (
	IngressProxy = "ingress-proxy"

	// OriginIngress is a marker to indicate by which ProxyGenerator resources were generated.
	OriginIngress = "ingress"
)

type IngressGenerator struct{}

func (i IngressGenerator) Generate(_ xds_context.Context, proxy *core_xds.Proxy) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	destinationsPerService := buildDestinations(proxy.ZoneIngressProxy)

	listener, err := i.generateLDS(proxy.ZoneIngressProxy.ZoneIngressResource, destinationsPerService, proxy.APIVersion)
	if err != nil {
		return nil, err
	}
	resources.Add(&core_xds.Resource{
		Name:     listener.GetName(),
		Origin:   OriginIngress,
		Resource: listener,
	})

	for _, mr := range proxy.ZoneIngressProxy.MeshResourceList {
		services := i.services(mr)

		cdsResources, err := i.generateCDS(services, destinationsPerService, proxy.APIVersion, mr)
		if err != nil {
			return nil, err
		}
		resources.Add(cdsResources...)

		edsResources, err := i.generateEDS(services, proxy.APIVersion, mr)
		if err != nil {
			return nil, err
		}
		resources.Add(edsResources...)
	}

	return resources, nil
}

// generateLDS generates one Ingress Listener. Generated listener assumes that
// mTLS is on. Using TLSInspector we sniff SNI value. SNI value has service name
// and tag values specified with the following format:
// "backend{cluster=2,version=1}". We take all possible destinations from
// TrafficRoutes + MeshHTTPRoutes + GatewayRoutes and generate
// FilterChainsMatcher for each unique destination. This approach has
// a limitation: additional tags on outbound in Universal mode won't work across
// different zones. Traffic is NOT decrypted here, therefore we don't need
// certificates and mTLS settings.
func (i IngressGenerator) generateLDS(
	ingress *core_mesh.ZoneIngressResource,
	destinations map[string][]envoy_tags.Tags,
	apiVersion core_xds.APIVersion,
) (envoy_common.NamedResource, error) {
	networking := ingress.Spec.GetNetworking()
	address, port := networking.GetAddress(), networking.GetPort()
	inboundListenerName := envoy_names.GetInboundListenerName(address, port)
	inboundListenerBuilder := envoy_listeners.NewListenerBuilder(apiVersion).
		Configure(envoy_listeners.InboundListener(inboundListenerName, address, port, core_xds.SocketAddressProtocolTCP)).
		Configure(envoy_listeners.TLSInspector())

	if len(ingress.Spec.AvailableServices) == 0 {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(envoy_listeners.NewFilterChainBuilder(apiVersion)))
	}

	sniUsed := map[string]bool{}

	for _, inbound := range ingress.Spec.GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		serviceDestinations := destinations[service]
		serviceDestinations = append(serviceDestinations, destinations[mesh_proto.MatchAllTag]...)
		clusterName := envoy_names.GetMeshClusterName(inbound.Mesh, service)

		for _, destination := range serviceDestinations {
			meshDestination := destination.
				WithTags(mesh_proto.ServiceTag, service).
				WithTags("mesh", inbound.GetMesh())
			sni := tls.SNIFromTags(meshDestination)
			if sniUsed[sni] {
				continue
			}
			sniUsed[sni] = true

			split := plugins_xds.NewSplitBuilder().
				WithClusterName(clusterName).
				WithLBMetadata(meshDestination.WithoutTags(mesh_proto.ServiceTag)).
				Build()

			inboundListenerBuilder = inboundListenerBuilder.Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(apiVersion).Configure(
					envoy_listeners.MatchTransportProtocol("tls"),
					envoy_listeners.MatchServerNames(sni),
					envoy_listeners.TCPProxy(clusterName, split),
				),
			))
		}
	}

	return inboundListenerBuilder.Build()
}

func tagsFromTargetRef(targetRef common_api.TargetRef) (envoy_tags.Tags, bool) {
	var service string
	var tags envoy_tags.Tags

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

	return mesh_proto.Merge(tags).WithTags(mesh_proto.ServiceTag, service), true
}

func (_ IngressGenerator) services(mr *core_xds.MeshIngressResources) []string {
	var services []string
	for service := range mr.EndpointMap {
		services = append(services, service)
	}
	sort.Strings(services)
	return services
}

func (i IngressGenerator) generateCDS(
	services []string,
	destinations map[string][]envoy_tags.Tags,
	apiVersion core_xds.APIVersion,
	mr *core_xds.MeshIngressResources,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		clusterName := envoy_names.GetMeshClusterName(mr.Mesh.GetMeta().GetName(), service)

		tagSlice := envoy_tags.TagsSlice(append(destinations[service], destinations[mesh_proto.MatchAllTag]...))
		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(envoy_tags.Without(mesh_proto.ServiceTag), envoy_tags.With("mesh"))

		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion).
			Configure(envoy_clusters.EdsCluster(clusterName)).
			Configure(envoy_clusters.LbSubset(tagKeySlice)).
			Configure(envoy_clusters.DefaultTimeout()).
			Build()
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   OriginIngress,
			Resource: edsCluster,
		})
	}
	return resources, nil
}

func (_ IngressGenerator) generateEDS(
	services []string,
	apiVersion core_xds.APIVersion,
	mr *core_xds.MeshIngressResources,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		endpoints := mr.EndpointMap[service]

		clusterName := envoy_names.GetMeshClusterName(mr.Mesh.GetMeta().GetName(), service)
		cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   OriginIngress,
			Resource: cla,
		})
	}
	return resources, nil
}
