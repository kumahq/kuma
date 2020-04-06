package api_server

import (
	"os"

	"github.com/emicklei/go-restful"

	"github.com/Kong/kuma/pkg/api-server/types"
	kuma_version "github.com/Kong/kuma/pkg/version"
)

func addIndexWsEndpoints(ws *restful.WebService) error {
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}
	ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		response := types.IndexResponse{
			Hostname: hostname,
			Tagline:  types.TaglineKuma,
			Version:  kuma_version.Build.Version,
		}
		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return nil
}
