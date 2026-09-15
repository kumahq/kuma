package mappers

import (
	"sort"

	api_types "github.com/kumahq/kuma/v3/api/openapi/types"
	api_common "github.com/kumahq/kuma/v3/api/openapi/types/common"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

func MapResourceTypeDescription(defs []model.ResourceTypeDescriptor, readOnly bool, isGlobal bool, federatedZone bool) api_types.ResourceTypeDescriptionList {
	response := api_types.ResourceTypeDescriptionList{}
	for _, def := range defs {
		td := api_common.ResourceTypeDescription{
			Name:                string(def.Name),
			ReadOnly:            readOnly || def.IsReadOnly(isGlobal, federatedZone),
			Path:                def.WsPath,
			SingularDisplayName: def.SingularDisplayName,
			PluralDisplayName:   def.PluralDisplayName,
			Scope:               api_common.ResourceTypeDescriptionScope(def.Scope),
			ShortName:           def.ShortName,
			IsInsight:           def.IsInsight(),
			AdminOnly:           def.AdminOnly,
			// Things in the federation export should be:
			//	1. not system managed .i.e: not ReadOnly
			//	2. have KDS from global to zone
			IncludeInFederation: def.KDSFlags.Has(model.GlobalToZonesFlag) && !def.ReadOnly,
		}
		if def.IsPolicy {
			td.Policy = &api_common.PolicyDescription{
				HasToTargetRef:    def.HasToTargetRef,
				HasFromTargetRef:  false,
				HasRulesTargetRef: def.HasRulesTargetRef,
				IsTargetRef:       true,
				IsFromAsRules:     false,
			}
		}
		response.Resources = append(response.Resources, td)
	}
	sort.SliceStable(response.Resources, func(i, j int) bool {
		return response.Resources[i].Name < response.Resources[j].Name
	})
	return response
}
