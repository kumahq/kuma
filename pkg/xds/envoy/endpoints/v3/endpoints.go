package endpoints

import (
	"sort"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	proto_wrappers "google.golang.org/protobuf/types/known/wrapperspb"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *envoy_endpoint.ClusterLoadAssignment {
	return &envoy_endpoint.ClusterLoadAssignment{
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

func CreateStaticEndpointUnixSocket(clusterName string, path string) *envoy_endpoint.ClusterLoadAssignment {
	return &envoy_endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*envoy_endpoint.LbEndpoint{{
				HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
					Endpoint: &envoy_endpoint.Endpoint{
						Address: &envoy_core.Address{
							Address: &envoy_core.Address_Pipe{
								Pipe: &envoy_core.Pipe{
									Path: path,
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint) *envoy_endpoint.ClusterLoadAssignment {
	localityLbEndpoints := LocalityLbEndpointsMap{}

	for _, ep := range endpoints {
		lbEndpoints := localityLbEndpoints.Get(ep)
		lbEndpoints.LbEndpoints = append(lbEndpoints.LbEndpoints, &envoy_endpoint.LbEndpoint{
			LoadBalancingWeight: &proto_wrappers.UInt32Value{
				Value: ep.Weight,
			},
			Metadata: envoy.EndpointMetadata(ep.Tags),
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

	for _, lbEndpoints := range localityLbEndpoints {
		// sort the slice to ensure stable Envoy configuration
		sortLbEndpoints(lbEndpoints.LbEndpoints)
	}

	return &envoy_endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   localityLbEndpoints.AsSlice(),
	}
}

func sortLbEndpoints(lbEndpoints []*envoy_endpoint.LbEndpoint) {
	sort.Slice(lbEndpoints, func(i, j int) bool {
		left, right := lbEndpoints[i], lbEndpoints[j]
		leftAddr := left.GetEndpoint().GetAddress().GetSocketAddress().GetAddress()
		rightAddr := right.GetEndpoint().GetAddress().GetSocketAddress().GetAddress()
		if leftAddr == rightAddr {
			return left.GetEndpoint().GetAddress().GetSocketAddress().GetPortValue() < right.GetEndpoint().GetAddress().GetSocketAddress().GetPortValue()
		}
		return leftAddr < rightAddr
	})
}

type LocalityLbEndpointsMap map[string]*envoy_endpoint.LocalityLbEndpoints

func (l LocalityLbEndpointsMap) Get(ep core_xds.Endpoint) *envoy_endpoint.LocalityLbEndpoints {
	key := ep.LocalityString()
	if _, ok := l[key]; !ok {
		var locality *envoy_core.Locality
		priority := uint32(0)
		if ep.HasLocality() {
			locality = &envoy_core.Locality{
				Zone: ep.Locality.Zone,
			}
			priority = ep.Locality.Priority
		}

		l[key] = &envoy_endpoint.LocalityLbEndpoints{
			LbEndpoints: make([]*envoy_endpoint.LbEndpoint, 0),
			Locality:    locality,
			Priority:    priority,
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
