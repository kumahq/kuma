package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/core/user"
)

type WhoamiResponse struct {
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

func addWhoamiEndpoints(ws *restful.WebService) {
	ws.Route(ws.GET("/who-am-i").To(func(req *restful.Request, resp *restful.Response) {
		u := user.FromCtx(req.Request.Context())
		whoamiResp := WhoamiResponse{
			Name:   u.Name,
			Groups: u.Groups,
		}

		if err := resp.WriteAsJson(whoamiResp); err != nil {
			log.Error(err, "Could not write the who-am-i response")
		}
	}))
}
