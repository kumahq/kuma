package zoneproxy

import (
	"fmt"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/util/pointer"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	envoy_clusters "github.com/kumahq/kuma/pkg/xds/envoy/clusters"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_listeners "github.com/kumahq/kuma/pkg/xds/envoy/listeners"
	envoy_names "github.com/kumahq/kuma/pkg/xds/envoy/names"
	envoy_tags "github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

func GenerateCDS(
	meshDestinations MeshDestinations,
	services envoy_common.Services,
	apiVersion core_xds.APIVersion,
	meshName string,
	origin string,
) ([]*core_xds.Resource, error) {
	matchAllDestinations := meshDestinations.KumaIoServices[mesh_proto.MatchAllTag]

	var resources []*core_xds.Resource
	for _, service := range services.Sorted() {
		clusters := services[service]

		var tagsSlice envoy_tags.TagsSlice
		for _, cluster := range clusters.Clusters() {
			tagsSlice = append(tagsSlice, cluster.Tags())
		}
		tagSlice := append(tagsSlice, matchAllDestinations...)

		tagKeySlice := tagSlice.ToTagKeysSlice().Transform(
			envoy_tags.Without(mesh_proto.ServiceTag),
		)
		clusterName := envoy_names.GetMeshClusterName(meshName, service)
		if clusters.BackendRef().LegacyBackendRef != nil && clusters.BackendRef().LegacyBackendRef.ReferencesRealObject() {
			clusterName = service
		}
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
			Origin:   origin,
			Resource: edsCluster,
		})
	}

	return resources, nil
}

func GenerateEDS(
	services envoy_common.Services,
	endpointMap core_xds.EndpointMap,
	apiVersion core_xds.APIVersion,
	meshName string,
	origin string,
) ([]*core_xds.Resource, error) {
	var resources []*core_xds.Resource

	for _, service := range services.Sorted() {
		clusters := services[service]
		endpoints := endpointMap[service]
		clusterName := envoy_names.GetMeshClusterName(meshName, service)
		if clusters.BackendRef().LegacyBackendRef != nil && clusters.BackendRef().LegacyBackendRef.ReferencesRealObject() {
			clusterName = service
		}
		cla, err := envoy_endpoints.CreateClusterLoadAssignment(clusterName, endpoints, apiVersion)
		if err != nil {
			return nil, err
		}
		resources = append(resources, &core_xds.Resource{
			Name:     clusterName,
			Origin:   origin,
			Resource: cla,
		})
	}

	return resources, nil
}

// AddFilterChains adds filter chains to a listener. Generated listener assumes that
// mTLS is on. Using TLSInspector we sniff SNI value. SNI value has service name
// and tag values specified with the following format:
// "backend{cluster=2,version=1}". We take all possible destinations from
// TrafficRoutes + MeshHTTPRoutes + GatewayRoutes and generate
// FilterChainsMatcher for each unique destination. This approach has
// a limitation: additional tags on outbound in Universal mode won't work across
// different zones. Traffic is NOT decrypted here, therefore we don't need
// certificates and mTLS settings.
func AddFilterChains(
	availableServices []*mesh_proto.ZoneIngress_AvailableService,
	apiVersion core_xds.APIVersion,
	listenerBuilder *envoy_listeners.ListenerBuilder,
	meshDestinations MeshDestinations,
	endpointMap core_xds.EndpointMap,
) envoy_common.Services {
	servicesAcc := envoy_common.NewServicesAccumulator(nil)

	sniUsed := map[string]struct{}{}
	for _, service := range availableServices {
		serviceName := service.Tags[mesh_proto.ServiceTag]
		destinations := meshDestinations.KumaIoServices[serviceName]
		destinations = append(destinations, meshDestinations.KumaIoServices[mesh_proto.MatchAllTag]...)
		clusterName := envoy_names.GetMeshClusterName(service.Mesh, serviceName)
		serviceEndpoints := endpointMap[serviceName]

		for _, destination := range destinations {
			sni := tls.SNIFromTags(destination.
				WithTags(mesh_proto.ServiceTag, serviceName).
				WithTags("mesh", service.Mesh),
			)
			if _, ok := sniUsed[sni]; ok {
				continue
			}
			sniUsed[sni] = struct{}{}

			// relevantTags is a set of tags for which it actually makes sense to do LB split on.
			// If the endpoint list is the same with or without the tag, we should just not do the split.
			// However, we should preserve full SNI, because the client expects Zone Proxy to support it.
			// This solves the problem that Envoy deduplicate endpoints of the same address and different metadata.
			// example 1:
			// Ingress1 (10.0.0.1) supports service:a,version:1 and service:a,version:2
			// Ingress2 (10.0.0.2) supports service:a,version:1 and service:a,version:2
			// If we want to split by version, we don't need to do LB subset on version.
			//
			// example 2:
			// Ingress1 (10.0.0.1) supports service:a,version:1
			// Ingress2 (10.0.0.2) supports service:a,version:2
			// If we want to split by version, we need LB subset.
			relevantTags := envoy_tags.Tags{}
			for key, value := range destination {
				matchedTargets := map[string]struct{}{}
				allTargets := map[string]struct{}{}
				for _, endpoint := range serviceEndpoints {
					address := endpoint.Address()
					if endpoint.Tags[key] == value || value == mesh_proto.MatchAllTag {
						matchedTargets[address] = struct{}{}
					}
					allTargets[address] = struct{}{}
				}
				if len(matchedTargets) < len(allTargets) {
					relevantTags[key] = value
				}
			}

			cluster := envoy_common.NewCluster(
				envoy_common.WithName(clusterName),
				envoy_common.WithService(serviceName),
				envoy_common.WithTags(relevantTags),
			)
			cluster.SetMesh(service.Mesh)

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

			listenerBuilder.Configure(filterChain)

			servicesAcc.Add(cluster)
		}
	}

	for _, refDest := range meshDestinations.BackendRefs {
		if _, ok := sniUsed[refDest.SNI]; ok {
			continue
		}
		sniUsed[refDest.SNI] = struct{}{}

		// todo(jakubdyszkiewicz) support splits
		relevantTags := envoy_tags.Tags{}

		// Destination name usually equals to kuma.io/service so we will add already existing cluster which will be
		// then deduplicated in later steps
		cluster := envoy_common.NewCluster(
			envoy_common.WithName(refDest.DestinationName),
			envoy_common.WithService(refDest.DestinationName),
			envoy_common.WithTags(relevantTags),
		)
		cluster.SetMesh(refDest.Mesh)

		filterChain := envoy_listeners.FilterChain(
			envoy_listeners.NewFilterChainBuilder(apiVersion, envoy_common.AnonymousResource).Configure(
				envoy_listeners.MatchTransportProtocol("tls"),
				envoy_listeners.MatchServerNames(refDest.SNI),
				envoy_listeners.TcpProxyDeprecatedWithMetadata(
					refDest.DestinationName,
					cluster,
				),
			),
		)

		listenerBuilder.Configure(filterChain)
		servicesAcc.AddBackendRef(pointer.Deref(refDest.Resource), cluster)
	}

	return servicesAcc.Services()
}

func GenerateEmptyDirectResponseListener(proxy *core_xds.Proxy, address string, port uint32) (envoy_common.NamedResource, error) {
	response := fmt.Sprintf(`{"proxy":"%s","zone":"%s"}`, proxy.Id.String(), proxy.Zone)
	filterChain := envoy_listeners.FilterChain(
		envoy_listeners.NewFilterChainBuilder(proxy.APIVersion, envoy_common.AnonymousResource).
			Configure(envoy_listeners.NetworkDirectResponse(response)))

	listenerBuilder := envoy_listeners.NewInboundListenerBuilder(proxy.APIVersion, address, port, core_xds.SocketAddressProtocolTCP).
		Configure(envoy_listeners.TLSInspector()).
		Configure(filterChain)

	listener, err := listenerBuilder.Build()
	if err != nil {
		return nil, err
	}

	return listener, nil
}
