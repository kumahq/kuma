package api_server

import (
	"github.com/emicklei/go-restful"

	kuma_version "github.com/Kong/kuma/pkg/version"
)

const TaglineKuma = "Kuma"

type IndexResponse struct {
	Tagline string `json:"tagline"`
	Version string `json:"version"`
}

func indexWs() *restful.WebService {
	ws := new(restful.WebService)
	return ws.Route(ws.GET("/").To(func(req *restful.Request, resp *restful.Response) {
		response := IndexResponse{
			Tagline: TaglineKuma,
			Version: kuma_version.Build.Version,
		}
		if err := resp.WriteAsJson(response); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
}
