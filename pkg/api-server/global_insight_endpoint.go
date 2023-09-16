package api_server

import (
	"encoding/json"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/insights"
)

const globalInsightPath = "/global-insight"

type globalInsightEndpoint struct {
	globalInsightService insights.GlobalInsightService
}

func (ge *globalInsightEndpoint) addEndpoint(ws *restful.WebService) {
	ws.Route(
		ws.GET(globalInsightPath).To(ge.getGlobalInsight).
			Doc("Get Global Insights").
			Returns(200, "OK", nil),
	)
}

func (ge *globalInsightEndpoint) getGlobalInsight(request *restful.Request, response *restful.Response) {
	ctx := request.Request.Context()
	globalInsight, err := ge.globalInsightService.GetGlobalInsight(ctx)
	if err != nil {
		errors.HandleError(ctx, response, err, "Could not retrieve Global Insight")
		return
	}

	marshal, err := json.Marshal(globalInsight)
	if err != nil {
		errors.HandleError(ctx, response, err, "Error parsing response")
		return
	}
	_, err = response.ResponseWriter.Write(marshal)
	if err != nil {
		errors.HandleError(ctx, response, err, "Could not retrieve Global Insight")
		return
	}
}
