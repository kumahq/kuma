package api_server

import (
	"github.com/Kong/kuma/pkg/catalogue"
	config_catalogue "github.com/Kong/kuma/pkg/config/api-server/catalogue"
	"github.com/emicklei/go-restful"
)

func catalogueWs(cfg config_catalogue.CatalogueConfig) *restful.WebService {
	ws := new(restful.WebService).Path("/catalogue")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		if err := response.WriteAsJson(catalogue.FromConfig(cfg)); err != nil {
			log.Error(err, "Could not write the components response")
		}
	}))
}
