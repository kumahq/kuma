package api_server

import (
	"sort"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/api-server/types"
	"github.com/kumahq/kuma/pkg/config/core"
	"github.com/kumahq/kuma/pkg/core/resources/model"
)

func addPoliciesWsEndpoints(ws *restful.WebService, mode core.CpMode, readOnly bool, defs []model.ResourceTypeDescriptor) error {
	ws.Route(ws.GET("/policies").To(func(req *restful.Request, resp *restful.Response) {
		response := types.PoliciesResponse{}
		for _, def := range defs {
			if !def.IsPolicy {
				continue
			}
			response.Policies = append(response.Policies, types.PolicyEntry{
				Name:        string(def.Name),
				ReadOnly:    readOnly || core.Zone == mode,
				Path:        def.WsPath,
				DisplayName: def.DisplayName,
			})
		}
		sort.SliceStable(response.Policies, func(i, j int) bool {
			return response.Policies[i].Name < response.Policies[j].Name
		})

		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return nil
}
