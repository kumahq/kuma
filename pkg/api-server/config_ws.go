package api_server

import (
	"github.com/Kong/kuma/pkg/config"
	"github.com/emicklei/go-restful"
)

func configWs(cfg config.Config) (*restful.WebService, error) {
	cfgForDisplay, err := config.ConfigForDisplayJSON(cfg)
	if err != nil {
		return nil, err
	}
	ws := new(restful.WebService).Path("/config")
	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write(cfgForDisplay); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return ws, nil
}
