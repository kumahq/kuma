package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/config"
)

func addConfigEndpoints(ws *restful.WebService, cfg config.Config) error {
	cfgForDisplay, err := config.ConfigForDisplay(cfg)
	if err != nil {
		return err
	}
	json, err := config.ToJson(cfgForDisplay)
	if err != nil {
		return err
	}
	ws.Route(ws.GET("/config").To(func(req *restful.Request, resp *restful.Response) {
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write(json); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return nil
}
