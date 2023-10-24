package v1alpha1

import (
	"fmt"
	"strings"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	policies_xds "github.com/kumahq/kuma/pkg/plugins/policies/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/envoy/tags"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

func ConfigureStaticEndpointsLocalityAware(
	proxy *core_xds.Proxy,
	endpoints policies_xds.EndpointMap,
	splitEndpoints policies_xds.EndpointMap,
	cluster *envoy_cluster.Cluster,
	conf api.Conf,
	serviceName string,
) error {
	var localZone string
	if inbounds := proxy.Dataplane.Spec.GetNetworking().GetInbound(); len(inbounds) != 0 {
		localZone = inbounds[0].GetTags()[mesh_proto.ZoneTag]
	}

	if conf.LocalityAwareness != nil {
		cla, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, endpoints, serviceName, localZone)
		if err != nil {
			return err
		}
		cluster.LoadAssignment = cla.(*envoy_endpoint.ClusterLoadAssignment)

		for key := range splitEndpoints {
			if tags.IsSplitCluster(key) && strings.HasPrefix(key, serviceName) {
				cla, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, splitEndpoints, key, localZone)
				if err != nil {
					return err
				}
				cluster.LoadAssignment = cla.(*envoy_endpoint.ClusterLoadAssignment)
			}
		}
	}
	return nil
}

func ConfigureEndpointsLocalityAware(
	proxy *core_xds.Proxy,
	endpoints policies_xds.EndpointMap,
	splitEndpoints policies_xds.EndpointMap,
	conf api.Conf,
	rs *core_xds.ResourceSet,
	serviceName string,
	localZone string,
) error {
	if conf.LocalityAwareness != nil {
		cla, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, endpoints, serviceName, localZone)
		if err != nil {
			return err
		}
		rs.Add(&core_xds.Resource{
			Name:     serviceName,
			Origin:   generator.OriginOutbound, // needs to be set so GatherEndpoints can get them
			Resource: cla,
		})

		for splitName := range splitEndpoints {
			if tags.IsSplitCluster(splitName) && strings.HasPrefix(splitName, serviceName) {
				cla, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, splitEndpoints, splitName, localZone)
				if err != nil {
					return err
				}
				rs.Add(&core_xds.Resource{
					Name:     splitName,
					Origin:   generator.OriginOutbound, // needs to be set so GatherEndpoints can get them
					Resource: cla,
				})
			}
		}
	}
	return nil
}

func ConfigureEnpointLocalityAwareLb(
	proxy *core_xds.Proxy,
	conf *api.Conf,
	endpoints policies_xds.EndpointMap,
	serviceName string,
	localZone string,
) (proto.Message, error) {
	if len(endpoints) == 0 {
		return nil, nil
	}
	localPriorityGroups, crossZonePriorityGroups := GetPriority(conf, proxy.Dataplane.Spec.TagSet(), localZone)
	endpointsList := []core_xds.Endpoint{}
	for _, endpoint := range endpoints[serviceName] {
		for _, localityLbEndpoint := range endpoint.Endpoints {
		endpointLoop:
			for _, lbEndpoint := range localityLbEndpoint.LbEndpoints {
				ed := createEndpoint(lbEndpoint, localZone)
				zoneName := ed.Tags[mesh_proto.ZoneTag]

				// nolint: gocritic
				if zoneName == localZone {
					configureLocalZoneEndpointLocality(localPriorityGroups, &ed, localZone)
				} else if len(crossZonePriorityGroups) > 0 {
					configureCrossZoneEndpointLocality(crossZonePriorityGroups, &ed, zoneName)
					// when endpoint wasn't matched with any rule
					if ed.Locality.Zone == localZone {
						break endpointLoop
					}
				} else {
					break endpointLoop
				}
				endpointsList = append(endpointsList, ed)
			}
		}
	}
	// TODO(lukidzi): use CLA Cache https://github.com/kumahq/kuma/issues/8121
	cla, err := envoy_endpoints.CreateClusterLoadAssignment(serviceName, endpointsList, proxy.APIVersion)
	if err != nil {
		return nil, err
	}

	clusterLb := cla.(*envoy_endpoint.ClusterLoadAssignment)
	overprovisingFactor := defaultOverprovisingFactor
	if conf.LocalityAwareness.FailoverThreshold != nil {
		overprovisingFactor = uint32(100/conf.LocalityAwareness.FailoverThreshold.Percentage.IntVal) * 100
	}
	if clusterLb.Policy == nil {
		clusterLb.Policy = &envoy_endpoint.ClusterLoadAssignment_Policy{
			OverprovisioningFactor: wrapperspb.UInt32(overprovisingFactor),
		}
	} else {
		clusterLb.Policy.OverprovisioningFactor = wrapperspb.UInt32(overprovisingFactor)
	}
	return cla, nil
}

func createEndpoint(lbEndpoint *envoy_endpoint.LbEndpoint, localZone string) core_xds.Endpoint {
	endpoint := core_xds.Endpoint{}
	endpoint.Weight = lbEndpoint.LoadBalancingWeight.GetValue()
	endpoint.Tags = envoy_metadata.ExtractLbTags(lbEndpoint.Metadata)
	address := lbEndpoint.GetEndpoint().GetAddress()
	if address.GetSocketAddress() != nil {
		endpoint.Target = address.GetSocketAddress().GetAddress()
		endpoint.Port = address.GetSocketAddress().GetPortValue()
	}
	if address.GetPipe() != nil {
		endpoint.UnixDomainPath = address.GetPipe().GetPath()
	}
	endpoint.Locality = &core_xds.Locality{
		Zone:     localZone,
		Weight:   1,
		Priority: 0,
	}
	return endpoint
}

func configureCrossZoneEndpointLocality(crossZonePriorityGroups []CrossZoneLbGroup, endpoint *core_xds.Endpoint, zoneName string) {
	for _, zoneRule := range crossZonePriorityGroups {
		switch zoneRule.Type {
		case api.Any:
			endpoint.Locality = &core_xds.Locality{
				Zone:     zoneName,
				Priority: zoneRule.Priority,
			}
			return
		case api.AnyExcept:
			if _, ok := zoneRule.Zones[zoneName]; ok {
				continue
			} else {
				endpoint.Locality = &core_xds.Locality{
					Zone:     zoneName,
					Priority: zoneRule.Priority,
				}
				return
			}
		case api.Only:
			if _, ok := zoneRule.Zones[zoneName]; ok {
				endpoint.Locality = &core_xds.Locality{
					Zone:     zoneName,
					Priority: zoneRule.Priority,
				}
				return
			}
		default:
		}
	}
}

func configureLocalZoneEndpointLocality(localPriorityGroups []LocalLbGroup, endpoint *core_xds.Endpoint, localZone string) {
	for _, localRule := range localPriorityGroups {
		val, ok := endpoint.Tags[localRule.Key]
		if ok && val == localRule.Value {
			endpoint.Locality = &core_xds.Locality{
				Zone:     localZone,
				SubZone:  fmt.Sprintf("%s=%s", localRule.Key, val),
				Weight:   localRule.Weight,
				Priority: 0,
			}
			break
		}
	}
}
