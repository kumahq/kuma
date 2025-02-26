package v1alpha1

import (
	"math"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/util/pointer"
)

type LocalLbGroup struct {
	Key    string
	Value  string
	Weight uint32
}

// Structure which hold information about cross zone with zones in the map make it more accessible
type CrossZoneLbGroup struct {
	Type     api.ToZoneType
	Zones    map[string]bool
	Priority uint32
}

func GetLocalityGroups(conf *api.Conf, inboundTags mesh_proto.MultiValueTagSet, localZone string) ([]LocalLbGroup, []CrossZoneLbGroup) {
	if conf.LocalityAwareness == nil {
		return nil, nil
	}
	return getLocalLbGroups(conf, inboundTags), getCrossZoneLbGroups(conf, localZone)
}

func getLocalLbGroups(conf *api.Conf, inboundTags mesh_proto.MultiValueTagSet) []LocalLbGroup {
	var localGroups []LocalLbGroup
	if conf.LocalityAwareness.LocalZone != nil {
		rulesLen := len(pointer.Deref(conf.LocalityAwareness.LocalZone.AffinityTags))
		for i, tag := range pointer.Deref(conf.LocalityAwareness.LocalZone.AffinityTags) {
			values := inboundTags.Values(tag.Key)
			// when weights are not provided we are generating weights by ourselves
			// the first rule has the highest priority which is 9 * 10^(number of rules - rules position -1)
			// that makes that the highest priority locality gets 90% of the traffic and sum of all weights is a
			// power of 10
			weight := pointer.DerefOr(tag.Weight, uint32(9*math.Pow10(rulesLen-i-1)))
			if len(values) != 0 {
				localGroups = append(localGroups, LocalLbGroup{
					Key:    tag.Key,
					Value:  values[0], // we are taking the first value from multiple, because locality tag shouldn't have different values
					Weight: weight,
				})
			}
		}
	}
	return localGroups
}

func getCrossZoneLbGroups(conf *api.Conf, localZone string) []CrossZoneLbGroup {
	var crossZoneGroups []CrossZoneLbGroup
	if conf.LocalityAwareness.CrossZone != nil && len(pointer.Deref(conf.LocalityAwareness.CrossZone.Failover)) > 0 {
		// iterator starts from 0 while for remote zones we always set priority that favors local zone
		// we are using priority based on the rule position in the list so we increment it even if it doesn't match
		// that doesn't affect envoy behavior
		for priority, rule := range pointer.Deref(conf.LocalityAwareness.CrossZone.Failover) {
			if !doesRuleApply(rule.From, localZone) {
				continue
			}
			lb := CrossZoneLbGroup{}
			switch rule.To.Type {
			case api.Any:
				lb.Type = api.Any
				lb.Priority = uint32(priority + 1)
			case api.AnyExcept:
				lb.Type = api.AnyExcept
				zones := map[string]bool{}
				for _, zone := range pointer.Deref(rule.To.Zones) {
					zones[zone] = true
				}
				lb.Zones = zones
				lb.Priority = uint32(priority + 1)
			case api.Only:
				lb.Type = api.Only
				zones := map[string]bool{}
				for _, zone := range pointer.Deref(rule.To.Zones) {
					zones[zone] = true
				}
				lb.Zones = zones
				lb.Priority = uint32(priority + 1)
			default:
				lb.Type = api.None
			}
			crossZoneGroups = append(crossZoneGroups, lb)
		}
	}
	return crossZoneGroups
}

func doesRuleApply(rule *api.FromZone, localZone string) bool {
	if rule == nil {
		return true
	}
	for _, zone := range rule.Zones {
		if zone == localZone {
			return true
		}
	}
	return false
}
