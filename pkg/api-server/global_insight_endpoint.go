package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/insights/globalinsight"
)

const GlobalInsightPath = "/global-insight"

type globalInsightEndpoint struct {
	globalInsightService globalinsight.GlobalInsightService
}

func (ge *globalInsightEndpoint) addEndpoint(ws *restful.WebService) {
	ws.Route(
		ws.GET(GlobalInsightPath).To(ge.getGlobalInsight).
			Doc("Get Global Insight").
			Returns(200, "OK", nil),
	)
}

func (ge *globalInsightEndpoint) getGlobalInsight(request *restful.Request, response *restful.Response) {
	ctx := request.Request.Context()
	globalInsight, err := ge.globalInsightService.GetGlobalInsight(ctx)
	if err != nil {
		errors.HandleError(ctx, response, err, "Could not retrieve GlobalInsight")
		return
	}

	if err = response.WriteAsJson(globalInsight); err != nil {
		errors.HandleError(ctx, response, err, "Could not write response")
		return
	}
}
