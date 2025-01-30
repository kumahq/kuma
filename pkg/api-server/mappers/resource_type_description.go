package mappers

import (
	"sort"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func MapResourceTypeDescription(defs []model.ResourceTypeDescriptor, readOnly bool, federatedZone bool) api_types.ResourceTypeDescriptionList {
	response := api_types.ResourceTypeDescriptionList{}
	for _, def := range defs {
		td := api_common.ResourceTypeDescription{
			Name:                string(def.Name),
			ReadOnly:            readOnly || federatedZone || def.ReadOnly,
			Path:                def.WsPath,
			SingularDisplayName: def.SingularDisplayName,
			PluralDisplayName:   def.PluralDisplayName,
			Scope:               api_common.ResourceTypeDescriptionScope(def.Scope),
			IncludeInFederation: includeInFederation(def.KDSFlags, def.Name),
		}
		if def.IsPolicy {
			td.Policy = &api_common.PolicyDescription{
				HasToTargetRef:              def.HasToTargetRef,
				HasFromTargetRef:            def.HasFromTargetRef,
				IsTargetRef:                 def.IsTargetRefBased,
				InterpretFromEntriesAsRules: def.InterpretFromEntriesAsRules,
			}
		}
		response.Resources = append(response.Resources, td)
	}
	sort.SliceStable(response.Resources, func(i, j int) bool {
		return response.Resources[i].Name < response.Resources[j].Name
	})
	return response
}

func includeInFederation(kdsFlags model.KDSFlagType, name model.ResourceType) bool {
	if name == mesh.ServiceInsightType {
		return false
	}
	return kdsFlags&model.GlobalToAllZonesFlag != 0
}
