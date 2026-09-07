package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/core/user"
)

type WhoamiResponse struct {
	Name   string   `json:"name"`
	Groups []string `json:"groups"`
}

func addWhoamiEndpoints(ws *restful.WebService) {
	ws.Route(ws.GET("/who-am-i").To(handle(func(req *restful.Request) (any, error) {
		u := user.FromCtx(req.Request.Context())
		return WhoamiResponse{
			Name:   u.Name,
			Groups: u.Groups,
		}, nil
	})))
}
