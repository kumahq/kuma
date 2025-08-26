package api_server

import (
	"sort"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/mappers"
	"github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func addPoliciesWsEndpoints(ws *restful.WebService, isGlobal, isFederatedZone, readOnly bool, defs []model.ResourceTypeDescriptor) {
	ws.Route(ws.GET("/policies").To(func(req *restful.Request, resp *restful.Response) {
		response := types.PoliciesResponse{}
		for _, def := range defs {
			if !def.IsPolicy {
				continue
			}
			response.Policies = append(response.Policies, types.PolicyEntry{
				Name:                string(def.Name),
				ReadOnly:            readOnly || def.IsReadOnly(isGlobal, isFederatedZone),
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
		response := mappers.MapResourceTypeDescription(defs, readOnly, isFederatedZone)
		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the response")
		}
	}))
}
