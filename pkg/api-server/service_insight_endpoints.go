package api_server

import (
	"fmt"

	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/insights"
)

type serviceInsightEndpoints struct {
	resourceEndpoints
}

func (s *serviceInsightEndpoints) addFindEndpoint(ws *restful.WebService, pathPrefix string) {
	ws.Route(ws.GET(pathPrefix+"/{service}").To(s.findResource).
		Filter(s.auth()).
		Doc(fmt.Sprintf("Get a %s", s.Name)).
		Param(ws.PathParameter("service", fmt.Sprintf("Name of a %s", s.Name)).DataType("string")).
		Returns(200, "OK", nil).
		Returns(404, "Not found", nil))
}

func (s *serviceInsightEndpoints) findResource(request *restful.Request, response *restful.Response) {
	service := request.PathParameter("service")
	meshName := s.meshFromRequest(request)

	serviceInsight := &mesh.ServiceInsightResource{}
	err := s.resManager.Get(request.Request.Context(), serviceInsight, store.GetByKey(insights.ServiceInsightName(meshName), meshName))
	if err != nil {
		rest_errors.HandleError(response, err, "Could not retrieve a resource")
	} else {
		stat := serviceInsight.Spec.Services[service]
		if stat == nil {
			stat = &v1alpha1.ServiceInsight_DataplaneStat{}
		}
		if err := response.WriteAsJson(stat); err != nil {
			core.Log.Error(err, "Could not write the response")
		}
	}
}
