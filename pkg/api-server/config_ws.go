package api_server

import (
	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/v3/pkg/config"
	"github.com/kumahq/kuma/v3/pkg/core/access"
	"github.com/kumahq/kuma/v3/pkg/core/user"
)

func addConfigEndpoints(ws *restful.WebService, access access.ControlPlaneMetadataAccess, cfg config.Config) error {
	cfgForDisplay, err := config.ConfigForDisplay(cfg)
	if err != nil {
		return err
	}
	ws.Route(ws.GET("/config").To(handle(func(req *restful.Request) (any, error) {
		ctx := req.Request.Context()
		if err := access.ValidateView(ctx, user.FromCtx(ctx)); err != nil {
			return nil, withTitle(err, "Access denied")
		}
		return rawResponse{contentType: "application/json", body: []byte(cfgForDisplay)}, nil
	})))
	return nil
}
