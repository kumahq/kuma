package api_server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
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

type inspectClient struct {
	adminClient admin.EnvoyAdminClient
	access      access.EnvoyAdminAccess
	rm          manager.ResourceManager
}

func addInspectEnvoyAdminEndpoints(
	ws *restful.WebService,
	cfg *kuma_cp.Config,
	rm manager.ResourceManager,
	adminAccess access.EnvoyAdminAccess,
	envoyAdminClient admin.EnvoyAdminClient,
) {
	cl := inspectClient{
		adminClient: envoyAdminClient,
		access:      adminAccess,
		rm:          rm,
	}
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{dataplane}/{type}").
			To(cl.inspectDataplaneAdmin()).
			Doc("inspect dataplane configuration and stats").
			Param(ws.PathParameter("mesh", "mesh name").DataType("string")).
			Param(ws.PathParameter("dataplane", "dataplane name").DataType("string")).
			Param(ws.PathParameter("type", "type of configuration to inspect").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneingresses/{zoneingress}/{type}").
			To(cl.inspectZoneIngressAdmin(cfg.Mode, cfg.Multizone.Zone.Name)).
			Doc("inspect zone ingresses XDS configuration").
			Produces("application/json").
			Param(ws.PathParameter("zoneingress", "zoneingress name").DataType("string")).
			Param(ws.PathParameter("type", "type of configuration to inspect").DataType("string")),
	)
	ws.Route(
		ws.GET("/zoneegresses/{zoneegress}/{type}").
			To(cl.inspectZoneEgressAdmin(cfg.Mode, cfg.Multizone.Zone.Name)).
			Doc("inspect zone egresses XDS configuration").
			Produces("application/json").
			Param(ws.PathParameter("zoneegress", "zoneegress name").DataType("string")).
			Param(ws.PathParameter("type", "type of configuration to inspect").DataType("string")),
	)
}

type adminType string

const (
	adminTypeConfigDump adminType = "xds"
	adminTypeClusters   adminType = "clusters"
	adminTypeStats      adminType = "stats"
)

func (cl *inspectClient) inspectDataplaneAdmin() restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		meshName := request.PathParameter("mesh")
		dataplaneName := request.PathParameter("dataplane")
		aType, err := cl.adminType(ctx, request.PathParameter("type"))
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not execute admin operation")
			return
		}

		dp := core_mesh.NewDataplaneResource()
		if err := cl.rm.Get(request.Request.Context(), dp, store.GetByKey(dataplaneName, meshName)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get dataplane resource")
			return
		}
		cl.inspectProxy(request, response, aType, dp)
	}
}

func (cl *inspectClient) inspectZoneIngressAdmin(mode core.CpMode, localZone string) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		zoneIngressName := request.PathParameter("zoneingress")

		aType, err := cl.adminType(ctx, request.PathParameter("type"))
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not execute admin operation")
			return
		}

		zi := core_mesh.NewZoneIngressResource()
		if err := cl.rm.Get(ctx, zi, store.GetByKey(zoneIngressName, core_model.NoMesh)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get zone ingress resource")
			return
		}

		if mode == core.Zone && zi.IsRemoteIngress(localZone) {
			rest_errors.HandleError(request.Request.Context(), response, &validators.ValidationError{}, "Could not connect to zone ingress that resides in another zone")
			return
		}
		cl.inspectProxy(request, response, aType, zi)
	}
}

func (cl *inspectClient) inspectZoneEgressAdmin(mode core.CpMode, localZone string) restful.RouteFunction {
	return func(request *restful.Request, response *restful.Response) {
		ctx := request.Request.Context()
		zoneEgressName := request.PathParameter("zoneegress")
		aType, err := cl.adminType(ctx, request.PathParameter("type"))
		if err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not execute admin operation")
			return
		}

		ze := core_mesh.NewZoneEgressResource()
		if err := cl.rm.Get(ctx, ze, store.GetByKey(zoneEgressName, core_model.NoMesh)); err != nil {
			rest_errors.HandleError(request.Request.Context(), response, err, "Could not get zone ingress resource")
			return
		}

		if mode == core.Zone && ze.IsRemoteEgress(localZone) {
			rest_errors.HandleError(request.Request.Context(), response, &validators.ValidationError{}, "Could not connect to zone ingress that resides in another zone")
			return
		}
		cl.inspectProxy(request, response, aType, ze)
	}
}

func (cl *inspectClient) adminType(ctx context.Context, adType string) (adminType, error) {
	res := adminType(adType)
	switch res {
	case adminTypeConfigDump:
		if err := cl.access.ValidateViewConfigDump(ctx, user.FromCtx(ctx)); err != nil {
			return "", err
		}
	case adminTypeClusters:
		if err := cl.access.ValidateViewClusters(ctx, user.FromCtx(ctx)); err != nil {
			return "", err
		}
	case adminTypeStats:
		if err := cl.access.ValidateViewStats(ctx, user.FromCtx(ctx)); err != nil {
			return "", err
		}
	default:
		return "", rest_errors.NewNotFoundError(fmt.Sprintf("invalid admin type: %s", adType))
	}
	return res, nil
}

func (cl *inspectClient) inspectProxy(request *restful.Request, response *restful.Response, adType adminType, proxy core_model.ResourceWithAddress) {
	ctx := request.Request.Context()
	contentType := contentTypeText

	var res []byte
	var err error
	format := v1alpha1.AdminOutputFormat_TEXT
	switch adType {
	case adminTypeConfigDump:
		includeEds, _ := strconv.ParseBool(request.QueryParameter("include_eds"))
		contentType = restful.MIME_JSON
		res, err = cl.adminClient.ConfigDump(ctx, proxy, includeEds)
	case adminTypeClusters:
		if request.QueryParameter("format") == "json" {
			contentType = restful.MIME_JSON
			format = v1alpha1.AdminOutputFormat_JSON
		}
		res, err = cl.adminClient.Clusters(ctx, proxy, format)
	case adminTypeStats:
		if request.QueryParameter("format") == "json" {
			contentType = restful.MIME_JSON
			format = v1alpha1.AdminOutputFormat_JSON
		}
		res, err = cl.adminClient.Stats(ctx, proxy, format)
	}
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not execute admin operation")
		return
	}

	response.AddHeader(restful.HEADER_ContentType, contentType)
	if _, err := response.Write(res); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Could not write response")
		return
	}
}
