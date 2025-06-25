package v1alpha1

import (
	"fmt"
	"strings"

	envoy_endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	util_maps "github.com/kumahq/kuma/pkg/util/maps"
	"github.com/kumahq/kuma/pkg/xds/cache/sha256"
	"github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
	envoy_metadata "github.com/kumahq/kuma/pkg/xds/envoy/metadata/v3"
	"github.com/kumahq/kuma/pkg/xds/generator/egress"
)

const defaultOverprovisioningFactor uint32 = 200

func NewEndpoints(
	existingEndpoints []*envoy_endpoint.LocalityLbEndpoints,
	tags mesh_proto.MultiValueTagSet,
	conf *api.Conf,
	localZone string,
	egressEnabled bool,
	origin string,
) []*envoy_endpoint.LocalityLbEndpoints {
	localPriorityGroups, crossZonePriorityGroups := GetLocalityGroups(conf, tags, localZone)
	var endpointsList []core_xds.Endpoint
	for _, localityLbEndpoint := range existingEndpoints {
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
	return endpoints.ToLocalityLbEndpoints(endpointsList)
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
