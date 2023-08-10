package generator

import (
	"context"
	"sort"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
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

	// OriginIngress is a marker to indicate by which ProxyGenerator resources
	// were generated.
	OriginIngress = "ingress"
)

type IngressGenerator struct{}

func (i IngressGenerator) Generate(
	_ context.Context,
	_ xds_context.Context,
	proxy *core_xds.Proxy,
) (*core_xds.ResourceSet, error) {
	resources := core_xds.NewResourceSet()

	dest := buildDestinations(proxy.ZoneIngressProxy)

	listener, err := i.generateLDS(
		proxy.ZoneIngressProxy.ZoneIngressResource,
		dest,
		proxy.APIVersion,
	)
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

		cdsResources, err := i.generateCDS(services, dest, proxy.APIVersion, mr)
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
	dest destinations,
	apiVersion core_xds.APIVersion,
) (envoy_common.NamedResource, error) {
	networking := ingress.Spec.GetNetworking()
	address, port := networking.GetAddress(), networking.GetPort()
	inboundListenerBuilder := envoy_listeners.NewInboundListenerBuilder(
		apiVersion,
		address,
		port,
		core_xds.SocketAddressProtocolTCP,
	).Configure(envoy_listeners.TLSInspector())

	if len(ingress.Spec.AvailableServices) == 0 {
		inboundListenerBuilder = inboundListenerBuilder.
			Configure(envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource),
			))
	}

	sniUsed := map[string]bool{}

	for _, inbound := range ingress.Spec.GetAvailableServices() {
		service := inbound.Tags[mesh_proto.ServiceTag]
		mesh := inbound.GetMesh()
		serviceDestinations := dest.get(mesh, service)
		clusterName := envoy_names.GetMeshClusterName(inbound.Mesh, service)

		for _, destination := range serviceDestinations {
			sni := tls.SNIFromTags(destination.
				WithTags(mesh_proto.ServiceTag, service).
				WithTags("mesh", mesh),
			)
			if sniUsed[sni] {
				continue
			}
			sniUsed[sni] = true

			cluster := envoy_common.NewCluster(
				envoy_common.WithName(clusterName),
				envoy_common.WithTags(destination),
			)

			filterChain := envoy_listeners.FilterChain(
				envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource).Configure(
					envoy_listeners.MatchTransportProtocol("tls"),
					envoy_listeners.MatchServerNames(sni),
					envoy_listeners.TcpProxyDeprecatedWithMetadata(
						clusterName,
						cluster,
					),
				),
			)

			inboundListenerBuilder = inboundListenerBuilder.Configure(filterChain)
		}
	}

	return inboundListenerBuilder.Build()
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
	dest destinations,
	apiVersion core_xds.APIVersion,
	mr *core_xds.MeshIngressResources,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource
	for _, service := range services {
		meshName := mr.Mesh.GetMeta().GetName()
		clusterName := envoy_names.GetMeshClusterName(meshName, service)

		tagSlice := envoy_tags.TagsSlice(dest.get(meshName, service))
		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(
			envoy_tags.Without(mesh_proto.ServiceTag),
		)

		edsCluster, err := envoy_clusters.NewClusterBuilder(apiVersion, clusterName).
			Configure(envoy_clusters.EdsCluster()).
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
		meshName := mr.Mesh.GetMeta().GetName()
		clusterName := envoy_names.GetMeshClusterName(meshName, service)

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
