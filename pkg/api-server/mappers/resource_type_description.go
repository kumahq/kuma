package mappers

import (
	"sort"

	api_types "github.com/kumahq/kuma/v2/api/openapi/types"
	api_common "github.com/kumahq/kuma/v2/api/openapi/types/common"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

func MapResourceTypeDescription(defs []model.ResourceTypeDescriptor, readOnly bool, federatedZone bool) api_types.ResourceTypeDescriptionList {
	sorted := make([]model.ResourceTypeDescriptor, len(defs))
	copy(sorted, defs)
	sort.SliceStable(sorted, func(i, j int) bool {
		di, dj := sorted[i], sorted[j]
		iOrdered, jOrdered := di.Order > 0, dj.Order > 0
		switch {
		case iOrdered && jOrdered:
			return di.Order < dj.Order
		case iOrdered:
			return false
		case jOrdered:
			return true
		default:
			return di.Name < dj.Name
		}
	})

	response := api_types.ResourceTypeDescriptionList{}
	for _, def := range sorted {
		td := api_common.ResourceTypeDescription{
			Name:                string(def.Name),
			ReadOnly:            readOnly || federatedZone || def.ReadOnly,
			Path:                def.WsPath,
			SingularDisplayName: def.SingularDisplayName,
			PluralDisplayName:   def.PluralDisplayName,
			Scope:               api_common.ResourceTypeDescriptionScope(def.Scope),
			ShortName:           def.ShortName,
			// Things in the federation export should be:
			//	1. not system managed .i.e: not ReadOnly (ServiceInsight for example is like this)
			//	2. have KDS from global to zone
			IncludeInFederation: def.KDSFlags.Has(model.GlobalToZonesFlag) && !def.ReadOnly,
		}
		if def.IsPolicy {
			td.Policy = &api_common.PolicyDescription{
				HasToTargetRef:    def.HasToTargetRef,
				HasFromTargetRef:  def.HasFromTargetRef,
				HasRulesTargetRef: def.HasRulesTargetRef,
				IsTargetRef:       def.IsTargetRefBased,
				IsFromAsRules:     def.IsFromAsRules,
			}
		}
		response.Resources = append(response.Resources, td)
	}
	return response
}
