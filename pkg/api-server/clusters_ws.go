package api_server

import (
	"github.com/emicklei/go-restful"

	"github.com/kumahq/kuma/pkg/zones/poller"
)

func clustersWs(zones poller.ZoneStatusPoller) *restful.WebService {
	ws := new(restful.WebService).Path("/status/zones")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		if err := response.WriteAsJson(zones.Zones()); err != nil {
			log.Error(err, "failed marshaling response")
		}
	}))
}
