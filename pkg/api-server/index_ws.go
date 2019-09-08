package api_server

import (
	"github.com/emicklei/go-restful"

	kuma_version "github.com/Kong/kuma/pkg/version"
)

func indexWs() *restful.WebService {
	ws := new(restful.WebService)
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		info := map[string]string{
			"tagline": "Kuma",
			"version": kuma_version.Build.Version,
		}
		if err := resp.WriteAsJson(info); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return ws
}
