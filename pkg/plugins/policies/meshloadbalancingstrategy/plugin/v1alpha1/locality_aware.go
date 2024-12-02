package v1alpha1

import (
	"fmt"
	"strings"

	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
	envoy_endpoints "github.com/kumahq/kuma/pkg/xds/envoy/endpoints"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

const defaultOverprovisingFactor uint32 = 200

func ConfigureStaticEndpointsLocalityAware(
	tags mesh_proto.MultiValueTagSet,
	endpoints []*envoy_endpoint.ClusterLoadAssignment,
	cluster *envoy_cluster.Cluster,
	conf api.Conf,
	serviceName string,
	localZone string,
	apiVersion core_xds.APIVersion,
	egressEnabled bool,
	origin string,
) error {
	if conf.LocalityAwareness != nil && (conf.LocalityAwareness.LocalZone != nil || conf.LocalityAwareness.CrossZone != nil) {
		for _, cla := range endpoints {
			if cla.ClusterName == serviceName {
				loadAssignment, err := ConfigureEndpointLocalityAwareLb(tags, &conf, cla, cla.ClusterName, localZone, apiVersion, egressEnabled, origin)
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
	tags mesh_proto.MultiValueTagSet,
	endpoints []*envoy_endpoint.ClusterLoadAssignment,
	conf api.Conf,
	rs *core_xds.ResourceSet,
	serviceName string,
	localZone string,
	apiVersion core_xds.APIVersion,
	egressEnabled bool,
	origin string,
) error {
	if conf.LocalityAwareness != nil && (conf.LocalityAwareness.LocalZone != nil || conf.LocalityAwareness.CrossZone != nil) {
		for _, cla := range endpoints {
			if cla.ClusterName == serviceName {
				loadAssignment, err := ConfigureEndpointLocalityAwareLb(tags, &conf, cla, cla.ClusterName, localZone, apiVersion, egressEnabled, origin)
				if err != nil {
					return err
				}
				rs.Add(&core_xds.Resource{
					Name:     cla.ClusterName,
					Origin:   origin, // needs to be set so GatherEndpoints can get them
					Resource: loadAssignment,
				})
			}
		}
	}
	return nil
}

func ConfigureEndpointLocalityAwareLb(
	tags mesh_proto.MultiValueTagSet,
	conf *api.Conf,
	cla *envoy_endpoint.ClusterLoadAssignment,
	serviceName string,
	localZone string,
	apiVersion core_xds.APIVersion,
	egressEnabled bool,
	origin string,
) (proto.Message, error) {
	localPriorityGroups, crossZonePriorityGroups := GetLocalityGroups(conf, tags, localZone)
	var endpointsList []core_xds.Endpoint
	for _, localityLbEndpoint := range cla.Endpoints {
		for _, lbEndpoint := range localityLbEndpoint.LbEndpoints {
			ed := createEndpoint(lbEndpoint, localZone)
			zoneName := ed.Tags[mesh_proto.ZoneTag]

			// nolint:gocritic
			if zoneName == localZone {
				configureLocalZoneEndpointLocality(localPriorityGroups, &ed, localZone)
				endpointsList = append(endpointsList, ed)
			} else if egressEnabled && origin != egress.OriginEgress {
				ed.Locality = egressLocality(crossZonePriorityGroups)
				endpointsList = append(endpointsList, ed)
			} else if configureCrossZoneEndpointLocality(crossZonePriorityGroups, &ed, zoneName) {
				endpointsList = append(endpointsList, ed)
			}
		}
	}
	// TODO(lukidzi): use CLA Cache https://github.com/kumahq/kuma/issues/8121
	loadAssignment, err := envoy_endpoints.CreateClusterLoadAssignment(serviceName, endpointsList, apiVersion)
	if err != nil {
		return nil, err
	}

	clusterLb := loadAssignment.(*envoy_endpoint.ClusterLoadAssignment)
	overprovisingFactor := defaultOverprovisingFactor
	if conf.LocalityAwareness.CrossZone != nil && conf.LocalityAwareness.CrossZone.FailoverThreshold != nil {
		val, err := common_api.NewDecimalFromIntOrString(conf.LocalityAwareness.CrossZone.FailoverThreshold.Percentage)
		if err == nil && !val.IsZero() {
			overprovisingFactor = uint32(100/val.InexactFloat64()) * 100
		}
	}
	if clusterLb.Policy == nil {
		clusterLb.Policy = &envoy_endpoint.ClusterLoadAssignment_Policy{}
	}
	clusterLb.Policy.OverprovisioningFactor = wrapperspb.UInt32(overprovisingFactor)
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

func egressLocality(crossZoneGroups []CrossZoneLbGroup) *core_xds.Locality {
	builder := strings.Builder{}
	for _, group := range crossZoneGroups {
		switch group.Type {
		case api.Only:
			builder.WriteString(fmt.Sprintf("%d:%s", group.Priority, strings.Join(util_maps.SortedKeys(group.Zones), ",")))
		case api.Any:
			builder.WriteString(fmt.Sprintf("%d:%s", group.Priority, group.Type))
		case api.AnyExcept:
			builder.WriteString(fmt.Sprintf("%d:%s:%s", group.Priority, group.Type, strings.Join(util_maps.SortedKeys(group.Zones), ",")))
		default:
			continue
		}
		builder.WriteString(";")
	}

	return &core_xds.Locality{
		Zone:     fmt.Sprintf("egress_%s", sha256.Hash(builder.String())[:8]),
		Priority: 1,
	}
}
