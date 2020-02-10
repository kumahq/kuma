package endpoints

import (
	pstruct "github.com/golang/protobuf/ptypes/struct"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"

	core_xds "github.com/Kong/kuma/pkg/core/xds"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *v2.ClusterLoadAssignment {
	return &v2.ClusterLoadAssignment{
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

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint) *v2.ClusterLoadAssignment {
	lbEndpoints := make([]*envoy_endpoint.LbEndpoint, 0, len(endpoints))
	for _, ep := range endpoints {
		lbEndpoints = append(lbEndpoints, &envoy_endpoint.LbEndpoint{
			Metadata: CreateLbMetadata(ep.Tags),
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
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*envoy_endpoint.LocalityLbEndpoints{{
			LbEndpoints: lbEndpoints,
		}},
	}
}

func CreateLbMetadata(tags map[string]string) *envoy_core.Metadata {
	if len(tags) == 0 {
		return nil
	}
	fields := map[string]*pstruct.Value{}
	for key, value := range tags {
		fields[key] = &pstruct.Value{
			Kind: &pstruct.Value_StringValue{
				StringValue: value,
			},
		}
	}
	return &envoy_core.Metadata{
		FilterMetadata: map[string]*pstruct.Struct{
			"envoy.lb": &pstruct.Struct{
				Fields: fields,
			},
		},
	}
}
