package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/pkg/config"
	"github.com/kumahq/kuma/pkg/core/access"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
)

func addConfigEndpoints(ws *restful.WebService, access access.ControlPlaneMetadataAccess, cfg config.Config) error {
	cfgForDisplay, err := config.ConfigForDisplay(cfg)
	if err != nil {
		return err
	}
	ws.Route(ws.GET("/config").To(func(req *restful.Request, resp *restful.Response) {
		ctx := req.Request.Context()
		if err := access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
			rest_errors.HandleError(ctx, resp, err, "Access denied")
			return
		}
		resp.AddHeader("content-type", "application/json")
		if _, err := resp.Write([]byte(cfgForDisplay)); err != nil {
			log.Error(err, "Could not write the index response")
		}
	}))
	return nil
}
