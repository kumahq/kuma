package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/version"
)

func versionsWs() *restful.WebService {
	ws := new(restful.WebService).Path("/versions")

	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if err := resp.WriteAsJson(version.CompatibilityMatrix); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))

	return ws
}
