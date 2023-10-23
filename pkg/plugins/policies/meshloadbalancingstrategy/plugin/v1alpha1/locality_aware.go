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
			if policies_xds.IsSplitClusterName(key) && strings.HasPrefix(key, serviceName) {
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
			Resource: cla,
		})

		for key := range splitEndpoints {
			if policies_xds.IsSplitClusterName(key) && strings.HasPrefix(key, serviceName) {
				cla, err := ConfigureEnpointLocalityAwareLb(proxy, &conf, splitEndpoints, key, localZone)
				if err != nil {
					return err
				}
				rs.Add(&core_xds.Resource{
					Name:     key,
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
				ed := core_xds.Endpoint{}
				ed.Weight = lbEndpoint.LoadBalancingWeight.GetValue()
				tags := envoy_metadata.ExtractLbTags(lbEndpoint.Metadata)
				zoneName := tags[mesh_proto.ZoneTag]
				ed.Locality = &core_xds.Locality{
					Zone:     localZone,
					Weight:   1,
					Priority: 0,
				}
				if zoneName == localZone {
					for _, localRule := range localPriorityGroups {
						val, ok := tags[localRule.Key]
						if ok && val == localRule.Value {
							ed.Locality = &core_xds.Locality{
								Zone:     localZone,
								SubZone:  fmt.Sprintf("%s=%s", localRule.Key, val),
								Weight:   localRule.Weight,
								Priority: 0,
							}
							break
						}
					}
				} else {
					if len(crossZonePriorityGroups) == 0 {
						break endpointLoop
					}
				ruleLoop:
					for _, zoneRule := range crossZonePriorityGroups {
						switch zoneRule.Type {
						case api.Any:
							ed.Locality = &core_xds.Locality{
								Zone:     zoneName,
								Priority: zoneRule.Priority,
							}
							break ruleLoop
						case api.AnyExcept:
							if _, ok := zoneRule.Zones[zoneName]; ok {
								continue
							} else {
								ed.Locality = &core_xds.Locality{
									Zone:     zoneName,
									Priority: zoneRule.Priority,
								}
								break ruleLoop
							}
						case api.Only:
							if _, ok := zoneRule.Zones[zoneName]; ok {
								ed.Locality = &core_xds.Locality{
									Zone:     zoneName,
									Priority: zoneRule.Priority,
								}
								break ruleLoop
							}
						default:
							break endpointLoop
						}
					}
					if ed.Locality.Zone == localZone {
						break endpointLoop
					}
				}
				ed.Tags = tags
				address := lbEndpoint.GetEndpoint().GetAddress()
				if address.GetSocketAddress() != nil {
					ed.Target = address.GetSocketAddress().GetAddress()
					ed.Port = address.GetSocketAddress().GetPortValue()
				}
				if address.GetPipe() != nil {
					ed.UnixDomainPath = address.GetPipe().GetPath()
				}
				endpointsList = append(endpointsList, ed)
			}
		}
	}
	// TODO(lukidzi): use CLA Cache
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
