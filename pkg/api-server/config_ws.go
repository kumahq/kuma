package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/config"
)

func configWs(cfg config.Config) (*restful.WebService, error) {
	cfgForDisplay, err := config.ConfigForDisplay(cfg)
	if err != nil {
		return nil, err
	}
	json, err := config.ToJson(cfgForDisplay)
	if err != nil {
		return nil, err
	}
	ws := new(restful.WebService).Path("/config")
	ws.Route(ws.GET("").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write(json); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return ws, nil
}
