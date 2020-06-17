package api_server

import (
	"encoding/json"
	"net/http"

	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/clusters/poller"
)

func clustersWs(clusters poller.ClusterStatusPoller) *restful.WebService {
	ws := new(restful.WebService).Path("/status/clusters")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		response.WriteHeader(http.StatusOK)
		if err := clusters.EncodeClusters(json.NewEncoder(response)); err != nil {
			log.Error(err, "failed marshaling response")
		}
	}))
}
