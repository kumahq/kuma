package api_server

import (
	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/clusters/poller"
)

func clustersWs(clusters poller.ClusterStatusPoller) *restful.WebService {
	ws := new(restful.WebService).Path("/status/clusters")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		if err := response.WriteAsJson(clusters.Clusters()); err != nil {
			log.Error(err, "failed marshaling response")
		}
	}))
}
