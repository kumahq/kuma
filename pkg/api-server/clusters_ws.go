package api_server

import (
	"github.com/emicklei/go-restful"

	clusters "github.com/Kong/kuma/pkg/clusters/server"
)

func clustersWs(clusters clusters.ClusterStatusServer) *restful.WebService {
	ws := new(restful.WebService).Path("/status/clusters")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		clusters.StatusHandler(response)
	}))
}
