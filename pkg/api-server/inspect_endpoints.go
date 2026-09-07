package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice"
	meshservice_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/meshservice/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
)

func addInspectEndpoints(
	ws *restful.WebService,
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/meshservices/{name}/_dataplanes").To(handle(inspectMeshServiceDataplanes(rm, resourceAccess))).
			Doc("inspect MeshService").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("name", "resource name").DataType("string")).
			Returns(200, "OK", nil),
	)
}

// inspectMeshServiceDataplanes provides the standardized /_dataplanes endpoint.
// Uses exact tag matching via meshservice.MatchesDataplane() to fix multizone aggregation issues.
func inspectMeshServiceDataplanes(
	rm manager.ResourceManager,
	resourceAccess access.ResourceAccess,
) handlerFunc {
	return func(request *restful.Request) (any, error) {
		return matchingDataplanesForFilter(
			request,
			meshservice_api.MeshServiceResourceTypeDescriptor,
			rm,
			resourceAccess,
			func(resource core_model.Resource) store.ListFilterFunc {
				meshService := resource.(*meshservice_api.MeshServiceResource)
				return func(rs core_model.Resource) bool {
					return meshservice.MatchesDataplane(meshService.Spec, rs.(*core_mesh.DataplaneResource))
				}
			},
		)
	}
}
