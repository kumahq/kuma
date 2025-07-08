package api_server

import (
	"net/http"
	"slices"
	"strings"

	"github.com/emicklei/go-restful/v3"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/api/openapi/types"
	api_common "github.com/kumahq/kuma/api/openapi/types/common"
	"github.com/kumahq/kuma/pkg/core/kri"
	"github.com/kumahq/kuma/pkg/core/resources/access"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	rest_errors "github.com/kumahq/kuma/pkg/core/rest/errors"
	"github.com/kumahq/kuma/pkg/core/user"
	util_slices "github.com/kumahq/kuma/pkg/util/slices"
	xds_context "github.com/kumahq/kuma/pkg/xds/context"
)

type dataplaneLayoutEndpoint struct {
	resManager         manager.ReadOnlyResourceManager
	meshContextBuilder xds_context.MeshContextBuilder
	resourceAccess     access.ResourceAccess
}

func newDataplaneLayoutEndpoint(resManager manager.ReadOnlyResourceManager, meshContextBuilder xds_context.MeshContextBuilder, resourceAccess access.ResourceAccess) *dataplaneLayoutEndpoint {
	return &dataplaneLayoutEndpoint{
		resManager:         resManager,
		meshContextBuilder: meshContextBuilder,
		resourceAccess:     resourceAccess,
	}
}

func (dle *dataplaneLayoutEndpoint) addEndpoint(ws *restful.WebService) {
	ws.Route(
		ws.GET("/meshes/{mesh}/dataplanes/{name}/_layout").
			To(dle.getLayout).
			Doc("Get Dataplane Layout").
			Returns(http.StatusOK, "OK", nil),
	)
}

func (dle *dataplaneLayoutEndpoint) getLayout(request *restful.Request, response *restful.Response) {
	meshName := request.PathParameter("mesh")
	dataplaneName := request.PathParameter("name")

	if err := dle.resourceAccess.ValidateGet(
		request.Request.Context(),
		core_model.ResourceKey{Mesh: meshName, Name: dataplaneName},
		core_mesh.DataplaneResourceTypeDescriptor,
		user.FromCtx(request.Request.Context()),
	); err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Access Denied")
		return
	}

	baseMeshContext, err := dle.meshContextBuilder.BuildBaseMeshContextIfChanged(request.Request.Context(), meshName, nil)
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to build MeshContext")
	}

	if baseMeshContext.Mesh.Spec.GetMeshServices().GetMode() == v1alpha1.Mesh_MeshServices_Disabled {
		rest_errors.HandleError(request.Request.Context(), response, rest_errors.NewBadRequestError("can't use _layout endpoint without meshService enabled"), "Bad Request")
		return
	}

	dataplane := core_mesh.NewDataplaneResource()
	err = dle.resManager.Get(request.Request.Context(), dataplane, store.GetByKey(dataplaneName, meshName))
	if err != nil {
		rest_errors.HandleError(request.Request.Context(), response, err, "Failed to retrieve Dataplane")
		return
	}

	inbounds := util_slices.Map(dataplane.Spec.GetNetworking().GetInbound(), func(inbound *v1alpha1.Dataplane_Networking_Inbound) api_common.DataplaneInbound {
		return api_common.DataplaneInbound{
			Kri:      kri.From(dataplane, inbound.GetSectionName()).String(),
			Port:     int32(inbound.GetPort()),
			Protocol: inbound.GetProtocol(),
		}
	})

	var outbounds []api_common.DataplaneOutbound
	reachableOutbounds, _ := baseMeshContext.GetReachableBackends(dataplane)
	for outboundKri, port := range reachableOutbounds {
		outbounds = append(outbounds, api_common.DataplaneOutbound{
			Kri:      outboundKri.String(),
			Port:     port.GetValue(),
			Protocol: string(port.GetProtocol()),
		})
	}

	slices.SortStableFunc(outbounds, func(a, b api_common.DataplaneOutbound) int {
		return strings.Compare(a.Kri, b.Kri)
	})

	networkingLayout := types.DataplaneNetworkingLayout{
		Inbounds:  inbounds,
		Kri:       kri.From(dataplane, "").String(),
		Labels:    dataplane.GetMeta().GetLabels(),
		Outbounds: outbounds,
	}

	err = response.WriteHeaderAndJson(http.StatusOK, networkingLayout, "application/json")
	if err != nil {
		log.Error(err, "Could not write response")
	}
}
