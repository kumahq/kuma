package api_server

import (
	"github.com/emicklei/go-restful"

	globalcp "github.com/Kong/kuma/pkg/globalcp/server"
)

func globalcpWs(globalcp globalcp.GlobalCP) *restful.WebService {
	ws := new(restful.WebService).Path("/globalcp")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		globalcp.StatusHandler(response)
	}))
}
