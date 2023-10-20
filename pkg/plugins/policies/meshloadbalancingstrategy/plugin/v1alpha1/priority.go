package v1alpha1

import api "github.com/kumahq/kuma/pkg/plugins/policies/meshloadbalancingstrategy/api/v1alpha1"


type LocalityLbGroups struct {
	Key string
	Value string
	Weight uint32
}

type CrossZoneLocalityLbGroups struct {
	Disabled bool
	AllZones bool
	Zones map[string]bool
	ExceptZones map[string]bool
	Priority uint32
}

func GetPriority(conf *api.Conf){
	localZonePriority := []LocalityLbGroups{}

	// each endpoint iterate
	crossZonePriority := []CrossZoneLocalityLbGroups{}
	if conf.LocalityAwareness != nil && conf.LocalityAwareness.LocalZone != nil {
		tagsSet := dataplane.Spec.TagSet()
		for _, tag := range conf.LocalityAwareness.LocalZone.AffinityTags {
			values := tagsSet.Values(tag.Key)
			if len(tagsSet.Values(tag.Key)) != 0 {
				localZonePriority = append(localZonePriority, LocalityLbGroups{
					Key: tag.Key,
					Value: values[0],
					Weight: uint32(tag.Weight.IntVal),
					// Zone: zone,
				})
			}
		}
		if conf.LocalityAwareness.CrossZone != nil && len(conf.LocalityAwareness.CrossZone.Failover) > 0 {
			for priority, rule := range conf.LocalityAwareness.CrossZone.Failover{
				lb := CrossZoneLocalityLbGroups{}
				if rule.From != nil {
					doesRuleApply := false
					for _, zone := range rule.From.Zones{
						if zone == localZone {
							doesRuleApply = true
						}
					}
					if !doesRuleApply {
						continue
					}
				}
				switch rule.To.Type{
				case api.Any:
					lb.Disabled = false
					lb.AllZones = true
					lb.Priority = uint32(priority + 1)
				case api.AnyExcept:
					lb.Disabled = false
					lb.AllZones = false
					exceptZones := map[string]bool{}
					for _, zone := range rule.To.Zones {
						exceptZones[zone] = true
					}
					lb.ExceptZones = exceptZones
					lb.Priority = uint32(priority + 1)
				case api.Only:
					lb.Disabled = false
					lb.AllZones = false
					onlyZones := map[string]bool{}
					for _, zone := range rule.To.Zones {
						onlyZones[zone] = true
					}
					lb.Zones = onlyZones
					lb.Priority = uint32(priority + 1)
				default: 
					lb.Disabled = true
				}
				crossZonePriority = append(crossZonePriority, lb)
			}	
		}
		
	}
}