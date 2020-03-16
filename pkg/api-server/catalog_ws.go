package api_server

import (
	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/catalog"
	config_catalog "github.com/Kong/kuma/pkg/config/api-server/catalog"
)

func catalogWs(cfg config_catalog.CatalogConfig) *restful.WebService {
	ws := new(restful.WebService).Path("/catalog")
	return ws.Route(ws.GET("").To(func(request *restful.Request, response *restful.Response) {
		if err := response.WriteAsJson(catalog.FromConfig(cfg)); err != nil {
			log.Error(err, "Could not write the components response")
		}
	}))
}
