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

// Structure which hold information about cross zone with zones in the map to gice better access
type CrossZoneLbGroup struct {
	Type     api.ToZoneType
	Zones    map[string]bool
	Priority uint32
}

func GetPriority(conf *api.Conf, inboundTags mesh_proto.MultiValueTagSet, localZone string) ([]LocalLbGroup, []CrossZoneLbGroup) {
	if conf.LocalityAwareness == nil {
		return nil, nil
	}
	return getLocalLbGroups(conf, inboundTags), getCrossZoneLbGroup(conf, localZone)
}

func getLocalLbGroups(conf *api.Conf, inboundTags mesh_proto.MultiValueTagSet) []LocalLbGroup {
	localGroups := []LocalLbGroup{}
	if conf.LocalityAwareness.LocalZone != nil {
		rulesLen := len(conf.LocalityAwareness.LocalZone.AffinityTags)
		for i, tag := range conf.LocalityAwareness.LocalZone.AffinityTags {
			values := inboundTags.Values(tag.Key)
			weight := pointer.DerefOr(tag.Weight, uint32(9*math.Pow10(rulesLen-i-1)))
			if len(values) != 0 {
				localGroups = append(localGroups, LocalLbGroup{
					Key:    tag.Key,
					Value:  values[0],
					Weight: weight,
				})
			}
		}
	}
	return localGroups
}

func getCrossZoneLbGroup(conf *api.Conf, localZone string) []CrossZoneLbGroup {
	crossZoneGroups := []CrossZoneLbGroup{}
	if conf.LocalityAwareness.CrossZone != nil && len(conf.LocalityAwareness.CrossZone.Failover) > 0 {
		// iterator starts from 0 while for remote zones we always sets priority that favors local zone
		for priority, rule := range conf.LocalityAwareness.CrossZone.Failover {
			lb := CrossZoneLbGroup{}
			if rule.From != nil {
				doesRuleApply := false
				for _, zone := range rule.From.Zones {
					if zone == localZone {
						doesRuleApply = true
						break
					}
				}
				if !doesRuleApply {
					continue
				}
			}
			switch rule.To.Type {
			case api.Any:
				lb.Type = api.Any
				lb.Priority = uint32(priority + 1)
			case api.AnyExcept:
				lb.Type = api.AnyExcept
				zones := map[string]bool{}
				for _, zone := range rule.To.Zones {
					zones[zone] = true
				}
				lb.Zones = zones
				lb.Priority = uint32(priority + 1)
			case api.Only:
				lb.Type = api.Only
				zones := map[string]bool{}
				for _, zone := range rule.To.Zones {
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
