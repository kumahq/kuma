package v1alpha1

import (
	"fmt"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/generator"
)

const defaultOverprovisingFactor uint32 = 200

func ConfigureStaticEndpointsLocalityAware(
	proxy *core_xds.Proxy,
	endpoints []*envoy_endpoint.ClusterLoadAssignment,
	cluster *envoy_cluster.Cluster,
	conf api.Conf,
	serviceName string,
) error {
	var localZone string
	if tags := proxy.Dataplane.Spec.TagSet().Values(mesh_proto.ZoneTag); len(tags) > 0 {
		localZone = tags[0]
	}

	if conf.LocalityAwareness != nil {
		for _, cla := range endpoints {
			if cla.ClusterName == serviceName {
				loadAssignment, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, cla, cla.ClusterName, localZone)
				if err != nil {
					return err
				}
				cluster.LoadAssignment = loadAssignment.(*envoy_endpoint.ClusterLoadAssignment)
			}
		}
	}
	return nil
}

func ConfigureEndpointsLocalityAware(
	proxy *core_xds.Proxy,
	endpoints []*envoy_endpoint.ClusterLoadAssignment,
	conf api.Conf,
	rs *core_xds.ResourceSet,
	serviceName string,
	localZone string,
) error {
	if conf.LocalityAwareness != nil {
		for _, cla := range endpoints {
			if cla.ClusterName == serviceName {
				loadAssignment, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, cla, cla.ClusterName, localZone)
				if err != nil {
					return err
				}
				rs.Add(&core_xds.Resource{
					Name:     cla.ClusterName,
					Origin:   generator.OriginOutbound, // needs to be set so GatherEndpoints can get them
					Resource: loadAssignment,
				})
			}
		}
	}
	return nil
}

func ConfigureEnpointLocalityAwareLb(
	proxy *core_xds.Proxy,
	conf *api.Conf,
	cla *envoy_endpoint.ClusterLoadAssignment,
	serviceName string,
	localZone string,
) (proto.Message, error) {
	localPriorityGroups, crossZonePriorityGroups := GetLocalityGroups(conf, proxy.Dataplane.Spec.TagSet(), localZone)
	endpointsList := []core_xds.Endpoint{}
	for _, localityLbEndpoint := range cla.Endpoints {
		for _, lbEndpoint := range localityLbEndpoint.LbEndpoints {
			ed := createEndpoint(lbEndpoint, localZone)
			zoneName := ed.Tags[mesh_proto.ZoneTag]

			if zoneName == localZone {
				configureLocalZoneEndpointLocality(localPriorityGroups, &ed, localZone)
				endpointsList = append(endpointsList, ed)
			} else if configureCrossZoneEndpointLocality(crossZonePriorityGroups, &ed, zoneName) {
				endpointsList = append(endpointsList, ed)
			}
		}
	}
	// TODO(lukidzi): use CLA Cache https://github.com/kumahq/kuma/issues/8121
	loadAssignment, err := envoy_endpoints.CreateClusterLoadAssignment(serviceName, endpointsList, proxy.APIVersion)
	if err != nil {
		return nil, err
	}

	clusterLb := loadAssignment.(*envoy_endpoint.ClusterLoadAssignment)
	overprovisingFactor := defaultOverprovisingFactor
	if conf.LocalityAwareness.CrossZone != nil && conf.LocalityAwareness.CrossZone.FailoverThreshold != nil {
		overprovisingFactor = uint32(100/conf.LocalityAwareness.CrossZone.FailoverThreshold.Percentage.IntVal) * 100
	}
	if clusterLb.Policy == nil {
		clusterLb.Policy = &envoy_endpoint.ClusterLoadAssignment_Policy{
			OverprovisioningFactor: wrapperspb.UInt32(overprovisingFactor),
		}
	} else {
		clusterLb.Policy.OverprovisioningFactor = wrapperspb.UInt32(overprovisingFactor)
	}
	return loadAssignment, nil
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

func configureCrossZoneEndpointLocality(crossZonePriorityGroups []CrossZoneLbGroup, endpoint *core_xds.Endpoint, zoneName string) bool {
	for _, zoneRule := range crossZonePriorityGroups {
		switch zoneRule.Type {
		case api.Any:
			endpoint.Locality = &core_xds.Locality{
				Zone:     zoneName,
				Priority: zoneRule.Priority,
			}
			return true
		case api.AnyExcept:
			if _, ok := zoneRule.Zones[zoneName]; ok {
				continue
			} else {
				endpoint.Locality = &core_xds.Locality{
					Zone:     zoneName,
					Priority: zoneRule.Priority,
				}
				return true
			}
		case api.Only:
			if _, ok := zoneRule.Zones[zoneName]; ok {
				endpoint.Locality = &core_xds.Locality{
					Zone:     zoneName,
					Priority: zoneRule.Priority,
				}
				return true
			}
		default:
		}
	}
	return false
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
