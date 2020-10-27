package endpoints

import (
	"sort"

	envoy_api "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	proto_wrappers "github.com/golang/protobuf/ptypes/wrappers"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *envoy_api.ClusterLoadAssignment {
	return &envoy_api.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*envoy_endpoint.LbEndpoint{{
				HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
					Endpoint: &envoy_endpoint.Endpoint{
						Address: &envoy_core.Address{
							Address: &envoy_core.Address_SocketAddress{
								SocketAddress: &envoy_core.SocketAddress{
									Protocol: envoy_core.SocketAddress_TCP,
									Address:  address,
									PortSpecifier: &envoy_core.SocketAddress_PortValue{
										PortValue: port,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint) *envoy_api.ClusterLoadAssignment {
	localityLbEndpoints := LocalityLbEndpointsMap{}

	for _, ep := range endpoints {
		lbEndpoints := localityLbEndpoints.Get(ep)
		lbEndpoints.LbEndpoints = append(lbEndpoints.LbEndpoints, &envoy_endpoint.LbEndpoint{
			LoadBalancingWeight: &proto_wrappers.UInt32Value{
				Value: ep.Weight,
			},
			Metadata: envoy_common.EndpointMetadata(ep.Tags),
			HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
				Endpoint: &envoy_endpoint.Endpoint{
					Address: &envoy_core.Address{
						Address: &envoy_core.Address_SocketAddress{
							SocketAddress: &envoy_core.SocketAddress{
								Protocol: envoy_core.SocketAddress_TCP,
								Address:  ep.Target,
								PortSpecifier: &envoy_core.SocketAddress_PortValue{
									PortValue: ep.Port,
								},
							},
						},
					},
				}},
		})
	}

	return &envoy_api.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   localityLbEndpoints.AsSlice(),
	}
}

type LocalityLbEndpointsMap map[string]*envoy_endpoint.LocalityLbEndpoints

func (l LocalityLbEndpointsMap) Get(ep core_xds.Endpoint) *envoy_endpoint.LocalityLbEndpoints {
	key := ep.LocalityString()
	if _, ok := l[key]; !ok {
		var locality *envoy_core.Locality
		priority := uint32(0)
		if ep.HasLocality() {
			locality = &envoy_core.Locality{
				Region:  ep.Locality.Region,
				Zone:    ep.Locality.Zone,
				SubZone: ep.Locality.SubZone,
			}
			priority = ep.Locality.Priority
		}

		l[key] = &envoy_endpoint.LocalityLbEndpoints{
			LbEndpoints: make([]*envoy_endpoint.LbEndpoint, 0),
			Locality:    locality,
			//LoadBalancingWeight:
			Priority: priority,
		}
	}

	return l[key]
}

func (l LocalityLbEndpointsMap) AsSlice() []*envoy_endpoint.LocalityLbEndpoints {
	slice := make([]*envoy_endpoint.LocalityLbEndpoints, 0)

	for _, lle := range l {
		slice = append(slice, lle)
	}

	// sort the slice to ensure stable Envoy configuration
	sort.Slice(slice, func(i, j int) bool {
		left, right := slice[i], slice[j]
		return (left.Locality.Region + left.Locality.Zone + left.Locality.SubZone) >
			(right.Locality.Region + right.Locality.Zone + right.Locality.SubZone)
	})

	return slice
}
