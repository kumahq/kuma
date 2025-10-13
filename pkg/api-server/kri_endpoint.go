package api_server

import (
	"fmt"

	"github.com/emicklei/go-restful/v3"
	"github.com/kumahq/kuma/pkg/core/kri"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/plugins/resources/k8s"
)

type kriEndpoint struct {
	k8sMapper              k8s.ResourceMapperFunc
}

func (r *kriEndpoint) addFindByKriEndpoint(ws *restful.WebService) {
	ws.Route(ws.GET("/_kri/{kri}").To(r.findByKriRoute()).Doc(fmt.Sprintf("Returns a resource by KRI")).
		Param(ws.PathParameter("kri", fmt.Sprintf("KRI of the resource")).DataType("string")).
		Returns(200, "OK", nil).
		Returns(400, "Bad request", nil).
		Returns(404, "Not found", nil))
}

func  (r *kriEndpoint) findByKriRoute() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		kriParam := request.PathParameter("kri")
		identifier, err := kri.FromString(kriParam)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not parse KRI")
		}

		resource, err := findByKri(identifier)
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not retrieve a resource")
		}

		res, err := formatResource(resource, request.QueryParameter("format"), r.k8sMapper , request.QueryParameter("namespace"))
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not format a resource")
		}

		if err := response.WriteAsJson(res); err != nil {
			log.Error(err, "Could not write the find response")
		}
	}
}

func findByKri(identifier kri.Identifier) (core_model.Resource, error) {
	return nil, nil
}

