package api_server

import (
	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/api-server/types"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

func indexWs() *restful.WebService {
	ws := new(restful.WebService)
	return ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		response := types.IndexResponse{
			Tagline: types.TaglineKuma,
			Version: kuma_version.Build.Version,
		}
		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
}
