package endpoints

import (
	"sort"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	proto_wrappers "google.golang.org/protobuf/types/known/wrapperspb"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
)

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint) *envoy_endpoint.ClusterLoadAssignment {
	localityLbEndpoints := LocalityLbEndpointsMap{}

	for _, ep := range endpoints {
		var address *envoy_core.Address
		if ep.UnixDomainPath != "" {
			address = &envoy_core.Address{
				Address: &envoy_core.Address_Pipe{
					Pipe: &envoy_core.Pipe{
						Path: ep.UnixDomainPath,
					},
				},
			}
		} else {
			address = &envoy_core.Address{
				Address: &envoy_core.Address_SocketAddress{
					SocketAddress: &envoy_core.SocketAddress{
						Protocol: envoy_core.SocketAddress_TCP,
						Address:  ep.Target,
						PortSpecifier: &envoy_core.SocketAddress_PortValue{
							PortValue: ep.Port,
						},
					},
				},
			}
		}
		lbEndpoint := &envoy_endpoint.LbEndpoint{
			Metadata: envoy.EndpointMetadata(ep.Tags),
			HostIdentifier: &envoy_endpoint.LbEndpoint_Endpoint{
				Endpoint: &envoy_endpoint.Endpoint{
					Address: address,
				}},
		}
		if ep.Weight > 0 {
			lbEndpoint.LoadBalancingWeight = &proto_wrappers.UInt32Value{
				Value: ep.Weight,
			}
		}
		localityLbEndpoints.append(ep, lbEndpoint)
	}

	for _, lbEndpoints := range localityLbEndpoints {
		// sort the slice to ensure stable Envoy configuration
		sortLbEndpoints(lbEndpoints.LbEndpoints)
	}

	return &envoy_endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints:   localityLbEndpoints.asSlice(),
	}
}

type LocalityLbEndpointsMap map[string]*envoy_endpoint.LocalityLbEndpoints

func (l LocalityLbEndpointsMap) append(ep core_xds.Endpoint, endpoint *envoy_endpoint.LbEndpoint) {
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
	l[key].LbEndpoints = append(l[key].LbEndpoints, endpoint)
}

func (l LocalityLbEndpointsMap) asSlice() []*envoy_endpoint.LocalityLbEndpoints {
	slice := make([]*envoy_endpoint.LocalityLbEndpoints, 0, len(l))

	for _, lle := range l {
		sortLbEndpoints(lle.LbEndpoints)
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
