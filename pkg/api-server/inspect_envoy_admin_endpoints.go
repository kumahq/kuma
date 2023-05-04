package api_server

import (
	"context"

	"github.com/emicklei/go-restful/v3"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/config/core"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/envoy/admin/access"
)

const (
	contentTypeText = "text/plain"
)

func addInspectEnvoyAdminEndpoints(
	ws *restful.WebService,
	cfg *kuma_cp.Config,
	rm manager.ResourceManager,
	adminAccess access.EnvoyAdminAccess,
	envoyAdminClient admin.EnvoyAdminClient,
) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/xds").
			To(inspectDataplaneAdmin(envoyAdminClient.ConfigDump, adminAccess.ValidateViewConfigDump, rm, restful.MIME_JSON)).
			Doc("inspect dataplane XDS configuration").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneingresses/{zoneingress}/xds").
			To(inspectZoneIngressAdmin(cfg.Mode, cfg.Multizone.Zone.Name, envoyAdminClient.ConfigDump, adminAccess.ValidateViewConfigDump, rm, restful.MIME_JSON)).
			Doc("inspect zone ingresses XDS configuration").
			Produces("application/json").
			Param(ws.PathParameter("zoneingress", "zoneingress name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneegresses/{zoneegress}/xds").
			To(inspectZoneEgressAdmin(envoyAdminClient.ConfigDump, adminAccess.ValidateViewConfigDump, rm, restful.MIME_JSON)).
			Doc("inspect zone egresses XDS configuration").
			Produces("application/json").
			Param(ws.PathParameter("zoneegress", "zoneegress name").DataType("string")),
	)

	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/stats").
			To(inspectDataplaneAdmin(envoyAdminClient.Stats, adminAccess.ValidateViewStats, rm, contentTypeText)).
			Doc("inspect dataplane stats").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneingresses/{zoneingress}/stats").
			To(inspectZoneIngressAdmin(cfg.Mode, cfg.Multizone.Zone.Name, envoyAdminClient.Stats, adminAccess.ValidateViewStats, rm, contentTypeText)).
			Doc("inspect zone ingresses stats").
			Param(ws.PathParameter("zoneingress", "zoneingress name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneegresses/{zoneegress}/stats").
			To(inspectZoneEgressAdmin(envoyAdminClient.Stats, adminAccess.ValidateViewStats, rm, contentTypeText)).
			Doc("inspect zone egresses stats").
			Param(ws.PathParameter("zoneegress", "zoneegress name").DataType("string")),
	)

	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/clusters").
			To(inspectDataplaneAdmin(envoyAdminClient.Clusters, adminAccess.ValidateViewClusters, rm, contentTypeText)).
			Doc("inspect dataplane clusters").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneingresses/{zoneingress}/clusters").
			To(inspectZoneIngressAdmin(cfg.Mode, cfg.Multizone.Zone.Name, envoyAdminClient.Clusters, adminAccess.ValidateViewClusters, rm, contentTypeText)).
			Doc("inspect zone ingresses clusters").
			Param(ws.PathParameter("zoneingress", "zoneingress name").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneegresses/{zoneegress}/clusters").
			To(inspectZoneEgressAdmin(envoyAdminClient.Clusters, adminAccess.ValidateViewClusters, rm, contentTypeText)).
			Doc("inspect zone egresses clusters").
			Param(ws.PathParameter("zoneegress", "zoneegress name").DataType("string")),
	)
}

func inspectDataplaneAdmin(
	adminFn func(context.Context, core_model.ResourceWithAddress) ([]byte, error),
	access func(user.User) error,
	rm manager.ResourceManager,
	contentType string,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")

		if err := access(user.FromCtx(ctx)); err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		dp := core_mesh.NewDataplaneResource()
		if err := rm.Get(ctx, dp, store.GetByKey(dataplaneName, meshName)); err != nil {
			rest_errors.HandleError(response, err, "Could not get dataplane resource")
			return
		}

		stats, err := adminFn(ctx, dp)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		response.AddHeader(restful.HEADER_ContentType, contentType)
		if _, err := response.Write(stats); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func inspectZoneIngressAdmin(
	mode core.CpMode,
	localZone string,
	adminFn func(context.Context, core_model.ResourceWithAddress) ([]byte, error),
	access func(user.User) error,
	rm manager.ResourceManager,
	contentType string,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		zoneIngressName := request.PathParameter("zoneingress")

		if err := access(user.FromCtx(ctx)); err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		zi := core_mesh.NewZoneIngressResource()
		if err := rm.Get(ctx, zi, store.GetByKey(zoneIngressName, core_model.NoMesh)); err != nil {
			rest_errors.HandleError(response, err, "Could not get zone ingress resource")
			return
		}

		if mode == core.Zone && zi.IsRemoteIngress(localZone) {
			rest_errors.HandleError(response, &validators.ValidationError{},
				"Could not connect to zone ingress that resides in another zone")
			return
		}

		stats, err := adminFn(ctx, zi)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		response.AddHeader(restful.HEADER_ContentType, contentType)
		if _, err := response.Write(stats); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}

func inspectZoneEgressAdmin(
	adminFn func(context.Context, core_model.ResourceWithAddress) ([]byte, error),
	access func(user.User) error,
	rm manager.ResourceManager,
	contentType string,
) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		name := request.PathParameter("zoneegress")

		if err := access(user.FromCtx(ctx)); err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		ze := core_mesh.NewZoneEgressResource()
		if err := rm.Get(ctx, ze, store.GetByKey(name, core_model.NoMesh)); err != nil {
			rest_errors.HandleError(response, err, "Could not get zone egress resource")
			return
		}

		stats, err := adminFn(ctx, ze)
		if err != nil {
			rest_errors.HandleError(response, err, "Could not execute admin operation")
			return
		}

		response.AddHeader(restful.HEADER_ContentType, contentType)
		if _, err := response.Write(stats); err != nil {
			rest_errors.HandleError(response, err, "Could not write response")
			return
		}
	}
}
