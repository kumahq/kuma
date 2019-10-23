package api_server

import (
	kuma_cp "github.com/Kong/kuma/pkg/config/app/kuma-cp"
	"github.com/Kong/kuma/pkg/coordinates"
	"github.com/emicklei/go-restful"
)

func componentsWs(cfg kuma_cp.Config) *restful.WebService {
	ws := new(restful.WebService).Path("/coordinates")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		if err := response.WriteAsJson(coordinates.FromConfig(cfg)); err != nil {
			log.Error(err, "Could not write the components response")
		}
	}))
}
