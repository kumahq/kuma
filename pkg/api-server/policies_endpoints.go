package api_server

import (
	"sort"

	"github.com/emicklei/go-restful/v3"

	api_types "github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func addPoliciesWsEndpoints(ws *restful.WebService, isGlobal bool, federatedZone bool, readOnly bool, defs []model.ResourceTypeDescriptor) {
	ws.Route(ws.GET("/policies").To(func(req *restful.Request, resp *restful.Response) {
		response := types.PoliciesResponse{}
		for _, def := range defs {
			if !def.IsPolicy {
				continue
			}
			response.Policies = append(response.Policies, types.PolicyEntry{
				Name:                string(def.Name),
				ReadOnly:            readOnly || federatedZone || def.ReadOnly,
				Path:                def.WsPath,
				SingularDisplayName: def.SingularDisplayName,
				PluralDisplayName:   def.PluralDisplayName,
				IsExperimental:      def.IsExperimental,
				IsTargetRefBased:    def.IsTargetRefBased,
				IsInbound:           def.HasFromTargetRef,
				IsOutbound:          def.HasToTargetRef,
			})
		}
		sort.SliceStable(response.Policies, func(i, j int) bool {
			return response.Policies[i].Name < response.Policies[j].Name
		})

		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the response")
		}
	}))
	ws.Route(ws.GET("/_resources").To(func(req *restful.Request, resp *restful.Response) {
		response := api_types.ResourceTypeDescriptionList{}
		for _, def := range defs {
			td := api_common.ResourceTypeDescription{
				Name:                string(def.Name),
				ReadOnly:            readOnly || federatedZone || def.ReadOnly,
				Path:                def.WsPath,
				SingularDisplayName: def.SingularDisplayName,
				PluralDisplayName:   def.PluralDisplayName,
				Scope:               api_common.ResourceTypeDescriptionScope(def.Scope),
				IncludeInDump:       isGlobal || def.DumpForGlobal,
			}
			if def.IsPolicy {
				td.Policy = &api_common.PolicyDescription{
					HasToTargetRef:   def.HasToTargetRef,
					HasFromTargetRef: def.HasFromTargetRef,
					IsTargetRef:      def.IsTargetRefBased,
				}
			}
			response.Resources = append(response.Resources, td)
		}
		sort.SliceStable(response.Resources, func(i, j int) bool {
			return response.Resources[i].Name < response.Resources[j].Name
		})
		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the response")
		}
	}))
}
