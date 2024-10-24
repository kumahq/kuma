package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
)

func addInspectMeshServiceEndpoints(
	ws *restful.WebService,
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/meshservices/{name}/_resources/dataplanes").
			To(matchingDataplanesForMeshServices(rm, resourceAccess)).
			Doc("inspect dataplane configuration and stats").
			Param(ws.PathParameter("name", "mesh service name").DataType("string")).
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")),
	)
}

func matchingDataplanesForMeshServices(resManager manager.ResourceManager, resourceAccess access.ResourceAccess) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		matchingDataplanesForFilter(
			request,
			response,
			meshservice_api.MeshServiceResourceTypeDescriptor,
			resManager,
			resourceAccess,
			func(resource model.Resource) store.ListFilterFunc {
				meshService := resource.(*meshservice_api.MeshServiceResource)
				return func(rs model.Resource) bool {
					return meshservice.MatchesDataplane(meshService.Spec, rs.(*core_mesh.DataplaneResource))
				}
			},
		)
	}
}
