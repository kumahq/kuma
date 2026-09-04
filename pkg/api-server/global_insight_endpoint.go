package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/insights/globalinsight"
)

const GlobalInsightPath = "/global-insight"

type globalInsightEndpoint struct {
	globalInsightService globalinsight.GlobalInsightService
}

func (ge *globalInsightEndpoint) addEndpoint(ws *restful.WebService) {
	ws.Route(
		ws.GET(GlobalInsightPath).To(handle(ge.getGlobalInsight)).
			Doc("Get Global Insight").
			Returns(200, "OK", nil),
	)
}

func (ge *globalInsightEndpoint) getGlobalInsight(request *restful.Request) (any, error) {
	globalInsight, err := ge.globalInsightService.GetGlobalInsight(request.Request.Context())
	if err != nil {
		return nil, withTitle(err, "Could not retrieve GlobalInsight")
	}
	return globalInsight, nil
}
